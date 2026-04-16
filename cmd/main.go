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
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
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

// GlobalFlags 全局命令行参数结构体
type GlobalFlags struct {
	// Global flags
	Region      string
	DeviceType  string
	DeviceID    int
	DeviceToken string
	DeviceKey   string
	Verbose     bool

	// Subcommand flags (shared across commands)
	AutoMode         bool
	Account          string
	Password         string
	AutoConnect      bool
	DiscoveryCount   int
	ShowCapabilities bool
	CapabilitiesFile string
	ShowEnergy       bool
	ShowAll          bool
	TempValue        string
	ModeValue        string
	FanValue         string
	SwingValue       string
	PowerValue       string
	Name             string
}

// setupFlags 设置命令行参数
func setupFlags(flags *GlobalFlags) {
	// Global flags
	pflag.StringVarP(&flags.Region, "region", "r", msmart.DefaultCloudRegion, "Cloud region (CN, US, EU)")
	pflag.StringVarP(&flags.DeviceType, "device_type", "d", "AC", "Device type (AC, CC)")
	pflag.IntVarP(&flags.DeviceID, "id", "i", 0, "Device ID for V3 devices")
	pflag.StringVarP(&flags.DeviceToken, "token", "T", "", "Auth token for V3 devices")
	pflag.StringVarP(&flags.DeviceKey, "key", "k", "", "Auth key for V3 devices")
	pflag.BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")

	// Subcommand flags
	pflag.BoolVarP(&flags.AutoMode, "auto", "a", false, "Auto discover device")
	pflag.StringVar(&flags.Account, "account", "", "Midea account")
	pflag.StringVar(&flags.Password, "password", "", "Midea password")
	pflag.BoolVarP(&flags.AutoConnect, "auto-connect", "c", false, "Auto connect and get token")
	pflag.IntVar(&flags.DiscoveryCount, "count", 3, "Discovery packet count")
	pflag.BoolVarP(&flags.ShowEnergy, "energy", "e", false, "Show energy info")
	pflag.BoolVar(&flags.ShowAll, "all", false, "Show all properties")
	pflag.StringVarP(&flags.TempValue, "temp", "t", "", "Temperature (16-30)")
	pflag.StringVarP(&flags.ModeValue, "mode", "m", "", "Mode (cool/heat/auto/dry/fan)")
	pflag.StringVarP(&flags.FanValue, "fan", "f", "", "Fan speed (auto/low/medium/high/silent)")
	pflag.StringVarP(&flags.SwingValue, "swing", "s", "", "Swing mode (off/vertical/horizontal/both)")
	pflag.StringVarP(&flags.PowerValue, "power", "p", "", "Power state (on/off)")
	pflag.StringVarP(&flags.Name, "name", "n", "", "Device name")

	pflag.Usage = printUsage
}

func main() {
	flags := &GlobalFlags{}
	setupFlags(flags)
	if err := run(flags); err != nil {
		// Print error if not already printed
		fmt.Printf("❌ 错误: %v\n", err)
		os.Exit(1)
	}
}

// parseCommand 解析命令行参数，返回获取命令参数的函数
// 使用 lambda 表达式封装参数访问
func parseCommand() func(int) string {
	pflag.Parse()
	args := pflag.Args()
	return func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
}

