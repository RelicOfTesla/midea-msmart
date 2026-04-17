// Package msmart provides minimal Midea cloud access.
package msmart

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultCloudRegion is the default cloud region
const DefaultCloudRegion = "US"

// CloudError is a generic exception for Midea cloud errors.
var CloudError = errors.New("cloud error")

// ApiError is an exception class for Midea cloud API errors.
type ApiError struct {
	Message string
	Code    interface{}
}

// Error implements the error interface
func (e *ApiError) Error() string {
	return fmt.Sprintf("Code: %v, Message: %s", e.Code, e.Message)
}

// NewApiError creates a new ApiError
func NewApiError(message string, code interface{}) *ApiError {
	return &ApiError{
		Message: message,
		Code:    code,
	}
}

// BaseCloud is the base class for minimal Midea cloud access.
type BaseCloud struct {
	// Misc constants for the API
	APP_ID      string
	CLIENT_TYPE int // Android
	FORMAT      int // JSON
	LANGUAGE    string
	DEVICE_ID   string // Random device ID

	// Default number of request retries
	RETRIES int

	CLOUD_CREDENTIALS map[string][2]string

	// Internal fields
	account        string
	password       string
	apiLock        sync.Mutex
	baseURL        string
	loginID        *string
	session        map[string]interface{}
	getAsyncClient func() *http.Client

	// parseResponseFunc is the function to parse HTTP responses
	// This allows subclasses to override the parsing behavior
	parseResponseFunc func(*http.Response) (interface{}, error)

	// apiRequestFunc is the function to make API requests
	// This allows subclasses to override the request behavior
	apiRequestFunc func(endpoint string, body map[string]interface{}) (map[string]interface{}, error)
}

// NewBaseCloud creates a new BaseCloud instance
func NewBaseCloud(baseURL string, region *string, account *string, password *string, getAsyncClient func() *http.Client) (*BaseCloud, error) {
	bc := &BaseCloud{
		APP_ID:            "",
		CLIENT_TYPE:       1, // Android
		FORMAT:            2, // JSON
		LANGUAGE:          "en_US",
		DEVICE_ID:         generateTokenHex(8), // Random device ID
		RETRIES:           3,
		CLOUD_CREDENTIALS: make(map[string][2]string),
		getAsyncClient:    getAsyncClient,
		parseResponseFunc: nil, // Will be set by subclasses
		apiRequestFunc:    nil, // Will be set by subclasses
	}

	// Validate incoming credentials and region
	if account != nil && password != nil {
		bc.account = *account
		bc.password = *password
	} else if account != nil || password != nil {
		return nil, errors.New("account and password must be specified")
	} else {
		if region == nil {
			return nil, errors.New("unknown cloud region ''")
		}
		creds, ok := bc.CLOUD_CREDENTIALS[*region]
		if !ok {
			return nil, fmt.Errorf("unknown cloud region '%s'", *region)
		}
		bc.account = creds[0]
		bc.password = creds[1]
	}

	bc.baseURL = baseURL
	bc.loginID = nil
	bc.session = make(map[string]interface{})

	// Setup method for getting a client
	if getAsyncClient == nil {
		bc.getAsyncClient = func() *http.Client {
			return &http.Client{}
		}
	}

	return bc, nil
}

// timestamp formats a timestamp for the API.
func (bc *BaseCloud) timestamp() string {
	return time.Now().UTC().Format("20060102150405")
}

// parseResponse parses a response from the cloud.
// This method uses the parseResponseFunc if set, otherwise returns not implemented
func (bc *BaseCloud) parseResponse(response *http.Response) (interface{}, error) {
	if bc.parseResponseFunc != nil {
		return bc.parseResponseFunc(response)
	}
	return nil, errors.New("not implemented")
}

