// Package tests provides tests for the msmart cloud functionality.
package tests

import (
	"net/http"
	"os"
	"testing"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

// skipIfCI skips the test if running in CI environment
func skipIfCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Tests skipped on CI")
	}
}

// TestNetHomePlusCloud_Login tests that we can login to the NetHome Plus cloud.
func TestNetHomePlusCloud_Login(t *testing.T) {
	skipIfCI(t)

	// Skip if no real credentials provided (these tests require actual cloud account)
	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	client, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, &password, nil)
	if err != nil {
		t.Fatalf("Failed to create NetHomePlusCloud: %v", err)
	}

	err = client.Login(false)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Verify session is set
	session := client.GetSession()
	if session == nil || len(session) == 0 {
		t.Error("Expected session to be set after login")
	}

	sessionID := client.GetSessionID()
	if sessionID == "" {
		t.Error("Expected sessionID to be set after login")
	}
}

// TestNetHomePlusCloud_LoginException tests that bad credentials raise an error.
func TestNetHomePlusCloud_LoginException(t *testing.T) {
	skipIfCI(t)

	account := "bad@account.com"
	password := "not_a_password"

	client, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, &password, nil)
	if err != nil {
		t.Fatalf("Failed to create NetHomePlusCloud: %v", err)
	}

	err = client.Login(false)
	if err == nil {
		t.Error("Expected error for bad credentials, got nil")
	}
	// The error should be either ApiError or CloudError
}

// TestNetHomePlusCloud_InvalidRegion tests that an invalid region raises an error.
func TestNetHomePlusCloud_InvalidRegion(t *testing.T) {
	skipIfCI(t)

	_, err := msmart.NewNetHomePlusCloud("NOT_A_REGION", nil, nil, nil)
	if err == nil {
		t.Error("Expected error for invalid region, got nil")
	}
}

// TestNetHomePlusCloud_InvalidCredentials tests that invalid credentials raise an error.
func TestNetHomePlusCloud_InvalidCredentials(t *testing.T) {
	skipIfCI(t)

	password := "some_password"
	_, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, nil, &password, nil)
	if err == nil {
		t.Error("Expected error when only password is specified, got nil")
	}

	account := "some_account"
	_, err = msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, nil, nil)
	if err == nil {
		t.Error("Expected error when only account is specified, got nil")
	}
}

// TestNetHomePlusCloud_GetToken tests that a token and key can be obtained from the cloud.
func TestNetHomePlusCloud_GetToken(t *testing.T) {
	skipIfCI(t)

	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	dummyUDPID := "4fbe0d4139de99dd88a0285e14657045"

	client, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, &password, nil)
	if err != nil {
		t.Fatalf("Failed to create NetHomePlusCloud: %v", err)
	}

	err = client.Login(false)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	token, key, err := client.GetToken(dummyUDPID)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be non-empty")
	}
	if key == "" {
		t.Error("Expected key to be non-empty")
	}
}

// TestNetHomePlusCloud_GetTokenException tests that an exception is thrown when a token and key
// can't be obtained from the cloud.
func TestNetHomePlusCloud_GetTokenException(t *testing.T) {
	skipIfCI(t)

	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	badUDPID := "NOT_A_UDPID"

	client, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, &password, nil)
	if err != nil {
		t.Fatalf("Failed to create NetHomePlusCloud: %v", err)
	}

	err = client.Login(false)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	_, _, err = client.GetToken(badUDPID)
	if err == nil {
		t.Error("Expected error for bad UDPID, got nil")
	}
}