func run(flags *GlobalFlags) error {
	// 一次性解析所有参数，返回 lambda 函数
	getCommand := parseCommand()

	if flags.Verbose {
		msmart.Verbose = true
	}

	// Validate device type
	deviceTypeStr := strings.ToUpper(flags.DeviceType)
	if _, ok := deviceTypeMap[deviceTypeStr]; !ok {
		fmt.Printf("❌ 不支持的设备类型: %s\n", flags.DeviceType)
		fmt.Println("   支持的设备类型: AC (空调), CC (商业空调)")
		return fmt.Errorf("unsupported device type: %s", flags.DeviceType)
	}

	// 获取命令
	command := getCommand(0)
	if command == "" {
		printUsage()
		return fmt.Errorf("no command provided")
	}

	configPath := config.DefaultConfigPath()

	// Commands that don't need config
	switch command {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "--version":
		fmt.Printf("midea %s\n", version)
		return nil
	}

	// Execute command
	switch command {
	case "discover":
		return handleDiscover(configPath, flags, getCommand(1))
	case "list":
		return handleList(configPath)
	case "bind":
		identifier := getCommand(1)
		if identifier == "" {
			return fmt.Errorf("bind requires identifier")
		}
		return handleBind(configPath, identifier, flags.Name)
	case "unbind":
		identifier := getCommand(1)
		if identifier == "" {
			return fmt.Errorf("unbind requires identifier")
		}
		return handleUnbind(configPath, identifier)
	case "status":
		return handleStatus(configPath, flags, deviceTypeStr, getCommand(1))
	case "on":
		return handlePower(configPath, flags, deviceTypeStr, true, getCommand(1))
	case "off":
		return handlePower(configPath, flags, deviceTypeStr, false, getCommand(1))
	case "temp":
		identifier := getCommand(1)
		tempStr := getCommand(2)
		if identifier == "" || tempStr == "" {
			return fmt.Errorf("temp requires identifier and temperature")
		}
		temp, err := ParseTemp(tempStr)
		if err != nil {
			return fmt.Errorf("invalid temperature: %s", tempStr)
		}
		return handleTemp(configPath, flags, deviceTypeStr, identifier, temp)
	case "mode":
		identifier := getCommand(1)
		modeStr := getCommand(2)
		if identifier == "" || modeStr == "" {
			return fmt.Errorf("mode requires identifier and mode")
		}
		mode, err := ParseMode(modeStr)
		if err != nil {
			return fmt.Errorf("invalid mode: %s", modeStr)
		}
		return handleMode(configPath, flags, deviceTypeStr, identifier, mode)
	case "fan":
		identifier := getCommand(1)
		fanStr := getCommand(2)
		if identifier == "" || fanStr == "" {
			return fmt.Errorf("fan requires identifier and fan speed")
		}
		speed, err := ParseFanSpeed(fanStr)
		if err != nil {
			return fmt.Errorf("invalid fan speed: %s", fanStr)
		}
		return handleFan(configPath, flags, deviceTypeStr, identifier, speed)
	case "swing":
		identifier := getCommand(1)
		swingStr := getCommand(2)
		if identifier == "" || swingStr == "" {
			return fmt.Errorf("swing requires identifier and swing mode")
		}
		swing, err := ParseSwingMode(swingStr)
		if err != nil {
			return fmt.Errorf("invalid swing mode: %s", swingStr)
		}
		return handleSwing(configPath, flags, deviceTypeStr, identifier, swing)
	case "set":
		return handleSet(configPath, flags, deviceTypeStr, getCommand(1))
	case "query":
		return handleQuery(configPath, flags, deviceTypeStr, getCommand(1), getCommand(2))
	case "download":
		return handleDownload(configPath, flags, getCommand(1))
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
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
  discover [<host>] [--auto-connect|-a] [--count <数量>] [--account <账号> --password <密码>]
                                发现设备并保存到配置
                                <host>: 可选,指定目标设备IP (发现单个设备)
                                --auto-connect: 自动连接并获取V3设备的token
                                --count: 广播包数量 (默认: 3)
                                --account/--password: 美的账号密码 (V3设备认证需要)
  list                          列出已保存的设备
  bind <id|sn|ip> -n <名称>   绑定设备别名
  unbind <name|id>            解绑设备

  status <name|id> [--auto] [--capabilities [FILE]] [--energy]
                                查询设备状态
                                --auto: 自动发现设备并获取token
                                --capabilities: 显示设备能力信息
                                --capabilities FILE: 将设备能力写入YAML文件
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
  midea discover                      # 发现局域网内的所有设备
  midea discover 192.168.1.60         # 发现指定IP的设备
  midea -v discover --auto-connect   # 使用verbose模式发现设备并自动获取V3设备token
  midea discover --auto-connect --account your@email.com --password yourpass
                                      # 使用自定义账号发现设备
  midea list                          # 列出已保存的设备
  midea bind 192.168.1.60 -n 客厅    # 绑定IP为192.168.1.60的设备,命名为"客厅"
  midea status 客厅                   # 查询"客厅"空调状态
  midea status 客厅 --capabilities    # 查询"客厅"空调状态并显示设备能力
  midea status 客厅 --capabilities caps.yaml  # 将设备能力写入 caps.yaml 文件
  midea status 客厅 --energy          # 查询"客厅"空调状态并显示能耗信息
  midea on 客厅                       # 打开"客厅"空调
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

func handleDiscover(configPath string, flags *GlobalFlags, targetHost string) error {
	fmt.Println("🔍 正在发现设备...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return err
	}

	// Discover devices
	discoverConfig := &msmart.DiscoverConfig{
		Timeout:          5 * time.Second,
		DiscoveryPackets: flags.DiscoveryCount,
		AutoConnect:      flags.AutoConnect,
		Region:           flags.Region,
	}

	// Set target host if provided (for single device discovery)
	if targetHost != "" {
		discoverConfig.Target = targetHost
		fmt.Printf("🎯 目标设备: %s\n", targetHost)
	}

	// Set account and password if provided
	if flags.Account != "" && flags.Password != "" {
		discoverConfig.Account = flags.Account
		discoverConfig.Password = flags.Password
	}

	devices, err := msmart.Discover(ctx, discoverConfig)

	// Even if there's an error, check if we discovered any devices
	if err != nil && len(devices) == 0 {
		fmt.Printf("❌ 发现设备失败: %v\n", err)
		return err
	}

	// Log warning if there was an error but we have devices
	if err != nil && len(devices) > 0 {
		fmt.Printf("⚠️  发现设备时有错误: %v\n", err)
	}

	if len(devices) == 0 {
		fmt.Println("⚠️  未发现任何设备")
		return nil
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
		return err
	}

	fmt.Printf("\n💾 配置已保存到: %s\n", configPath)
	fmt.Println("\n💡 使用 'midea bind <id|ip> -n <名称>' 来为设备命名")
	return nil
}

func handleList(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return err
	}

	devices := cfg.ListDevices()
	if len(devices) == 0 {
		fmt.Println("📋 暂无设备配置")
		fmt.Println("💡 使用 'midea discover' 来发现设备")
		return nil
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
	return nil
}

// ============================================================================
// Device Management Commands
// ============================================================================

func handleBind(configPath string, identifier string, name string) error {
	if name == "" {
		fmt.Println("❌ 请指定名称: midea bind <id|sn|ip> -n <名称>")
		return fmt.Errorf("name not specified for bind command")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return err
	}

	if !cfg.BindName(identifier, name) {
		fmt.Printf("❌ 未找到设备: %s\n", identifier)
		fmt.Println("💡 使用 'midea list' 查看设备列表")
		return fmt.Errorf("device not found: %s", identifier)
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("❌ 保存配置失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已绑定: %s -> %s\n", identifier, name)
	return nil
}

func handleUnbind(configPath string, identifier string) error {
	if identifier == "" {
		fmt.Println("❌ 用法: midea unbind <name|id>")
		return fmt.Errorf("insufficient arguments for unbind command")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return err
	}

	if !cfg.RemoveDevice(identifier) {
		fmt.Printf("❌ 未找到设备: %s\n", identifier)
		fmt.Println("💡 使用 'midea list' 查看设备列表")
		return fmt.Errorf("device not found: %s", identifier)
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Printf("❌ 保存配置失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 已解绑: %s\n", identifier)
	return nil
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

// mustGetACDevice extracts an AC device from interface{}, returns error if not AC type
func mustGetACDevice(device interface{}) (*ac.AirConditioner, error) {
	acDevice, ok := device.(*ac.AirConditioner)
	if !ok {
		return nil, fmt.Errorf("此命令只支持空调设备 (AC)，请使用 --device_type AC 指定空调设备")
	}
	return acDevice, nil
}

func getDevice(configPath, identifier string, deviceTypeStr string) (*config.Device, interface{}, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
	}

	device := cfg.GetDevice(identifier)
	if device == nil {
		return nil, nil, fmt.Errorf("未找到设备: %s (使用 'midea list' 查看设备列表 或使用 --auto 自动发现设备)", identifier)
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
		return nil, nil, fmt.Errorf("无效的设备ID: %s", device.ID)
	}

	// Create device based on type
	switch effectiveType {
	case DeviceTypeCC:
		// Create Commercial Air Conditioner
		ccDevice := cc.NewCommercialAirConditioner(device.IP, int(deviceID), device.Port)
		fmt.Println("i️  商业空调设备 (CC) 支持有限,部分命令可能不可用")
		return device, ccDevice, nil
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
				return nil, nil, fmt.Errorf("V3设备需要token和key进行认证 (使用 '--auto' 参数自动获取token/key,或使用 'midea discover --auto-connect' 重新发现设备)")
			}

			token, err := hex.DecodeString(device.Token)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的Token: %w", err)
			}
			key, err := hex.DecodeString(device.Key)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的Key: %w", err)
			}
			
			// Check if already authenticated (reuse cached connection)
			if acDevice.IsAuthenticated() {
				fmt.Println("✅ 已认证，复用现有连接")
			} else {
				fmt.Println("🔐 正在认证...")
				if err := acDevice.Authenticate(token, key); err != nil {
					return nil, nil, fmt.Errorf("认证失败: %w", err)
				}
				fmt.Println("✅ 认证成功")
			}
		}
		return device, acDevice, nil
	}
}