// postRequest posts a request to the cloud.
func (bc *BaseCloud) postRequest(urlStr string, headers map[string]string, rawData []byte, formData map[string]string, retries int) (map[string]interface{}, error) {
	if retries == 0 {
		retries = bc.RETRIES
	}

	client := bc.getAsyncClient()
	if client == nil {
		client = &http.Client{}
	}

	for retries > 0 {
		var req *http.Request
		var err error

		if rawData != nil {
			req, err = http.NewRequest("POST", urlStr, bytes.NewReader(rawData))
		} else if formData != nil {
			form := url.Values{}
			for k, v := range formData {
				form.Set(k, v)
			}
			req, err = http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
		} else {
			req, err = http.NewRequest("POST", urlStr, nil)
		}

		if err != nil {
			return nil, err
		}

		// Set Content-Type for form data
		if formData != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		// Set headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		client.Timeout = 10 * time.Second
		resp, err := client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
				if retries > 1 {
					retries--
					continue
				}
				return nil, CloudError
			}
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		// Parse the response
		result, err := bc.parseResponse(resp)
		if err != nil {
			return nil, err
		}

		if result == nil {
			return nil, nil
		}

		if m, ok := result.(map[string]interface{}); ok {
			return m, nil
		}

		return nil, nil
	}

	return nil, CloudError
}

// apiRequest makes a request to the cloud and return the results.
// This method uses the apiRequestFunc if set, otherwise returns not implemented
func (bc *BaseCloud) apiRequest(endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	if bc.apiRequestFunc != nil {
		return bc.apiRequestFunc(endpoint, body)
	}
	return nil, errors.New("not implemented")
}

// buildRequestBody builds a request body.
func (bc *BaseCloud) buildRequestBody(data map[string]interface{}) map[string]interface{} {
	// Set up the initial body
	body := map[string]interface{}{
		"appId":      bc.APP_ID,
		"src":        bc.APP_ID,
		"format":     bc.FORMAT,
		"clientType": bc.CLIENT_TYPE,
		"language":   bc.LANGUAGE,
		"deviceId":   bc.DEVICE_ID,
		"stamp":      bc.timestamp(),
	}

	// Add additional fields to the body
	for k, v := range data {
		body[k] = v
	}

	return body
}

// getLoginID gets a login ID for the cloud account.
func (bc *BaseCloud) getLoginID() (string, error) {
	response, err := bc.apiRequest(
		"/v1/user/login/id/get",
		bc.buildRequestBody(map[string]interface{}{
			"loginAccount": bc.account,
		}),
	)
	if err != nil {
		return "", err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return "", CloudError
	}

	loginID, ok := response["loginId"].(string)
	if !ok {
		return "", CloudError
	}

	return loginID, nil
}

// GetSession returns the current session.
func (bc *BaseCloud) GetSession() map[string]interface{} {
	return bc.session
}

// SetBaseURL sets the base URL for the cloud.
func (bc *BaseCloud) SetBaseURL(url string) {
	bc.baseURL = url
}

// GetToken gets token and key for the provided udpid.
func (bc *BaseCloud) GetToken(udpid string) (string, string, error) {
	response, err := bc.apiRequest(
		"/v1/iot/secure/getToken",
		bc.buildRequestBody(map[string]interface{}{
			"udpid": udpid,
		}),
	)
	if err != nil {
		return "", "", err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return "", "", CloudError
	}

	tokenList, ok := response["tokenlist"].([]interface{})
	if !ok {
		return "", "", CloudError
	}

	for _, t := range tokenList {
		token, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if token["udpId"] == udpid {
			tokenStr, _ := token["token"].(string)
			keyStr, _ := token["key"].(string)
			return tokenStr, keyStr, nil
		}
	}

	// No matching udpId in the tokenlist
	return "", "", fmt.Errorf("no token/key found for udpid %s", udpid)
}

// SmartHomeCloud is a class for minimal Midea SmartHome cloud access.
type SmartHomeCloud struct {
	*BaseCloud

	// Misc constants for the SmartHome cloud
	// APP_ID is inherited from BaseCloud, we'll set it in constructor

	// Base URLs
	BASE_URL       string
	BASE_URL_CHINA string

	// Internal fields
	accessToken string
	security    *SmartHomeCloudSecurity
}

// SmartHomeCloudSecurity is a class for SmartHome cloud specific security.
type SmartHomeCloudSecurity struct {
	HMAC_KEY  string
	IOT_KEY   string
	LOGIN_KEY string
	APP_KEY   string

	useChinaServer bool
}

