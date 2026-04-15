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
)

var version = "1.0.0"

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
		handleDiscover(configPath)
	case "list":
		handleList(configPath)
	case "bind":
		handleBind(configPath)
	case "unbind":
		handleUnbind(configPath)
	case "status":
		handleStatus(configPath)
	case "on":
		handlePower(configPath, true)
	case "off":
		handlePower(configPath, false)
	case "temp":
		handleTemp(configPath)
	case "mode":
		handleMode(configPath)
	case "fan":
		handleFan(configPath)
	case "swing":
		handleSwing(configPath)
	case "set":
		handleSet(configPath)
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
  midea [-v|--verbose] <command> [arguments]

全局选项:
  -v, --verbose    显示详细调试日志

命令:
  discover [--auto-connect|-a] [--account <账号> --password <密码>]
                                发现设备并保存到配置
                                --auto-connect: 自动连接并获取V3设备的token
                                --account/--password: 美的账号密码 (V3设备认证需要)
  list                          列出已保存的设备
  bind <id|sn|ip> -n <名称>   绑定设备别名
  unbind <name|id>            解绑设备
  
  status <name|id>            查询设备状态
  on <name|id>                开机
  off <name|id>               关机
  temp <name|id> <温度>       设置温度 (范围: 16-30°C)
  mode <name|id> <模式>       设置运行模式
  fan <name|id> <风速>        设置风速
  swing <name|id> <模式>      设置摆风模式
  set <name|id> [选项]        多参数设置 (一次设置多个属性)

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
  midea bind 192.168.1.60 -n 客厅    # 绑定IP为192.168.1.60的设备，命名为"客厅"
  midea status 客厅                   # 查询"客厅"空调状态
  midea on 客厅                       # 打开"客厅"空调
  midea temp 客厅 26                  # 设置温度为26°C
  midea mode 客厅 cool                # 设置为制冷模式
  midea fan 客厅 high                 # 设置为高风速
  midea swing 客厅 vertical           # 设置为上下摆风

  # 多参数设置 (一次命令设置多个属性)
  midea set 客厅 --temp 26 --mode cool --fan high
  midea set 客厅 --power on --temp 24

配置文件: 当前目录的 midea.json (优先) 或 ~/.config/midea/config.json
`)
}

// ============================================================================
// Discovery Commands
// ============================================================================

func handleDiscover(configPath string) {
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

		// For V3 devices without tokens, show a warning
		if version == 3 && token == "" {
			fmt.Printf("  ⚠️  V3设备需要云端认证。使用 'midea discover --auto-connect' 获取token。\n")
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

func getDevice(configPath, identifier string) (*config.Device, *ac.AirConditioner) {
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

	// Parse device ID
	deviceID, err := strconv.ParseInt(device.ID, 10, 64)
	if err != nil {
		fmt.Printf("❌ 无效的设备ID: %s\n", device.ID)
		os.Exit(1)
	}

	// Create AC device
	acDevice := ac.NewAirConditioner(
		device.IP,
		device.Port,
		int(deviceID),
		msmart.WithName(device.Name),
		msmart.WithVersion(device.Version),
	)

	// Set token and key if available
	// Note: Only authenticate if version is explicitly 3
	// Version 0 or 2 devices use V2 protocol which doesn't need token/key authentication
	if device.Version == 3 {
		if device.Token == "" || device.Key == "" {
			fmt.Println("❌ V3设备需要token和key进行认证")
			fmt.Println("💡 请使用 'midea status <ip> --auto' 自动获取token")
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

// getDeviceAuto automatically discovers a device and gets token/key for V3 devices
func getDeviceAuto(identifier string, configPath string) (*config.Device, *ac.AirConditioner) {
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
		fmt.Println("❌ V3设备需要token和key进行认证，但自动获取失败")
		fmt.Println("💡 请尝试使用 'midea discover --auto-connect' 命令")
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

	// Create AC device
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

func handleStatus(configPath string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea status <name|id> [--auto]")
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
	var acDevice *ac.AirConditioner

	if autoMode {
		// Auto mode: discover device and get token/key automatically
		device, acDevice = getDeviceAuto(identifier, configPath)
	} else {
		// Normal mode: load from config
		device, acDevice = getDevice(configPath, identifier)
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get capabilities first
	if err := acDevice.GetCapabilities(ctx); err != nil {
		fmt.Printf("⚠️  获取设备能力失败: %v\n", err)
	}

	// Refresh state
	if err := acDevice.Refresh(ctx); err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		os.Exit(1)
	}

	printACState(acDevice)
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

func handlePower(configPath string, on bool) {
	if len(os.Args) < 3 {
		action := "on"
		if !on {
			action = "off"
		}
		fmt.Printf("❌ 用法: midea %s <name|id>\n", action)
		os.Exit(1)
	}

	device, acDevice := getDevice(configPath, os.Args[2])

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

func handleTemp(configPath string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea temp <name|id> <温度>")
		fmt.Println("   温度范围: 16-30°C")
		os.Exit(1)
	}

	device, acDevice := getDevice(configPath, os.Args[2])

	temp, err := strconv.ParseFloat(os.Args[3], 64)
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

func handleMode(configPath string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea mode <name|id> <模式>")
		fmt.Println("   模式: cool(制冷), heat(制热), auto(自动), dry(除湿), fan(送风)")
		os.Exit(1)
	}

	device, acDevice := getDevice(configPath, os.Args[2])
	modeStr := os.Args[3]

	// Map mode string to OperationalMode
	modeMap := map[string]ac.OperationalMode{
		"cool": ac.OperationalModeCool,
		"heat": ac.OperationalModeHeat,
		"auto": ac.OperationalModeAuto,
		"dry":  ac.OperationalModeDry,
		"fan":  ac.OperationalModeFanOnly,
	}

	mode, ok := modeMap[modeStr]
	if !ok {
		fmt.Printf("❌ 未知模式: %s\n", modeStr)
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

func handleFan(configPath string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea fan <name|id> <风速>")
		fmt.Println("   风速: auto(自动), low(低), medium(中), high(高), silent(静音)")
		os.Exit(1)
	}

	device, acDevice := getDevice(configPath, os.Args[2])
	speedStr := os.Args[3]

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

func handleSwing(configPath string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 用法: midea swing <name|id> <模式>")
		fmt.Println("   模式: off(关闭), vertical(上下), horizontal(左右), both(全方位)")
		os.Exit(1)
	}

	device, acDevice := getDevice(configPath, os.Args[2])
	swingStr := os.Args[3]

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
func handleSet(configPath string) {
	if len(os.Args) < 3 {
		fmt.Println("❌ 用法: midea set <name|id> [选项]")
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

	device, acDevice := getDevice(configPath, os.Args[2])

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