// getDeviceDirect creates a device directly with host, id, token and key
// This is used when --id, --token, --key are provided (similar to Python CLI)
func getDeviceDirect(host string, deviceID int, tokenStr, keyStr string, deviceTypeStr string) (*config.Device, interface{}, error) {
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
		fmt.Println("i️  商业空调设备 (CC) 支持有限,部分命令可能不可用")
		return device, ccDevice, nil
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
				return nil, nil, fmt.Errorf("无效的Token: %w", err)
			}
			key, err := hex.DecodeString(keyStr)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的Key: %w", err)
			}
			
			// Check if already authenticated (reuse cached connection)
			if acDevice.IsAuthenticated() {
				fmt.Println("✅ 已认证，复用现有连接")
			} else {
				fmt.Println("🔐 正在认证...")
				if err := acDevice.Authenticate(token, key); err != nil {
					return nil, nil, fmt.Errorf("认证失败: %w", err)
				}
				fmt.Println("✅ 认证成功")
			}
		}

		return device, acDevice, nil
	}
}

// getDeviceAuto automatically discovers a device and gets token/key for V3 devices
func getDeviceAuto(identifier string, configPath string, deviceTypeStr string) (*config.Device, interface{}, error) {
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
		Region:          msmart.DefaultCloudRegion, // Use default region for default credentials
	}

	devices, err := msmart.Discover(ctx, discoverConfig)
	if err != nil {
		return nil, nil, err
	}

	if len(devices) == 0 {
		return nil, nil, fmt.Errorf("未找到设备")
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
		return nil, nil, fmt.Errorf("V3设备未能获取token/key")
	}

	// Load config and save device
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
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
		fmt.Println("i️  商业空调设备 (CC) 支持有限,部分命令可能不可用")
		return device, ccDevice, nil
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
				return nil, nil, fmt.Errorf("无效的Token: %w", err)
			}
			keyBytes, err := hex.DecodeString(key)
			if err != nil {
				return nil, nil, fmt.Errorf("无效的Key: %w", err)
			}
			
			// Check if already authenticated (reuse cached connection)
			if acDevice.IsAuthenticated() {
				fmt.Println("✅ 已认证，复用现有连接")
			} else {
				fmt.Println("🔐 正在认证...")
				if err := acDevice.Authenticate(tokenBytes, keyBytes); err != nil {
					return nil, nil, fmt.Errorf("认证失败: %w", err)
				}
				fmt.Println("✅ 认证成功")
			}
		}
		return device, acDevice, nil
	}
}