// NewSmartHomeCloudSecurity creates a new SmartHomeCloudSecurity instance
func NewSmartHomeCloudSecurity(useChinaServer bool) *SmartHomeCloudSecurity {
	s := &SmartHomeCloudSecurity{
		HMAC_KEY: "PROD_VnoClJI9aikS8dyy",

		IOT_KEY:   "meicloud",
		LOGIN_KEY: "ac21b9f9cbfe4ca5a88562ef25e2b768",

		APP_KEY: "ac21b9f9cbfe4ca5a88562ef25e2b768",

		useChinaServer: useChinaServer,
	}

	if useChinaServer {
		s.IOT_KEY = "prod_secret123@muc"
		s.LOGIN_KEY = "ad0ee21d48a64bf49f4fb583ab76e799"
	}

	return s
}

// getIotKey gets the IOT key for the appropriate server.
func (s *SmartHomeCloudSecurity) getIotKey() string {
	return s.IOT_KEY
}

// getLoginKey gets the login key for the appropriate server.
func (s *SmartHomeCloudSecurity) getLoginKey() string {
	return s.LOGIN_KEY
}

// Sign generates a HMAC signature for the provided data and random data.
func (s *SmartHomeCloudSecurity) Sign(data string, random string) string {
	msg := s.getIotKey() + data + random

	h := hmac.New(sha256.New, []byte(s.HMAC_KEY))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// EncryptPassword encrypts the password for cloud password.
func (s *SmartHomeCloudSecurity) EncryptPassword(loginID string, password string) string {
	// Hash the password
	m1 := sha256.Sum256([]byte(password))

	// Create the login hash with the login ID + password hash + login key, then hash it all AGAIN
	loginHash := loginID + hex.EncodeToString(m1[:]) + s.getLoginKey()
	m2 := sha256.Sum256([]byte(loginHash))

	return hex.EncodeToString(m2[:])
}

// EncryptIAMPassword encrypts password for cloud iampwd field.
func (s *SmartHomeCloudSecurity) EncryptIAMPassword(loginID string, password string) string {
	// Hash the password
	m1 := md5.Sum([]byte(password))

	// Hash the password hash
	m2 := md5.Sum([]byte(hex.EncodeToString(m1[:])))

	if s.useChinaServer {
		return hex.EncodeToString(m2[:])
	}

	loginHash := loginID + hex.EncodeToString(m2[:]) + s.getLoginKey()
	sha := sha256.Sum256([]byte(loginHash))

	return hex.EncodeToString(sha[:])
}

// getAppKeyAndIV gets the app key and IV
func (s *SmartHomeCloudSecurity) getAppKeyAndIV() ([]byte, []byte) {
	hash := sha256.Sum256([]byte(s.APP_KEY))
	hexHash := hex.EncodeToString(hash[:])
	return []byte(hexHash[:16]), []byte(hexHash[16:32])
}

// EncryptAESAppKey encrypts data with AES using the app key
func (s *SmartHomeCloudSecurity) EncryptAESAppKey(data []byte) ([]byte, error) {
	key, iv := s.getAppKeyAndIV()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Pad data
	paddedData := pkcs7Pad(data, aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return ciphertext, nil
}

// DecryptAESAppKey decrypts data with AES using the app key
func (s *SmartHomeCloudSecurity) DecryptAESAppKey(data []byte) ([]byte, error) {
	key, iv := s.getAppKeyAndIV()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, data)

	// Unpad with proper error handling
	return pkcs7Unpad(plaintext)
}

// NewSmartHomeCloud creates a new SmartHomeCloud instance
func NewSmartHomeCloud(region string, account *string, password *string, useChinaServer bool, getAsyncClient func() *http.Client) (*SmartHomeCloud, error) {
	// Allow override China server from environment
	if os.Getenv("MIDEA_CHINA_SERVER") == "1" {
		useChinaServer = true
	}

	baseURL := "https://mp-prod.appsmb.com"
	if useChinaServer {
		baseURL = "https://mp-prod.smartmidea.net"
	}

	// Pre-define cloud credentials before creating BaseCloud
	cloudCredentials := map[string][2]string{
		"DE": {"midea_eu@mailinator.com", "das_ist_passwort1"},
		"KR": {"midea_sea@mailinator.com", "password_for_sea1"},
		"US": {"midea@mailinator.com", "this_is_a_password1"},
	}

	// Validate credentials: both must be provided, or neither (use defaults)
	if (account == nil) != (password == nil) {
		return nil, errors.New("account and password must be specified together")
	}

	// Use default credentials if no account/password provided
	if account == nil && password == nil {
		creds, ok := cloudCredentials[region]
		if !ok {
			return nil, fmt.Errorf("unknown cloud region '%s'", region)
		}
		account = &creds[0]
		password = &creds[1]
	}

	bc, err := NewBaseCloud(baseURL, &region, account, password, getAsyncClient)
	if err != nil {
		return nil, err
	}

	// Override APP_ID for SmartHome cloud
	bc.APP_ID = "1010"

	shc := &SmartHomeCloud{
		BaseCloud:      bc,
		BASE_URL:       "https://mp-prod.appsmb.com",
		BASE_URL_CHINA: "https://mp-prod.smartmidea.net",
		accessToken:    "",
		security:       NewSmartHomeCloudSecurity(useChinaServer),
	}

	// Set cloud credentials for reference
	bc.CLOUD_CREDENTIALS = cloudCredentials

	// Set the parseResponse function for this cloud type
	bc.parseResponseFunc = shc.parseResponse

	// Set the apiRequest function for this cloud type
	bc.apiRequestFunc = shc.apiRequest

	return shc, nil
}

// parseResponse parses a response from the cloud.
func (shc *SmartHomeCloud) parseResponse(response *http.Response) (interface{}, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	responseCode, ok := result["code"].(float64)
	if !ok {
		return nil, NewApiError("invalid response code", result["code"])
	}

	if int(responseCode) == 0 {
		return result["data"], nil
	}

	msg, _ := result["msg"].(string)
	return nil, NewApiError(msg, int(responseCode))
}

// apiRequest makes a request to the cloud and return the results.
func (shc *SmartHomeCloud) apiRequest(endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	// Encode body as JSON
	contents, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	random := generateTokenHex(16)

	// Sign the contents and add it to the header
	sign := shc.security.Sign(string(contents), random)
	headers := map[string]string{
		"Content-Type":  "application/json",
		"secretVersion": "1",
		"sign":          sign,
		"random":        random,
		"accessToken":   shc.accessToken,
	}

	// Build complete request URL
	urlStr := fmt.Sprintf("%s/mas/v5/app/proxy?alias=%s", shc.baseURL, endpoint)

	// Lock the API and post the request
	shc.apiLock.Lock()
	defer shc.apiLock.Unlock()

	return shc.postRequest(urlStr, headers, contents, nil, shc.RETRIES)
}

// buildRequestBody builds a request body.
func (shc *SmartHomeCloud) buildRequestBody(data map[string]interface{}) map[string]interface{} {
	// Set up the initial body
	body := shc.BaseCloud.buildRequestBody(map[string]interface{}{
		"reqId": generateTokenHex(16),
	})

	// Add additional fields to the body
	for k, v := range data {
		body[k] = v
	}

	return body
}

// Login logs in to the cloud.
func (shc *SmartHomeCloud) Login(force bool) error {
	// Don't login if session already exists
	if len(shc.session) > 0 && !force {
		return nil
	}

	// Get a login ID if we don't have one
	if shc.loginID == nil {
		loginID, err := shc.getLoginID()
		if err != nil {
			return err
		}
		shc.loginID = &loginID
	}

	// Build the login data
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"platform": shc.FORMAT,
			"deviceId": shc.DEVICE_ID,
		},
		"iotData": map[string]interface{}{
			"appId":        shc.APP_ID,
			"src":          shc.APP_ID,
			"clientType":   shc.CLIENT_TYPE,
			"loginAccount": shc.account,
			"iampwd":       shc.security.EncryptIAMPassword(*shc.loginID, shc.password),
			"password":     shc.security.EncryptPassword(*shc.loginID, shc.password),
			"pushToken":    generateTokenURLSafe(120),
			"stamp":        shc.timestamp(),
			"reqId":        generateTokenHex(16),
		},
	}

	// Login and store the session
	response, err := shc.apiRequest("/mj/user/login", body)
	if err != nil {
		return err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return CloudError
	}

	shc.session = response
	mdata, ok := response["mdata"].(map[string]interface{})
	if !ok {
		return CloudError
	}
	shc.accessToken, ok = mdata["accessToken"].(string)
	if !ok {
		return CloudError
	}

	return nil
}

