// midea - CLI tool for controlling Midea smart air conditioners
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RelicOfTesla/midea-msmart/cmd/config"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
	"github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
	"github.com/RelicOfTesla/midea-msmart/msmart/device/cc"
)

var version = "1.0.0"

// DeviceType string constants for CLI
const (
	DeviceTypeAC = "AC"  // Air Conditioner (空调)
	DeviceTypeCC = "CC"  // Commercial Air Conditioner (商业空调)
)

// deviceTypeMap maps CLI device type strings to msmart DeviceType
var deviceTypeMap = map[string]msmart.DeviceType{
	DeviceTypeAC: msmart.DeviceTypeAirConditioner,
	DeviceTypeCC: msmart.DeviceTypeCommercialAC,
}

func main() {
	// Check for global verbose flag
	verbose := false
	for _, arg := range os.Args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
			msmart.Verbose = true
			break
		}
	}
	_ = verbose // Avoid unused variable warning

	// Parse global --region flag
	region := msmart.DefaultCloudRegion
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--region" && i+1 < len(os.Args) {
			region = os.Args[i+1]
			i++
			break
		}
	}

	// Parse global --device_type flag
	deviceTypeStr := DeviceTypeAC // Default to AC (Air Conditioner)
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--device_type" && i+1 < len(os.Args) {
			deviceTypeStr = strings.ToUpper(os.Args[i+1])
			// Validate device type
			if _, ok := deviceTypeMap[deviceTypeStr]; !ok {
				fmt.Printf("❌ 不支持的设备类型: %s\n", os.Args[i+1])
				fmt.Println("   支持的设备类型: AC (空调), CC (商业空调)")
				os.Exit(1)
			}
			i++
			break
		}
	}

	// Parse global --id flag (Device ID for V3 devices)
	deviceID := 0
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--id" && i+1 < len(os.Args) {
			id, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Printf("❌ 无效的设备 ID: %s\n", os.Args[i+1])
				os.Exit(1)
			}
			deviceID = id
			i++
			break
		}
	}

	// Parse global --token flag (Authentication token for V3 devices)
	var deviceToken string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--token" && i+1 < len(os.Args) {
			deviceToken = os.Args[i+1]
			i++
			break
		}
	}

	// Parse global --key flag (Authentication key for V3 devices)
	var deviceKey string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--key" && i+1 < len(os.Args) {
			deviceKey = os.Args[i+1]
			i++
			break
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	configPath := config.DefaultConfigPath()

	// Commands that don't need config
	switch command {
	case "help", "-h", "--help":
		printUsage()
		return
	case "version", "--version":
		fmt.Printf("midea %s\n", version)
		return
	}

	// Execute command
	switch command {
	case "discover":
		handleDiscover(configPath, region)
	case "list":
		handleList(configPath)
	case "bind":
		handleBind(configPath)
	case "unbind":
		handleUnbind(configPath)
	case "status":
		handleStatus(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "on":
		handlePower(configPath, true, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "off":
		handlePower(configPath, false, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "temp":
		handleTemp(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "mode":
		handleMode(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "fan":
		handleFan(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "swing":
		handleSwing(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "set":
		handleSet(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "query":
		handleQuery(configPath, deviceTypeStr, deviceID, deviceToken, deviceKey)
	case "download":
		handleDownload(configPath, region)
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
midea - 美的空调控制 CLI v` + version + `

用法:
  midea [-v|--verbose] [--region <地区>] [--device_type <类型>] <command> [arguments]

全局选项:
  -v, --verbose        显示详细调试日志
  --region <地区>      云端服务地区 (DE, KR, US), 默认: US
  --device_type <类型> 设备类型: AC (空调), CC (商业空调), 默认: AC

命令:
  discover [--auto-connect|-a] [--account <账号> --password <密码>]
                                发现设备并保存到配置
                                --auto-connect: 自动连接并获取V3设备的token
                                --account/--password: 美的账号密码 (V3设备认证需要)
  list                          列出已保存的设备
  bind <id|sn|ip> -n <名称>   绑定设备别名
  unbind <name|id>            解绑设备

  status <name|id> [--auto] [--capabilities] [--energy]
                                查询设备状态
                                --auto: 自动发现设备并获取token
                                --capabilities: 显示设备能力信息
                                --energy: 显示能耗信息
  on <name|id>                开机
  off <name|id>               关机
  temp <name|id> <温度>       设置温度 (范围: 16-30°C)
  mode <name|id> <模式>       设置运行模式
  fan <name|id> <风速>        设置风速
  swing <name|id> <模式>      设置摆风模式
  set <name|id> [选项]        多参数设置 (一次设置多个属性)
  query <name|id> [key] [--all] [--auto]
                                查询设备属性
                                key: 属性名称 (如: temp, mode, fan, swing, power)
                                --all: 显示所有属性 (默认)
                                --auto: 自动发现设备并获取token
  download <host> [--account <账号> --password <密码>]
                                下载设备的 Lua 协议和插件
                                --account/--password: 美的账号密码 (下载需要)

参数范围:
  温度: 16, 17, 18, ..., 29, 30 (°C)
  模式: cool(制冷), heat(制热), auto(自动), dry(除湿), fan(送风)
  风速: auto(自动), low(低), medium(中), high(高), silent(静音)
  摆风: off(关闭), vertical(上下), horizontal(左右), both(全方位)

set命令选项:
  --temp <温度>      设置温度
  --mode <模式>      设置运行模式
  --fan <风速>       设置风速
  --swing <模式>     设置摆风
  --power <on|off>   设置电源

示例:
  midea discover                      # 发现局域网内的设备
  midea -v discover --auto-connect   # 使用verbose模式发现设备并自动获取V3设备token
  midea discover --auto-connect --account your@email.com --password yourpass
                                      # 使用自定义账号发现设备
  midea list                          # 列出已保存的设备
  midea bind 192.168.1.60 -n 客厅    # 绑定IP为192.168.1.60的设备,命名为"客厅"
  midea status 客厅                   # 查询“客厅”空调状态
  midea status 客厅 --capabilities    # 查询“客厅”空调状态并显示设备能力
  midea status 客厅 --energy          # 查询“客厅”空调状态并显示能耗信息
  midea on 客厅                       # 打开“客厅”空调
  midea temp 客厅 26                  # 设置温度为26°C
  midea mode 客厅 cool                # 设置为制冷模式
  midea fan 客厅 high                 # 设置为高风速
  midea swing 客厅 vertical           # 设置为上下摆风

  # 多参数设置 (一次命令设置多个属性)
  midea set 客厅 --temp 26 --mode cool --fan high
  midea set 客厅 --power on --temp 24

  # 下载设备协议和插件
  midea download 192.168.1.60 --account your@email.com --password yourpass

配置文件: 当前目录的 midea.json (优先) 或 ~/.config/midea/config.json
`)
}

// ============================================================================
// Discovery Commands
// ============================================================================

func handleDiscover(configPath string, region string) {
	fmt.Println("🔍 正在发现设备...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// Parse flags
	autoConnect := false
	var account, password string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--auto-connect", "-a":
			autoConnect = true
		case "--account":
			if i+1 < len(os.Args) {
				account = os.Args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(os.Args) {
				password = os.Args[i+1]
				i++
			}
		}
	}

	// Discover devices
	discoverConfig := &msmart.DiscoverConfig{
		Timeout:          5 * time.Second,
		DiscoveryPackets: 3,
		AutoConnect:      autoConnect,
		Region:           region,
	}

	// Set account and password if provided
	if account != "" && password != "" {
		discoverConfig.Account = account
		discoverConfig.Password = password
	}

	devices, err := msmart.Discover(ctx, discoverConfig)

	// Even if there's an error, check if we discovered any devices
	if err != nil && len(devices) == 0 {
		fmt.Printf("❌ 发现设备失败: %v\n", err)
		os.Exit(1)
	}

	// Log warning if there was an error but we have devices
	if err != nil && len(devices) > 0 {
		fmt.Printf("⚠️  发现设备时有错误: %v\n", err)
	}

	if len(devices) == 0 {
		fmt.Println("⚠️  未发现任何设备")
		return
	}

	fmt.Printf("\n✅ 发现 %d 台设备\n\n", len(devices))

	for _, d := range devices {
		// Get device info
		deviceID := fmt.Sprintf("%d", d.GetID())
		deviceType := "未知设备"
		if d.GetType() == msmart.DeviceTypeAirConditioner {
			deviceType = "空调"
		}

		// Get name
		name := ""
		if n := d.GetName(); n != nil {
			name = *n
		}

		// Get SN
		sn := ""
		if s := d.GetSN(); s != nil {
			sn = *s
		}

		// Get version
		version := 2
		if v := d.GetVersion(); v != nil {
			version = *v
		}

		// Get token/key
		var token, key string
		if t := d.GetToken(); t != nil {
			token = *t
		}
		if k := d.GetKey(); k != nil {
			key = *k
		}

		// Check if device already exists in config
		existingDevice := cfg.GetDevice(deviceID)
		if existingDevice != nil {
			// Keep the existing name
			name = existingDevice.Name
		}

		// Save device to config
		device := config.Device{
			ID:      deviceID,
			Name:    name,
			IP:      d.GetIP(),
			Port:    d.GetPort(),
			SN:      sn,
			Type:    int(d.GetType()),
			Token:   token,
			Key:     key,
			Version: version,
			Online:  d.GetOnline(),
		}
		cfg.AddDevice(device)

		// Print device info
		status := "🔴 离线"
		if d.GetOnline() {
			status = "🟢 在线"
		}

		displayName := name
		if displayName == "" {
			displayName = "(未命名)"
		}

		fmt.Printf("  %s %s - %s (%s) [ID: %s]\n", status, deviceType, displayName, d.GetIP(), deviceID)
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("\n❌ 保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n💾 配置已保存到: %s\n", configPath)
	fmt.Println("\n💡 使用 'midea bind <id|ip> -n <名称>' 来为设备命名")
}

func handleList(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	devices := cfg.ListDevices()
	if len(devices) == 0 {
		fmt.Println("📋 暂无设备配置")
		fmt.Println("💡 使用 'midea discover' 来发现设备")
		return
	}

	fmt.Println("\n📋 设备列表:")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for _, d := range devices {
		status := "🔴 离线"
		if d.Online {
			status = "🟢 在线"
		}

		deviceType := "未知设备"
		if d.Type == 0xAC {
			deviceType = "空调"
		}

		name := d.Name
		if name == "" {
			name = "(未命名)"
		}

		fmt.Printf("  %s %-12s - %-8s %s (ID: %s)\n", status, name, deviceType, d.IP, d.ID)
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("配置文件: %s\n", configPath)
}

// ============================================================================
// Device Management Commands
// ============================================================================

func handleBind(configPath string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea bind <id|sn|ip> -n <名称>")
		os.Exit(1)
	}

	identifier := os.Args[2]
	name := ""

	// Parse -n flag
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "-n" && i+1 < len(os.Args) {
			name = os.Args[i+1]
			break
		}
	}

	if name == "" {
		fmt.Println("❌ 请指定名称: midea bind <id|sn|ip> -n <名称>")
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if !cfg.BindName(identifier, name) {
		fmt.Printf("❌ 未找到设备: %s\n", identifier)
		fmt.Println("💡 使用 'midea list' 查看设备列表")
		os.Exit(1)
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("❌ 保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已绑定: %s -> %s\n", identifier, name)
}

func handleUnbind(configPath string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea unbind <name|id>")
		os.Exit(1)
	}

	identifier := os.Args[2]

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if !cfg.RemoveDevice(identifier) {
		fmt.Printf("❌ 未找到设备: %s\n", identifier)
		fmt.Println("💡 使用 'midea list' 查看设备列表")
		os.Exit(1)
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("❌ 保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 已解绑: %s\n", identifier)
}

// ============================================================================
// Device Control Commands
// ============================================================================

// Device is a common interface for all device types
// This allows command functions to work with different device types
type Device interface {
	// Basic device operations (these are common to all devices)
	GetIP() string
	GetPort() int
}

// mustGetACDevice extracts an AC device from interface{}, exits if not AC type
func mustGetACDevice(device interface{}) *ac.AirConditioner {
	acDevice, ok := device.(*ac.AirConditioner)
	if !ok {
		fmt.Println("❌ 此命令只支持空调设备 (AC)")
		fmt.Println("💡 使用 --device_type AC 指定空调设备")
		os.Exit(1)
	}
	return acDevice
}

func getDevice(configPath, identifier string, deviceTypeStr string) (*config.Device, interface{}) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	device := cfg.GetDevice(identifier)
	if device == nil {
		fmt.Printf("❌ 未找到设备: %s\n", identifier)
		fmt.Println("💡 使用 'midea list' 查看设备列表 或使用 --auto 自动发现设备")
		os.Exit(1)
	}

	// Use device type from parameter, or from config, or default to AC
	effectiveType := deviceTypeStr
	if effectiveType == "" {
		if device.Type == int(msmart.DeviceTypeCommercialAC) {
			effectiveType = DeviceTypeCC
		} else {
			effectiveType = DeviceTypeAC // Default
		}
	}

	// Parse device ID
	deviceID, err := strconv.ParseInt(device.ID, 10, 64)
	if err != nil {
		fmt.Printf("❌ 无效的设备ID: %s\n", device.ID)
		os.Exit(1)
	}

	// Create device based on type
	switch effectiveType {
	case DeviceTypeCC:
		// Create Commercial Air Conditioner
		ccDevice := cc.NewCommercialAirConditioner(device.IP, int(deviceID), device.Port)
		fmt.Println("ℹ️  商业空调设备 (CC) 支持有限，部分命令可能不可用")
		return device, ccDevice
	default:
		// Create Air Conditioner (default)
		acDevice := ac.NewAirConditioner(
			device.IP,
			device.Port,
			int(deviceID),
			msmart.WithName(device.Name),
			msmart.WithVersion(device.Version),
		)

		// Set token and key if available (only for V3 devices)
		if device.Version == 3 {
			if device.Token == "" || device.Key == "" {
				fmt.Printf("❌ V3设备需要token和key进行认证\n")
				fmt.Println("💡 使用 '--auto' 参数自动获取token/key，或使用 'midea discover --auto-connect' 重新发现设备")
				os.Exit(1)
			}

			token, err := hex.DecodeString(device.Token)
			if err != nil {
				fmt.Printf("❌ 无效的Token: %v\n", err)
				os.Exit(1)
			}
			key, err := hex.DecodeString(device.Key)
			if err != nil {
				fmt.Printf("❌ 无效的Key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("🔐 正在认证...")
			if err := acDevice.Authenticate(token, key); err != nil {
				fmt.Printf("❌ 认证失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ 认证成功")
		}
		return device, acDevice
	}
}

// getDeviceDirect creates a device directly with host, id, token and key
// This is used when --id, --token, --key are provided (similar to Python CLI)
func getDeviceDirect(host string, deviceID int, tokenStr, keyStr string, deviceTypeStr string) (*config.Device, interface{}) {
	// Create a dummy config device for display purposes
	device := &config.Device{
		ID:      fmt.Sprintf("%d", deviceID),
		Name:    "(Direct)",
		IP:      host,
		Port:    6444,
		Version: 3, // V3 devices require token/key
		Online:  true,
	}

	// Use device type from parameter or default to AC
	effectiveType := deviceTypeStr
	if effectiveType == "" {
		effectiveType = DeviceTypeAC
	}

	// Create device based on type
	switch effectiveType {
	case DeviceTypeCC:
		// Create Commercial Air Conditioner
		ccDevice := cc.NewCommercialAirConditioner(host, deviceID, 6444)
		fmt.Println("ℹ️  商业空调设备 (CC) 支持有限，部分命令可能不可用")
		return device, ccDevice
	default:
		// Create Air Conditioner (default)
		acDevice := ac.NewAirConditioner(
			host,
			6444,
			deviceID,
			msmart.WithName("(Direct)"),
			msmart.WithVersion(3),
		)

		// Authenticate if token and key are provided
		if tokenStr != "" && keyStr != "" {
			token, err := hex.DecodeString(tokenStr)
			if err != nil {
				fmt.Printf("❌ 无效的Token: %v\n", err)
				os.Exit(1)
			}
			key, err := hex.DecodeString(keyStr)
			if err != nil {
				fmt.Printf("❌ 无效的Key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("🔐 正在认证...")
			if err := acDevice.Authenticate(token, key); err != nil {
				fmt.Printf("❌ 认证失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ 认证成功")
		}

		return device, acDevice
	}
}

// getDeviceAuto automatically discovers a device and gets token/key for V3 devices
func getDeviceAuto(identifier string, configPath string, deviceTypeStr string) (*config.Device, interface{}) {
	fmt.Printf("🔍 正在自动发现设备: %s\n", identifier)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Discover the device
	discoverConfig := &msmart.DiscoverConfig{
		Target:          identifier,
		Timeout:         5 * time.Second,
		DiscoveryPackets: 3,
		AutoConnect:      true, // Enable auto-connect to get token/key
	}

	devices, err := msmart.Discover(ctx, discoverConfig)
	if err != nil {
		fmt.Printf("❌ 发现设备失败: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("❌ 未找到设备")
		os.Exit(1)
	}

	// Get the first discovered device
	d := devices[0]

	// Get device info
	deviceID := d.GetID()
	deviceType := "未知设备"
	if d.GetType() == msmart.DeviceTypeAirConditioner {
		deviceType = "空调"
	} else if d.GetType() == msmart.DeviceTypeCommercialAC {
		deviceType = "商业空调"
	}

	// Get name
	name := ""
	if n := d.GetName(); n != nil {
		name = *n
	}

	// Get SN
	sn := ""
	if s := d.GetSN(); s != nil {
		sn = *s
	}

	// Get version
	version := 2
	if v := d.GetVersion(); v != nil {
		version = *v
	}

	// Get token/key
	var token, key string
	if t := d.GetToken(); t != nil {
		token = *t
	}
	if k := d.GetKey(); k != nil {
		key = *k
	}

	fmt.Printf("✅ 发现设备: %s (%s) [ID: %d, 版本: V%d]\n", deviceType, d.GetIP(), deviceID, version)

	// Check if it's a V3 device
	if version == 3 && (token == "" || key == "") {
		os.Exit(1)
	}

	// Load config and save device
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// Create config device
	device := &config.Device{
		ID:      fmt.Sprintf("%d", deviceID),
		Name:    name,
		IP:      d.GetIP(),
		Port:    d.GetPort(),
		SN:      sn,
		Type:    int(d.GetType()),
		Token:   token,
		Key:     key,
		Version: version,
		Online:  d.GetOnline(),
	}

	// Check if device already exists
	existingDevice := cfg.GetDevice(fmt.Sprintf("%d", deviceID))
	if existingDevice != nil {
		// Keep the existing name
		device.Name = existingDevice.Name
	}

	// Save to config
	cfg.AddDevice(*device)
	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("⚠️  保存配置失败: %v\n", err)
	}

	// Use device type from parameter, or from discovered device, or default to AC
	effectiveType := deviceTypeStr
	if effectiveType == "" {
		if d.GetType() == msmart.DeviceTypeCommercialAC {
			effectiveType = DeviceTypeCC
		} else {
			effectiveType = DeviceTypeAC // Default
		}
	}

	// Create device based on type
	switch effectiveType {
	case DeviceTypeCC:
		// Create Commercial Air Conditioner
		ccDevice := cc.NewCommercialAirConditioner(d.GetIP(), int(deviceID), d.GetPort())
		fmt.Println("ℹ️  商业空调设备 (CC) 支持有限，部分命令可能不可用")
		return device, ccDevice
	default:
		// Create Air Conditioner (default)
		acDevice := ac.NewAirConditioner(
			d.GetIP(),
			d.GetPort(),
			int(deviceID),
			msmart.WithName(name),
			msmart.WithVersion(version),
		)

		// Authenticate if V3
		if version == 3 && token != "" && key != "" {
			tokenBytes, err := hex.DecodeString(token)
			if err != nil {
				fmt.Printf("❌ 无效的Token: %v\n", err)
				os.Exit(1)
			}
			keyBytes, err := hex.DecodeString(key)
			if err != nil {
				fmt.Printf("❌ 无效的Key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("🔐 正在认证...")
			if err := acDevice.Authenticate(tokenBytes, keyBytes); err != nil {
				fmt.Printf("❌ 认证失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ 认证成功")
		}
		return device, acDevice
	}
}

func handleStatus(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea status <name|id> [--auto] [--capabilities] [--energy]")
		fmt.Println("   或者: midea status <host> --id <device-id> --token <token> --key <key>")
		os.Exit(1)
	}

	// Parse --auto, --capabilities and --energy flags
	autoMode := false
	showCapabilities := false
	showEnergy := false
	identifier := os.Args[2]
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
		if os.Args[i] == "--capabilities" {
			showCapabilities = true
		}
		if os.Args[i] == "--energy" {
			showEnergy = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Direct mode: if deviceID is provided, use direct connection
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		// Auto mode: discover device and get token/key automatically
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		// Normal mode: load from config
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device (currently only AC is fully supported)
	acDevice := mustGetACDevice(deviceObj)

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get capabilities first
	if err := acDevice.GetCapabilities(ctx); err != nil {
		fmt.Printf("⚠️  获取设备能力失败: %v\n", err)
	} else if showCapabilities {
		// Display capabilities if requested
		printCapabilities(acDevice)
	}

	// Enable energy usage requests if --energy flag is set
	if showEnergy {
		acDevice.SetEnableEnergyUsageRequests(true)
	}

	// Refresh state
	if err := acDevice.Refresh(ctx); err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		os.Exit(1)
	}

	printACState(acDevice)

	// Display energy usage if requested
	if showEnergy {
		printEnergyUsage(acDevice)
	}
}

func printCapabilities(acDevice *ac.AirConditioner) {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         📋 设备能力信息                ║")
	fmt.Println("╠════════════════════════════════════════╣")

	// Get capabilities dictionary
	caps := acDevice.CapabilitiesDict()
	if caps == nil || len(caps) == 0 {
		fmt.Println("║  ⚠️  无能力信息                        ║")
		fmt.Println("╚════════════════════════════════════════╝")
		return
	}

	// Display supported features
	if flags, ok := caps["supported_features"].([]string); ok && len(flags) > 0 {
		fmt.Println("║  支持的功能:                           ║")
		for _, flag := range flags {
			fmt.Printf("║    • %-32s║\n", flag)
		}
	}

	// Display supported modes
	if modes, ok := caps["supported_modes"].([]string); ok && len(modes) > 0 {
		fmt.Println("║  支持的模式:                           ║")
		fmt.Printf("║    %s                                ║\n", strings.Join(modes, ", "))
	}

	// Display supported fan speeds
	if fans, ok := caps["supported_fan_speeds"].([]string); ok && len(fans) > 0 {
		fmt.Println("║  支持的风速:                           ║")
		fmt.Printf("║    %s                                ║\n", strings.Join(fans, ", "))
	}

	// Display supported swing modes
	if swings, ok := caps["supported_swing_modes"].([]string); ok && len(swings) > 0 {
		fmt.Println("║  支持的摆风:                           ║")
		fmt.Printf("║    %s                              ║\n", strings.Join(swings, ", "))
	}

	// Display temperature range
	if minTemp, ok := caps["min_temperature"].(int); ok {
		if maxTemp, ok := caps["max_temperature"].(int); ok {
			fmt.Printf("║  温度范围: %d°C - %d°C                 ║\n", minTemp, maxTemp)
		}
	}

	fmt.Println("╚════════════════════════════════════════╝")
}

func printACState(acDevice *ac.AirConditioner) {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║           📊 空调状态                  ║")
	fmt.Println("╠════════════════════════════════════════╣")

	// Power state
	powerState := acDevice.PowerState()
	if powerState != nil && *powerState {
		fmt.Println("║  电源: 🟢 开启                         ║")
	} else {
		fmt.Println("║  电源: 🔴 关闭                         ║")
		fmt.Println("╚════════════════════════════════════════╝")
		return
	}

	// Temperature
	fmt.Printf("║  目标温度: %.0f°C                      ║\n", acDevice.TargetTemperature())
	if temp := acDevice.IndoorTemperature(); temp != nil {
		fmt.Printf("║  室内温度: %.1f°C                      ║\n", *temp)
	}
	if temp := acDevice.OutdoorTemperature(); temp != nil {
		fmt.Printf("║  室外温度: %.1f°C                      ║\n", *temp)
	}

	// Mode
	modeNames := map[ac.OperationalMode]string{
		ac.OperationalModeCool:    "❄️ 制冷",
		ac.OperationalModeHeat:    "🔥 制热",
		ac.OperationalModeAuto:    "🔄 自动",
		ac.OperationalModeDry:     "💧 除湿",
		ac.OperationalModeFanOnly: "🌀 送风",
	}
	modeName := modeNames[acDevice.OperationalMode()]
	fmt.Printf("║  运行模式: %-24s║\n", modeName)

	// Fan speed
	fanSpeed := acDevice.FanSpeed()
	var fanName string
	switch fs := fanSpeed.(type) {
	case ac.FanSpeed:
		fanNames := map[ac.FanSpeed]string{
			ac.FanSpeedAuto:   "自动",
			ac.FanSpeedLow:    "低",
			ac.FanSpeedMedium: "中",
			ac.FanSpeedHigh:   "高",
			ac.FanSpeedSilent: "静音",
		}
		fanName = fanNames[fs]
	default:
		fanName = fmt.Sprintf("%v", fs)
	}
	fmt.Printf("║  风速: %-30s║\n", fanName)

	// Swing mode
	swingNames := map[ac.SwingMode]string{
		ac.SwingModeOff:        "关闭",
		ac.SwingModeVertical:   "上下摆风",
		ac.SwingModeHorizontal: "左右摆风",
		ac.SwingModeBoth:       "全方位摆风",
	}
	fmt.Printf("║  摆风: %-30s║\n", swingNames[acDevice.SwingMode()])

	// Additional features
	if acDevice.Eco() {
		fmt.Println("║  🌿 ECO模式: 开启                      ║")
	}
	if acDevice.Turbo() {
		fmt.Println("║  🚀 强力模式: 开启                     ║")
	}

	fmt.Println("╚════════════════════════════════════════╝")
}

func printEnergyUsage(acDevice *ac.AirConditioner) {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║           ⚡ 能耗信息                  ║")
	fmt.Println("╠════════════════════════════════════════╣")

	hasEnergyData := false

	// Real-time power
	if power := acDevice.GetRealTimePowerUsage(ac.EnergyDataFormatBCD); power != nil {
		fmt.Printf("║  实时功率: %.1f W                      ║\n", *power)
		hasEnergyData = true
	}

	// Current energy usage (current month)
	if current := acDevice.GetCurrentEnergyUsage(ac.EnergyDataFormatBCD); current != nil {
		fmt.Printf("║  本月能耗: %.2f kWh                    ║\n", *current)
		hasEnergyData = true
	}

	// Total energy usage
	if total := acDevice.GetTotalEnergyUsage(ac.EnergyDataFormatBCD); total != nil {
		fmt.Printf("║  累计能耗: %.2f kWh                    ║\n", *total)
		hasEnergyData = true
	}

	if !hasEnergyData {
		fmt.Println("║  ⚠️  无能耗数据                        ║")
	}

	fmt.Println("╚════════════════════════════════════════╝")
}

func handlePower(configPath string, on bool, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 3 {
		action := "on"
		if !on {
			action = "off"
		}
		fmt.Printf("❌ 用法: midea %s <name|id> [--auto]\n", action)
		fmt.Println("   或者: midea %s <host> --id <device-id> --token <token> --key <key>")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Direct mode: if deviceID is provided, use direct connection
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set power state
	acDevice.SetPowerState(on)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	action := "已开机 ✅"
	if !on {
		action = "已关机 ⏹️"
	}
	fmt.Printf("✅ %s %s\n", device.Name, action)
}

func handleTemp(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea temp <name|id> <温度> [--auto]")
		fmt.Println("   或者: midea temp <host> <温度> --id <device-id> --token <token> --key <key>")
		fmt.Println("   温度范围: 16-30°C")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	tempArg := os.Args[3]
	for i := 4; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Direct mode: if deviceID is provided, use direct connection
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	temp, err := strconv.ParseFloat(tempArg, 64)
	if err != nil {
		fmt.Println("❌ 无效的温度值")
		fmt.Println("   温度范围: 16-30°C")
		os.Exit(1)
	}

	if temp < 16 || temp > 30 {
		fmt.Println("❌ 温度超出范围")
		fmt.Println("   温度范围: 16-30°C")
		os.Exit(1)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set temperature
	acDevice.SetTargetTemperature(temp)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s 温度已设置为 %.0f°C\n", device.Name, temp)
}

func handleMode(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea mode <name|id> <模式> [--auto]")
		fmt.Println("   模式: cool(制冷), heat(制热), auto(自动), dry(除湿), fan(送风)")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	modeArg := os.Args[3]
	for i := 4; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Use direct connection if deviceID, token and key are provided
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	// Map mode string to OperationalMode
	modeMap := map[string]ac.OperationalMode{
		"cool": ac.OperationalModeCool,
		"heat": ac.OperationalModeHeat,
		"auto": ac.OperationalModeAuto,
		"dry":  ac.OperationalModeDry,
		"fan":  ac.OperationalModeFanOnly,
	}

	mode, ok := modeMap[modeArg]
	if !ok {
		fmt.Printf("❌ 未知模式: %s\n", modeArg)
		fmt.Println("   模式: cool(制冷), heat(制热), auto(自动), dry(除湿), fan(送风)")
		os.Exit(1)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set mode
	acDevice.SetOperationalMode(mode)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	modeNames := map[ac.OperationalMode]string{
		ac.OperationalModeCool:    "制冷",
		ac.OperationalModeHeat:    "制热",
		ac.OperationalModeAuto:    "自动",
		ac.OperationalModeDry:     "除湿",
		ac.OperationalModeFanOnly: "送风",
	}
	fmt.Printf("✅ %s 模式已设置为 %s\n", device.Name, modeNames[mode])
}

func handleFan(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea fan <name|id> <风速> [--auto]")
		fmt.Println("   风速: auto(自动), low(低), medium(中), high(高), silent(静音)")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	speedStr := os.Args[3]
	for i := 4; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Use direct connection if deviceID, token and key are provided
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	// Map speed string to FanSpeed
	speedMap := map[string]ac.FanSpeed{
		"auto":   ac.FanSpeedAuto,
		"low":    ac.FanSpeedLow,
		"medium": ac.FanSpeedMedium,
		"high":   ac.FanSpeedHigh,
		"silent": ac.FanSpeedSilent,
	}

	speed, ok := speedMap[speedStr]
	if !ok {
		fmt.Printf("❌ 未知风速: %s\n", speedStr)
		fmt.Println("   风速: auto(自动), low(低), medium(中), high(高), silent(静音)")
		os.Exit(1)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set fan speed
	acDevice.SetFanSpeed(speed)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	speedNames := map[ac.FanSpeed]string{
		ac.FanSpeedAuto:   "自动",
		ac.FanSpeedLow:    "低",
		ac.FanSpeedMedium: "中",
		ac.FanSpeedHigh:   "高",
		ac.FanSpeedSilent: "静音",
	}
	fmt.Printf("✅ %s 风速已设置为 %s\n", device.Name, speedNames[speed])
}

func handleSwing(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea swing <name|id> <模式> [--auto]")
		fmt.Println("   模式: off(关闭), vertical(上下), horizontal(左右), both(全方位)")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	swingStr := os.Args[3]
	for i := 4; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Use direct connection if deviceID, token and key are provided
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	// Map swing string to SwingMode
	swingMap := map[string]ac.SwingMode{
		"off":        ac.SwingModeOff,
		"vertical":   ac.SwingModeVertical,
		"horizontal": ac.SwingModeHorizontal,
		"both":       ac.SwingModeBoth,
	}

	swingMode, ok := swingMap[swingStr]
	if !ok {
		fmt.Printf("❌ 未知摆风模式: %s\n", swingStr)
		fmt.Println("   模式: off(关闭), vertical(上下), horizontal(左右), both(全方位)")
		os.Exit(1)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set swing mode
	acDevice.SetSwingMode(swingMode)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	swingNames := map[ac.SwingMode]string{
		ac.SwingModeOff:        "关闭",
		ac.SwingModeVertical:   "上下摆风",
		ac.SwingModeHorizontal: "左右摆风",
		ac.SwingModeBoth:       "全方位摆风",
	}
	fmt.Printf("✅ %s 摆风已设置为 %s\n", device.Name, swingNames[swingMode])
}

// handleSet handles the set command for multi-parameter control
func handleSet(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea set <name|id> [选项] [--auto]")
		fmt.Println("   选项:")
		fmt.Println("     --temp <温度>      设置温度 (16-30°C)")
		fmt.Println("     --mode <模式>      设置模式 (cool/heat/auto/dry/fan)")
		fmt.Println("     --fan <风速>       设置风速 (auto/low/medium/high/silent)")
		fmt.Println("     --swing <模式>     设置摆风 (off/vertical/horizontal/both)")
		fmt.Println("     --power <on|off>   设置电源")
		fmt.Println("")
		fmt.Println("   示例:")
		fmt.Println("     midea set 客厅 --temp 26 --mode cool --fan high")
		fmt.Println("     midea set 客厅 --power on --temp 24")
		os.Exit(1)
	}

	// Parse --auto flag
	autoMode := false
	identifier := os.Args[2]
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--auto" || os.Args[i] == "-a" {
			autoMode = true
			break
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Use direct connection if deviceID, token and key are provided
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	// Parse flags
	var hasChanges bool
	var changes []string

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--temp":
			if i+1 >= len(os.Args) {
				fmt.Println("❌ --temp 需要指定温度")
				os.Exit(1)
			}
			temp, err := strconv.ParseFloat(os.Args[i+1], 64)
			if err != nil || temp < 16 || temp > 30 {
				fmt.Printf("❌ 无效的温度: %s (范围: 16-30°C)\n", os.Args[i+1])
				os.Exit(1)
			}
			acDevice.SetTargetTemperature(temp)
			changes = append(changes, fmt.Sprintf("温度 %.0f°C", temp))
			hasChanges = true
			i++

		case "--mode":
			if i+1 >= len(os.Args) {
				fmt.Println("❌ --mode 需要指定模式")
				os.Exit(1)
			}
			modeMap := map[string]ac.OperationalMode{
				"cool": ac.OperationalModeCool,
				"heat": ac.OperationalModeHeat,
				"auto": ac.OperationalModeAuto,
				"dry":  ac.OperationalModeDry,
				"fan":  ac.OperationalModeFanOnly,
			}
			mode, ok := modeMap[os.Args[i+1]]
			if !ok {
				fmt.Printf("❌ 无效的模式: %s\n", os.Args[i+1])
				os.Exit(1)
			}
			acDevice.SetOperationalMode(mode)
			modeNames := map[ac.OperationalMode]string{
				ac.OperationalModeCool:    "制冷",
				ac.OperationalModeHeat:    "制热",
				ac.OperationalModeAuto:    "自动",
				ac.OperationalModeDry:     "除湿",
				ac.OperationalModeFanOnly: "送风",
			}
			changes = append(changes, fmt.Sprintf("模式 %s", modeNames[mode]))
			hasChanges = true
			i++

		case "--fan":
			if i+1 >= len(os.Args) {
				fmt.Println("❌ --fan 需要指定风速")
				os.Exit(1)
			}
			speedMap := map[string]ac.FanSpeed{
				"auto":   ac.FanSpeedAuto,
				"low":    ac.FanSpeedLow,
				"medium": ac.FanSpeedMedium,
				"high":   ac.FanSpeedHigh,
				"silent": ac.FanSpeedSilent,
			}
			speed, ok := speedMap[os.Args[i+1]]
			if !ok {
				fmt.Printf("❌ 无效的风速: %s\n", os.Args[i+1])
				os.Exit(1)
			}
			acDevice.SetFanSpeed(speed)
			fanNames := map[ac.FanSpeed]string{
				ac.FanSpeedAuto:   "自动",
				ac.FanSpeedLow:    "低",
				ac.FanSpeedMedium: "中",
				ac.FanSpeedHigh:   "高",
				ac.FanSpeedSilent: "静音",
			}
			changes = append(changes, fmt.Sprintf("风速 %s", fanNames[speed]))
			hasChanges = true
			i++

		case "--swing":
			if i+1 >= len(os.Args) {
				fmt.Println("❌ --swing 需要指定摆风模式")
				os.Exit(1)
			}
			swingMap := map[string]ac.SwingMode{
				"off":        ac.SwingModeOff,
				"vertical":   ac.SwingModeVertical,
				"horizontal": ac.SwingModeHorizontal,
				"both":       ac.SwingModeBoth,
			}
			swing, ok := swingMap[os.Args[i+1]]
			if !ok {
				fmt.Printf("❌ 无效的摆风模式: %s\n", os.Args[i+1])
				os.Exit(1)
			}
			acDevice.SetSwingMode(swing)
			swingNames := map[ac.SwingMode]string{
				ac.SwingModeOff:        "关闭",
				ac.SwingModeVertical:   "上下摆风",
				ac.SwingModeHorizontal: "左右摆风",
				ac.SwingModeBoth:       "全方位摆风",
			}
			changes = append(changes, fmt.Sprintf("摆风 %s", swingNames[swing]))
			hasChanges = true
			i++

		case "--power":
			if i+1 >= len(os.Args) {
				fmt.Println("❌ --power 需要指定 on 或 off")
				os.Exit(1)
			}
			power := os.Args[i+1]
			if power != "on" && power != "off" {
				fmt.Printf("❌ 无效的电源状态: %s (应为 on 或 off)\n", power)
				os.Exit(1)
			}
			isOn := power == "on"
			acDevice.SetPowerState(isOn)
			if isOn {
				changes = append(changes, "开机")
			} else {
				changes = append(changes, "关机")
			}
			hasChanges = true
			i++
		}
	}

	if !hasChanges {
		fmt.Println("❌ 未指定任何更改")
		os.Exit(1)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s 已设置: %s\n", device.Name, strings.Join(changes, ", "))
}

// handleQuery handles the query command for querying device properties
func handleQuery(configPath string, deviceTypeStr string, deviceID int, deviceToken, deviceKey string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea query <name|id> [key] [--all] [--auto]")
		fmt.Println("   key: 属性名称 (如: temp, mode, fan, swing, power)")
		fmt.Println("   --all: 显示所有属性 (默认)")
		fmt.Println("   --auto: 自动发现设备并获取token")
		fmt.Println("")
		fmt.Println("   示例:")
		fmt.Println("     midea query 客厅              # 显示所有属性")
		fmt.Println("     midea query 客厅 temp         # 只显示温度")
		fmt.Println("     midea query 客厅 --auto       # 自动发现设备")
		os.Exit(1)
	}

	// Parse flags and arguments
	identifier := os.Args[2]
	var key string
	showAll := true
	autoMode := false

	// Parse arguments
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--all" {
			showAll = true
		} else if arg == "--auto" || arg == "-a" {
			autoMode = true
		} else if !strings.HasPrefix(arg, "--") {
			// This is the key argument
			key = arg
			showAll = false
		}
	}

	var device *config.Device
	var deviceObj interface{}

	// Use direct connection if deviceID, token and key are provided
	if deviceID > 0 {
		device, deviceObj = getDeviceDirect(identifier, deviceID, deviceToken, deviceKey, deviceTypeStr)
	} else if autoMode {
		device, deviceObj = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj = getDevice(configPath, identifier, deviceTypeStr)
	}

	// Get AC device
	acDevice := mustGetACDevice(deviceObj)

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Refresh state
	if err := acDevice.Refresh(ctx); err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if showAll || key == "" {
		printACState(acDevice)
	} else {
		printSpecificAttribute(acDevice, key)
	}
}

func printSpecificAttribute(acDevice *ac.AirConditioner, key string) {
	fmt.Println()

	switch strings.ToLower(key) {
	case "temp", "temperature", "target_temp":
		fmt.Printf("🌡️  目标温度: %.0f°C\n", acDevice.TargetTemperature())

	case "indoor_temp", "indoor_temperature":
		if temp := acDevice.IndoorTemperature(); temp != nil {
			fmt.Printf("🏠 室内温度: %.1f°C\n", *temp)
		} else {
			fmt.Println("⚠️  室内温度不可用")
		}

	case "outdoor_temp", "outdoor_temperature":
		if temp := acDevice.OutdoorTemperature(); temp != nil {
			fmt.Printf("🌤️  室外温度: %.1f°C\n", *temp)
		} else {
			fmt.Println("⚠️  室外温度不可用")
		}

	case "mode", "operational_mode":
		modeNames := map[ac.OperationalMode]string{
			ac.OperationalModeCool:    "制冷",
			ac.OperationalModeHeat:    "制热",
			ac.OperationalModeAuto:    "自动",
			ac.OperationalModeDry:     "除湿",
			ac.OperationalModeFanOnly: "送风",
		}
		fmt.Printf("🔄 运行模式: %s\n", modeNames[acDevice.OperationalMode()])

	case "fan", "fan_speed":
		fanSpeed := acDevice.FanSpeed()
		var fanName string
		switch fs := fanSpeed.(type) {
		case ac.FanSpeed:
			fanNames := map[ac.FanSpeed]string{
				ac.FanSpeedAuto:   "自动",
				ac.FanSpeedLow:    "低",
				ac.FanSpeedMedium: "中",
				ac.FanSpeedHigh:   "高",
				ac.FanSpeedSilent: "静音",
			}
			fanName = fanNames[fs]
		default:
			fanName = fmt.Sprintf("%v", fs)
		}
		fmt.Printf("🌀 风速: %s\n", fanName)

	case "swing", "swing_mode":
		swingNames := map[ac.SwingMode]string{
			ac.SwingModeOff:        "关闭",
			ac.SwingModeVertical:   "上下摆风",
			ac.SwingModeHorizontal: "左右摆风",
			ac.SwingModeBoth:       "全方位摆风",
		}
		fmt.Printf("🔀 摆风模式: %s\n", swingNames[acDevice.SwingMode()])

	case "power", "power_state":
		if powerState := acDevice.PowerState(); powerState != nil && *powerState {
			fmt.Println("⚡ 电源状态: 开启")
		} else {
			fmt.Println("⚡ 电源状态: 关闭")
		}

	case "eco":
		if acDevice.Eco() {
			fmt.Println("🌿 ECO模式: 开启")
		} else {
			fmt.Println("🌿 ECO模式: 关闭")
		}

	case "turbo":
		if acDevice.Turbo() {
			fmt.Println("🚀 强力模式: 开启")
		} else {
			fmt.Println("🚀 强力模式: 关闭")
		}

	default:
		fmt.Printf("❌ 未知属性: %s\n", key)
		fmt.Println("支持的属性:")
		fmt.Println("  temp, temperature, target_temp       - 目标温度")
		fmt.Println("  indoor_temp, indoor_temperature      - 室内温度")
		fmt.Println("  outdoor_temp, outdoor_temperature    - 室外温度")
		fmt.Println("  mode, operational_mode               - 运行模式")
		fmt.Println("  fan, fan_speed                       - 风速")
		fmt.Println("  swing, swing_mode                    - 摆风模式")
		fmt.Println("  power, power_state                   - 电源状态")
		fmt.Println("  eco                                  - ECO模式")
		fmt.Println("  turbo                                - 强力模式")
		os.Exit(1)
	}
}

// handleDownload handles the download command for downloading device protocol and plugin
func handleDownload(configPath string, region string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea download <host> [--account <账号> --password <密码>]")
		os.Exit(1)
	}

	host := os.Args[2]

	// Parse --account and --password flags
	var account, password string
	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--account":
			if i+1 < len(os.Args) {
				account = os.Args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(os.Args) {
				password = os.Args[i+1]
				i++
			}
		}
	}

	fmt.Printf("🔍 正在发现设备: %s\n", host)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Discover the device (no auto-connect, we just need SN)
	discoverConfig := &msmart.DiscoverConfig{
		Target:          host,
		Timeout:         5 * time.Second,
		DiscoveryPackets: 3,
		AutoConnect:      false, // Don't connect, just discover
	}

	devices, err := msmart.Discover(ctx, discoverConfig)
	if err != nil {
		fmt.Printf("❌ 发现设备失败: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("❌ 未找到设备")
		os.Exit(1)
	}

	// Get the first discovered device
	d := devices[0]

	// Get device info
	deviceType := d.GetType()
	sn := d.GetSN()

	if sn == nil || *sn == "" {
		fmt.Println("❌ 设备没有 SN,无法下载协议")
		os.Exit(1)
	}

	fmt.Printf("✅ 发现设备: 类型=%02X, SN=%s\n", deviceType, *sn)

	// Create cloud client
	fmt.Println("☁️  正在连接云端...")

	var cloud *msmart.SmartHomeCloud
	var accountPtr, passwordPtr *string

	if account != "" && password != "" {
		accountPtr = &account
		passwordPtr = &password
	}

	cloud, err = msmart.NewSmartHomeCloud(region, accountPtr, passwordPtr, false, nil)
	if err != nil {
		fmt.Printf("❌ 创建云端客户端失败: %v\n", err)
		os.Exit(1)
	}

	// Login to cloud
	if err := cloud.Login(false); err != nil {
		fmt.Printf("❌ 云端登录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 云端登录成功")

	// Download Lua protocol
	fmt.Println("📥 正在下载 Lua 协议...")
	luaName, luaContent, err := cloud.GetProtocolLua(deviceType, *sn)
	if err != nil {
		fmt.Printf("❌ 下载 Lua 协议失败: %v\n", err)
		os.Exit(1)
	}

	// Save Lua file
	if err := os.WriteFile(luaName, []byte(luaContent), 0644); err != nil {
		fmt.Printf("❌ 保存 Lua 文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Lua 协议已保存: %s\n", luaName)

	// Download plugin
	fmt.Println("📥 正在下载插件...")
	pluginName, pluginData, err := cloud.GetPlugin(deviceType, *sn)
	if err != nil {
		fmt.Printf("❌ 下载插件失败: %v\n", err)
		os.Exit(1)
	}

	// Save plugin file
	if err := os.WriteFile(pluginName, pluginData, 0644); err != nil {
		fmt.Printf("❌ 保存插件文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 插件已保存: %s\n", pluginName)

	fmt.Println("\n✅ 下载完成!")
}