func handleStatus(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Direct mode: if deviceID is provided, use direct connection
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		// Auto mode: discover device and get token/key automatically
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		// Normal mode: load from config
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device (currently only AC is fully supported)
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get capabilities first
	if err := acDevice.GetCapabilities(ctx); err != nil {
		fmt.Printf("⚠️  获取设备能力失败: %v\n", err)
	} else if flags.ShowCapabilities {
		// If a file path is specified, write to file
		if flags.CapabilitiesFile != "" {
			if err := writeCapabilitiesToYAML(acDevice, flags.CapabilitiesFile); err != nil {
				fmt.Printf("❌ 写入能力信息到文件失败: %v\n", err)
			} else {
				fmt.Printf("✅ 设备能力已写入: %s\n", flags.CapabilitiesFile)
			}
		} else {
			// Display capabilities to screen
			printCapabilities(acDevice)
		}
	}

	// Enable energy usage requests if --energy flag is set
	if flags.ShowEnergy {
		acDevice.SetEnableEnergyUsageRequests(true)
	}

	// Refresh state
	if err := acDevice.Refresh(ctx); err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return err
	}

	printACState(acDevice)

	// Display energy usage if requested
	if flags.ShowEnergy {
		printEnergyUsage(acDevice)
	}
	return nil
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