// GetAccessToken returns the access token.
func (shc *SmartHomeCloud) GetAccessToken() string {
	return shc.accessToken
}

// GetProtocolLua fetches and decodes the protocol Lua file.
func (shc *SmartHomeCloud) GetProtocolLua(deviceType DeviceType, sn string) (string, string, error) {
	encryptedSn, err := shc.security.EncryptAESAppKey([]byte(sn))
	if err != nil {
		return "", "", err
	}

	response, err := shc.apiRequest(
		"/v2/luaEncryption/luaGet",
		shc.buildRequestBody(map[string]interface{}{
			"applianceMFCode": "0000",
			"applianceSn":     hex.EncodeToString(encryptedSn),
			"applianceType":   fmt.Sprintf("0x%x", deviceType),
			"encryptedType ":  2,
			"version":         "0",
		}),
	)
	if err != nil {
		return "", "", err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return "", "", CloudError
	}

	fileName, ok := response["fileName"].(string)
	if !ok {
		return "", "", CloudError
	}

	urlStr, ok := response["url"].(string)
	if !ok {
		return "", "", CloudError
	}

	client := shc.getAsyncClient()
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = 10 * time.Second

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	encryptedData, err := hex.DecodeString(string(body))
	if err != nil {
		return "", "", err
	}

	fileData, err := shc.security.DecryptAESAppKey(encryptedData)
	if err != nil {
		return "", "", err
	}

	return fileName, string(fileData), nil
}