// TestNetHomePlusCloud_ConnectException tests that an exception is thrown when the cloud connection fails.
func TestNetHomePlusCloud_ConnectException(t *testing.T) {
	skipIfCI(t)

	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	// Create a client with a custom HTTP client that will fail
	client, err := msmart.NewNetHomePlusCloud(msmart.DefaultCloudRegion, &account, &password, func() *http.Client {
		return &http.Client{
			Timeout: 1, // Very short timeout to force failure
		}
	})
	if err != nil {
		t.Fatalf("Failed to create NetHomePlusCloud: %v", err)
	}

	// Override base URL to an invalid domain
	client.SetBaseURL("https://fake_server.invalid.")

	err = client.Login(false)
	if err == nil {
		t.Error("Expected error for invalid server, got nil")
	}
}

// TestSmartHomeCloud_Login tests that we can login to the SmartHome cloud.
func TestSmartHomeCloud_Login(t *testing.T) {
	skipIfCI(t)

	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	client, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, nil)
	if err != nil {
		t.Fatalf("Failed to create SmartHomeCloud: %v", err)
	}

	err = client.Login(false)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Verify session is set
	session := client.GetSession()
	if session == nil || len(session) == 0 {
		t.Error("Expected session to be set after login")
	}

	accessToken := client.GetAccessToken()
	if accessToken == "" {
		t.Error("Expected accessToken to be set after login")
	}
}

// TestSmartHomeCloud_LoginException tests that bad credentials raise an error.
func TestSmartHomeCloud_LoginException(t *testing.T) {
	skipIfCI(t)

	account := "bad@account.com"
	password := "not_a_password"

	client, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, nil)
	if err != nil {
		t.Fatalf("Failed to create SmartHomeCloud: %v", err)
	}

	err = client.Login(false)
	if err == nil {
		t.Error("Expected error for bad credentials, got nil")
	}
}

// TestSmartHomeCloud_InvalidRegion tests that an invalid region raises an error.
func TestSmartHomeCloud_InvalidRegion(t *testing.T) {
	skipIfCI(t)

	_, err := msmart.NewSmartHomeCloud("NOT_A_REGION", nil, nil, false, nil)
	if err == nil {
		t.Error("Expected error for invalid region, got nil")
	}
}

// TestSmartHomeCloud_InvalidCredentials tests that invalid credentials raise an error.
func TestSmartHomeCloud_InvalidCredentials(t *testing.T) {
	skipIfCI(t)

	password := "some_password"
	_, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, nil, &password, false, nil)
	if err == nil {
		t.Error("Expected error when only password is specified, got nil")
	}

	account := "some_account"
	_, err = msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, nil, false, nil)
	if err == nil {
		t.Error("Expected error when only account is specified, got nil")
	}
}

// TestSmartHomeCloud_GetToken is disabled until the broken API is fixed.
// See Python test for details.
func TestSmartHomeCloud_GetToken(t *testing.T) {
	skipIfCI(t)
	t.Skip("Get token tests disabled until we can solve the broken API")
}

// TestSmartHomeCloud_GetTokenException is disabled until the broken API is fixed.
// See Python test for details.
func TestSmartHomeCloud_GetTokenException(t *testing.T) {
	skipIfCI(t)
	t.Skip("Get token tests disabled until we can solve the broken API")
}

// TestSmartHomeCloud_ConnectException tests that an exception is thrown when the cloud connection fails.
func TestSmartHomeCloud_ConnectException(t *testing.T) {
	skipIfCI(t)

	account := os.Getenv("MIDEA_ACCOUNT")
	password := os.Getenv("MIDEA_PASSWORD")
	if account == "" || password == "" {
		t.Skip("Skipping: requires MIDEA_ACCOUNT and MIDEA_PASSWORD environment variables")
	}

	// Create a client with a custom HTTP client that will fail
	client, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, func() *http.Client {
		return &http.Client{
			Timeout: 1, // Very short timeout to force failure
		}
	})
	if err != nil {
		t.Fatalf("Failed to create SmartHomeCloud: %v", err)
	}

	// Override base URL to an invalid domain
	client.SetBaseURL("https://fake_server.invalid.")

	err = client.Login(false)
	if err == nil {
		t.Error("Expected error for invalid server, got nil")
	}
}