// writeCapabilitiesToYAML writes device capabilities to a YAML file
func writeCapabilitiesToYAML(acDevice *ac.AirConditioner, filename string) error {
	// Get capabilities dictionary
	caps := acDevice.CapabilitiesDict()
	if caps == nil || len(caps) == 0 {
		return fmt.Errorf("无能力信息")
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(caps)
	if err != nil {
		return fmt.Errorf("转换 YAML 失败: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
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
		fanName = SpeedNames[fs]
	default:
		fanName = fmt.Sprintf("%v", fs)
	}
	fmt.Printf("║  风速: %-30s║\n", fanName)

	// Swing mode
	fmt.Printf("║  摆风: %-30s║\n", SwingNames[acDevice.SwingMode()])

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

func handlePower(configPath string, flags *GlobalFlags, deviceTypeStr string, on bool, identifier string) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Direct mode: if deviceID is provided, use direct connection
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
	}

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
		return err
	}

	action := "已开机 ✅"
	if !on {
		action = "已关机 ⏹️"
	}
	fmt.Printf("✅ %s %s\n", device.Name, action)
	return nil
}

func handleTemp(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, temp float64) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Direct mode: if deviceID is provided, use direct connection
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
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
		return err
	}

	fmt.Printf("✅ %s 温度已设置为 %.0f°C\n", device.Name, temp)
	return nil
}

func handleMode(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, mode ac.OperationalMode) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Use direct connection if deviceID, token and key are provided
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
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
		return err
	}

	fmt.Printf("✅ %s 模式已设置为 %s\n", device.Name, ModeNames[mode])
	return nil
}

func handleFan(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, speed ac.FanSpeed) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Use direct connection if deviceID, token and key are provided
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
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
		return err
	}

	fmt.Printf("✅ %s 风速已设置为 %s\n", device.Name, SpeedNames[speed])
	return nil
}

func handleSwing(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, swing ac.SwingMode) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Use direct connection if deviceID, token and key are provided
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set swing mode
	acDevice.SetSwingMode(swing)

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ %s 摆风已设置为 %s\n", device.Name, SwingNames[swing])
	return nil
}

// handleSet handles the set command for multi-parameter control
func handleSet(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Use direct connection if deviceID, token and key are provided
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
	}

	// Track changes
	var hasChanges bool
	var changes []string

	// Apply temperature if specified
	if flags.TempValue != "" {
		temp, err := ParseTemp(flags.TempValue)
		if err != nil {
			return fmt.Errorf("invalid temperature: %s", flags.TempValue)
		}
		acDevice.SetTargetTemperature(temp)
		changes = append(changes, fmt.Sprintf("温度 %.0f°C", temp))
		hasChanges = true
	}

	// Apply mode if specified
	if flags.ModeValue != "" {
		mode, err := ParseMode(flags.ModeValue)
		if err != nil {
			return fmt.Errorf("invalid mode: %s", flags.ModeValue)
		}
		acDevice.SetOperationalMode(mode)
		changes = append(changes, fmt.Sprintf("模式 %s", ModeNames[mode]))
		hasChanges = true
	}

	// Apply fan speed if specified
	if flags.FanValue != "" {
		fanSpeed, err := ParseFanSpeed(flags.FanValue)
		if err != nil {
			return fmt.Errorf("invalid fan speed: %s", flags.FanValue)
		}
		acDevice.SetFanSpeed(fanSpeed)
		changes = append(changes, fmt.Sprintf("风速 %s", SpeedNames[fanSpeed]))
		hasChanges = true
	}

	// Apply swing mode if specified
	if flags.SwingValue != "" {
		swing, err := ParseSwingMode(flags.SwingValue)
		if err != nil {
			return fmt.Errorf("invalid swing mode: %s", flags.SwingValue)
		}
		acDevice.SetSwingMode(swing)
		changes = append(changes, fmt.Sprintf("摆风 %s", SwingNames[swing]))
		hasChanges = true
	}

	// Apply power state if specified
	if flags.PowerValue != "" {
		var power bool
		switch strings.ToLower(flags.PowerValue) {
		case "on", "true", "1":
			power = true
		case "off", "false", "0":
			power = false
		default:
			return fmt.Errorf("invalid power state: %s", flags.PowerValue)
		}
		acDevice.SetPowerState(power)
		if power {
			changes = append(changes, "开机")
		} else {
			changes = append(changes, "关机")
		}
		hasChanges = true
	}

	if !hasChanges {
		fmt.Println("❌ 未指定任何更改")
		return fmt.Errorf("no changes specified")
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		fmt.Printf("❌ 控制失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ %s 已设置: %s\n", device.Name, strings.Join(changes, ", "))
	return nil
}
func handleQuery(configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, key string) error {
	var device *config.Device
	var deviceObj interface{}
	var err error

	// Use direct connection if deviceID, token and key are provided
	if flags.DeviceID > 0 {
		device, deviceObj, err = getDeviceDirect(identifier, flags.DeviceID, flags.DeviceToken, flags.DeviceKey, deviceTypeStr)
	} else if flags.AutoMode {
		device, deviceObj, err = getDeviceAuto(identifier, configPath, deviceTypeStr)
	} else {
		device, deviceObj, err = getDevice(configPath, identifier, deviceTypeStr)
	}
	if err != nil {
		return err
	}

	// Get AC device
	acDevice, err := mustGetACDevice(deviceObj)
	if err != nil {
		fmt.Println("❌ " + err.Error())
		return err
	}

	fmt.Printf("\n🎯 目标设备: %s (%s)\n", device.Name, device.IP)
	fmt.Println("🔌 正在连接...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Refresh state
	if err := acDevice.Refresh(ctx); err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return err
	}

	// Display results
	if flags.ShowAll || key == "" {
		printACState(acDevice)
	} else {
		if err := printSpecificAttribute(acDevice, key); err != nil {
			return err
		}
	}
	return nil
}