// GetPlugin requests and downloads the device plugin.
func (shc *SmartHomeCloud) GetPlugin(deviceType DeviceType, sn string) (string, []byte, error) {
	if len(sn) < 17 {
		return "", nil, errors.New("serial number too short")
	}

	response, err := shc.apiRequest(
		"/v1/plugin/update/overseas/get",
		shc.buildRequestBody(map[string]interface{}{
			"clientVersion": "0",
			"uid":           generateTokenHex(16),
			"applianceList": []interface{}{
				map[string]interface{}{
					"appModel":    sn[9:17],
					"appType":     fmt.Sprintf("0x%x", deviceType),
					"modelNumber": "0",
				},
			},
		}),
	)
	if err != nil {
		return "", nil, err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return "", nil, CloudError
	}

	resultList, ok := response["result"].([]interface{})
	if !ok || len(resultList) == 0 {
		return "", nil, CloudError
	}

	result, ok := resultList[0].(map[string]interface{})
	if !ok {
		return "", nil, CloudError
	}

	fileName, ok := result["title"].(string)
	if !ok {
		return "", nil, CloudError
	}

	urlStr, ok := result["url"].(string)
	if !ok {
		return "", nil, CloudError
	}

	client := shc.getAsyncClient()
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = 10 * time.Second

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return fileName, fileData, nil
}

// NetHomePlusCloud is a class for minimal NetHome Plus cloud access.
type NetHomePlusCloud struct {
	*BaseCloud

	// Misc constants for the NetHome Plus cloud
	// APP_ID is inherited from BaseCloud, we'll set it in constructor

	BASE_URL string

	// Internal fields
	sessionID string
	security  *NetHomePlusCloudSecurity
}

// NetHomePlusCloudSecurity is a class for NetHome Plus cloud specific security.
type NetHomePlusCloudSecurity struct {
	// NetHome Plus
	APP_KEY string
}

// NewNetHomePlusCloudSecurity creates a new NetHomePlusCloudSecurity instance
func NewNetHomePlusCloudSecurity() *NetHomePlusCloudSecurity {
	return &NetHomePlusCloudSecurity{
		APP_KEY: "3742e9e5842d4ad59c2db887e12449f9",
	}
}

// Sign generates a signature for the provided data and URL.
func (s *NetHomePlusCloudSecurity) Sign(urlStr string, data map[string]interface{}) string {
	// Get path portion of request
	// If urlStr is already a path (starts with '/'), use it directly
	var path string
	if strings.HasPrefix(urlStr, "/") {
		path = urlStr
	} else {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			path = urlStr // Fallback to using urlStr directly
		} else {
			path = parsedURL.Path
		}
	}

	// Sort request and create a query string (matching Python's behavior)
	// Python: query = unquote_plus(urlencode(sorted(data.items())))
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	// First URL encode, then decode (matching Python's unquote_plus behavior)
	var pairs []string
	for _, k := range keys {
		// Convert value to string and URL encode it
		v := url.QueryEscape(fmt.Sprintf("%v", data[k]))
		// Decode it back (unquote_plus in Python decodes '+' to space and %XX to chars)
		vDecoded, err := url.QueryUnescape(v)
		if err == nil {
			v = vDecoded
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	query := strings.Join(pairs, "&")

	msg := path + query + s.APP_KEY

	hash := sha256.Sum256([]byte(msg))
	sign := hex.EncodeToString(hash[:])
	return sign
}

// EncryptPassword encrypts the login password.
func (s *NetHomePlusCloudSecurity) EncryptPassword(loginID string, password string) string {
	// Hash the password
	m1 := sha256.Sum256([]byte(password))

	// Create the login hash with the login ID + password hash + app key, then hash it all AGAIN
	loginHash := loginID + hex.EncodeToString(m1[:]) + s.APP_KEY
	m2 := sha256.Sum256([]byte(loginHash))

	return hex.EncodeToString(m2[:])
}

// NewNetHomePlusCloud creates a new NetHomePlusCloud instance
func NewNetHomePlusCloud(region string, account *string, password *string, getAsyncClient func() *http.Client) (*NetHomePlusCloud, error) {
	baseURL := "https://mapp.appsmb.com"

	// Pre-define cloud credentials before creating BaseCloud
	cloudCredentials := map[string][2]string{
		"DE": {"nethome+de@mailinator.com", "password1"},
		"KR": {"nethome+sea@mailinator.com", "password1"},
		"US": {"nethome+us@mailinator.com", "password1"},
	}

	// Validate credentials: both must be provided, or neither (use defaults)
	if (account == nil) != (password == nil) {
		return nil, errors.New("account and password must be specified together")
	}

	// Use default credentials if no account/password provided
	if account == nil && password == nil {
		creds, ok := cloudCredentials[region]
		if !ok {
			return nil, fmt.Errorf("unknown cloud region '%s'", region)
		}
		account = &creds[0]
		password = &creds[1]
	}

	bc, err := NewBaseCloud(baseURL, &region, account, password, getAsyncClient)
	if err != nil {
		return nil, err
	}

	// Override APP_ID for NetHome Plus cloud
	bc.APP_ID = "1017"

	nhpc := &NetHomePlusCloud{
		BaseCloud: bc,
		BASE_URL:  "https://mapp.appsmb.com",
		sessionID: "",
		security:  NewNetHomePlusCloudSecurity(),
	}

	// Set cloud credentials for reference
	bc.CLOUD_CREDENTIALS = cloudCredentials

	// Set the parseResponse function for this cloud type
	bc.parseResponseFunc = nhpc.parseResponse

	// Set the apiRequest function for this cloud type
	bc.apiRequestFunc = nhpc.apiRequest

	return nhpc, nil
}

// parseResponse parses a response from the cloud.
func (nhpc *NetHomePlusCloud) parseResponse(response *http.Response) (interface{}, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	responseCode, ok := result["errorCode"].(float64)
	if !ok {
		// Try parsing as string
		if codeStr, ok := result["errorCode"].(string); ok {
			if code, err := strconv.Atoi(codeStr); err == nil {
				if code == 0 {
					return result["result"], nil
				}
				msg, _ := result["msg"].(string)
				return nil, NewApiError(msg, code)
			}
		}
		return nil, NewApiError("invalid error code", result["errorCode"])
	}

	if int(responseCode) == 0 {
		return result["result"], nil
	}

	msg, _ := result["msg"].(string)
	return nil, NewApiError(msg, int(responseCode))
}

// apiRequest makes a request to the cloud and return the results.
func (nhpc *NetHomePlusCloud) apiRequest(endpoint string, body map[string]interface{}) (map[string]interface{}, error) {
	// Sign the contents and add it to the body
	body["sign"] = nhpc.security.Sign(endpoint, body)

	// Build complete request URL
	urlStr := fmt.Sprintf("%s%s", nhpc.baseURL, endpoint)

	// Lock the API and post the request
	nhpc.apiLock.Lock()
	defer nhpc.apiLock.Unlock()

	return nhpc.postRequest(urlStr, nil, nil, mapStringInterfaceToMapString(body), nhpc.RETRIES)
}

// buildRequestBody builds a request body.
func (nhpc *NetHomePlusCloud) buildRequestBody(data map[string]interface{}) map[string]interface{} {
	// Set up the initial body
	body := nhpc.BaseCloud.buildRequestBody(map[string]interface{}{
		"sessionId": nhpc.sessionID,
	})

	// Add additional fields to the body
	for k, v := range data {
		body[k] = v
	}

	return body
}

// Login logs in to the cloud.
func (nhpc *NetHomePlusCloud) Login(force bool) error {
	// Don't login if session already exists
	if len(nhpc.session) > 0 && !force {
		return nil
	}

	// Get a login ID if we don't have one
	if nhpc.loginID == nil {
		loginID, err := nhpc.getLoginID()
		if err != nil {
			return err
		}
		nhpc.loginID = &loginID
	}

	// Login and store the session
	response, err := nhpc.apiRequest(
		"/v1/user/login",
		nhpc.buildRequestBody(map[string]interface{}{
			"loginAccount": nhpc.account,
			"password":     nhpc.security.EncryptPassword(*nhpc.loginID, nhpc.password),
		}),
	)
	if err != nil {
		return err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return CloudError
	}

	nhpc.session = response
	sessionID, ok := response["sessionId"].(string)
	if !ok {
		return CloudError
	}
	nhpc.sessionID = sessionID

	return nil
}

// GetSessionID returns the session ID.
func (nhpc *NetHomePlusCloud) GetSessionID() string {
	return nhpc.sessionID
}

// GetToken gets token and key for the provided udpid.
// This method overrides BaseCloud.GetToken to include sessionId in the request.
func (nhpc *NetHomePlusCloud) GetToken(udpid string) (string, string, error) {
	response, err := nhpc.apiRequest(
		"/v1/iot/secure/getToken",
		nhpc.buildRequestBody(map[string]interface{}{
			"udpid": udpid,
		}),
	)
	if err != nil {
		return "", "", err
	}

	// Assert response is not None since we should throw on errors
	if response == nil {
		return "", "", CloudError
	}

	tokenList, ok := response["tokenlist"].([]interface{})
	if !ok {
		return "", "", CloudError
	}

	for _, t := range tokenList {
		token, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if token["udpId"] == udpid {
			tokenStr, _ := token["token"].(string)
			keyStr, _ := token["key"].(string)
			return tokenStr, keyStr, nil
		}
	}

	// No matching udpId in the tokenlist
	return "", "", fmt.Errorf("no token/key found for udpid %s", udpid)
}

// GetProtocolLua fetches and decodes the protocol Lua file.
func (nhpc *NetHomePlusCloud) GetProtocolLua(deviceType DeviceType, sn string) (string, string, error) {
	return "", "", errors.New("not implemented")
}

// GetPlugin requests and downloads the device plugin.
func (nhpc *NetHomePlusCloud) GetPlugin(deviceType DeviceType, sn string) (string, []byte, error) {
	return "", nil, errors.New("not implemented")
}

// Helper functions

// generateTokenHex generates a random hex string of specified length
func generateTokenHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n*2]
}

// generateTokenURLSafe generates a URL-safe random string (base64 URL-safe encoding)
func generateTokenURLSafe(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// pkcs7Pad pads data using PKCS7
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding with proper validation
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, errors.New("invalid padding")
	}

	// Verify padding bytes
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// mapStringInterfaceToMapString converts map[string]interface{} to map[string]string
func mapStringInterfaceToMapString(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