func printSpecificAttribute(acDevice *ac.AirConditioner, key string) error {
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
		fmt.Printf("🔄 运行模式: %s\n", ModeNames[acDevice.OperationalMode()])

	case "fan", "fan_speed":
		fanSpeed := acDevice.FanSpeed()
		var fanName string
		switch fs := fanSpeed.(type) {
		case ac.FanSpeed:
			fanName = SpeedNames[fs]
		default:
			fanName = fmt.Sprintf("%v", fs)
		}
		fmt.Printf("🌀 风速: %s\n", fanName)

	case "swing", "swing_mode":
		fmt.Printf("🔀 摆风模式: %s\n", SwingNames[acDevice.SwingMode()])

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
		return fmt.Errorf("unknown attribute: %s", key)
	}
	return nil
}

// handleDownload handles the download command for downloading device protocol and plugin
func handleDownload(configPath string, flags *GlobalFlags, host string) error {
	if host == "" {
		fmt.Println("❌ 用法: midea download <host> [--account <账号> --password <密码>]")
		return fmt.Errorf("insufficient arguments for download command")
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
		return err
	}

	if len(devices) == 0 {
		return fmt.Errorf("未找到设备")
	}

	// Get the first discovered device
	d := devices[0]

	// Get device info
	deviceType := d.GetType()
	sn := d.GetSN()

	if sn == nil || *sn == "" {
		return fmt.Errorf("设备没有 SN,无法下载协议")
	}

	fmt.Printf("✅ 发现设备: 类型=%02X, SN=%s\n", deviceType, *sn)

	// Create cloud client
	fmt.Println("☁️  正在连接云端...")

	var cloud *msmart.SmartHomeCloud
	var accountPtr, passwordPtr *string

	if flags.Account != "" && flags.Password != "" {
		accountPtr = &flags.Account
		passwordPtr = &flags.Password
	}

	cloud, err = msmart.NewSmartHomeCloud(flags.Region, accountPtr, passwordPtr, false, nil)
	if err != nil {
		fmt.Printf("❌ 创建云端客户端失败: %v\n", err)
		return err
	}

	// Login to cloud
	if err := cloud.Login(false); err != nil {
		fmt.Printf("❌ 云端登录失败: %v\n", err)
		return err
	}

	fmt.Println("✅ 云端登录成功")

	// Download Lua protocol
	fmt.Println("📥 正在下载 Lua 协议...")
	luaName, luaContent, err := cloud.GetProtocolLua(deviceType, *sn)
	if err != nil {
		fmt.Printf("❌ 下载 Lua 协议失败: %v\n", err)
		return err
	}

	// Save Lua file
	if err := os.WriteFile(luaName, []byte(luaContent), 0644); err != nil {
		fmt.Printf("❌ 保存 Lua 文件失败: %v\n", err)
		return err
	}
	fmt.Printf("✅ Lua 协议已保存: %s\n", luaName)

	// Download plugin
	fmt.Println("📥 正在下载插件...")
	pluginName, pluginData, err := cloud.GetPlugin(deviceType, *sn)
	if err != nil {
		fmt.Printf("❌ 下载插件失败: %v\n", err)
		return err
	}

	// Save plugin file
	if err := os.WriteFile(pluginName, pluginData, 0644); err != nil {
		fmt.Printf("❌ 保存插件文件失败: %v\n", err)
		return err
	}
	fmt.Printf("✅ 插件已保存: %s\n", pluginName)

	fmt.Println("\n✅ 下载完成!")
	return nil
}
