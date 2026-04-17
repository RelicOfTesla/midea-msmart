// midea - CLI tool for controlling Midea smart air conditioners
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/RelicOfTesla/midea-msmart/cmd/config"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
	"github.com/RelicOfTesla/midea-msmart/msmart/device"
	"github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
	"github.com/RelicOfTesla/midea-msmart/msmart/device/xc"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var version = "1.0.0"

// DeviceType string constants for CLI
const (
	DeviceTypeAC = "AC" // Air Conditioner (空调)
	DeviceTypeCC = "CC" // Commercial Air Conditioner (商业空调)
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
	JSON        bool // Output in JSON format

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

// ============================================================================
// Output Types with TextMarshaler and JSONMarshaler
// ============================================================================

// ACState represents AC device state for output
type ACState struct {
	Power              bool     `json:"power"`
	TargetTemperature  float64  `json:"target_temperature"`
	IndoorTemperature  *float64 `json:"indoor_temperature,omitempty"`
	OutdoorTemperature *float64 `json:"outdoor_temperature,omitempty"`
	Mode               string   `json:"mode"`
	FanSpeed           string   `json:"fan_speed"`
	SwingMode          string   `json:"swing_mode"`
	Eco                bool     `json:"eco"`
	Turbo              bool     `json:"turbo"`
}

// MarshalText implements encoding.TextMarshaler for ACState
func (s ACState) MarshalText() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("\n╔════════════════════════════════════════╗\n")
	buf.WriteString("║           📊 空调状态                  ║\n")
	buf.WriteString("╠════════════════════════════════════════╣\n")

	if s.Power {
		buf.WriteString("║  电源: 🟢 开启                         ║\n")
	} else {
		buf.WriteString("║  电源: 🔴 关闭                         ║\n")
		buf.WriteString("╚════════════════════════════════════════╝\n")
		return []byte(buf.String()), nil
	}

	buf.WriteString(fmt.Sprintf("║  目标温度: %.0f°C                      ║\n", s.TargetTemperature))
	if s.IndoorTemperature != nil {
		buf.WriteString(fmt.Sprintf("║  室内温度: %.1f°C                      \n", *s.IndoorTemperature))
	}
	if s.OutdoorTemperature != nil {
		buf.WriteString(fmt.Sprintf("║  室外温度: %.1f°C                      \n", *s.OutdoorTemperature))
	}
	buf.WriteString(fmt.Sprintf("║  运行模式: %-24s║\n", s.Mode))
	buf.WriteString(fmt.Sprintf("║  风速: %-30s║\n", s.FanSpeed))
	buf.WriteString(fmt.Sprintf("║  摆风: %-30s║\n", s.SwingMode))

	if s.Eco {
		buf.WriteString("║  🌿 ECO模式: 开启                      ║\n")
	}
	if s.Turbo {
		buf.WriteString("║  🚀 强力模式: 开启                     ║\n")
	}

	buf.WriteString("╚════════════════════════════════════════╝\n")
	return []byte(buf.String()), nil
}

// MarshalJSON implements json.Marshaler for ACState
// Returns structured JSON object instead of text table
func (s ACState) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias ACState
	return json.Marshal(Alias(s))
}

// EnergyUsage represents energy usage data for output
type EnergyUsage struct {
	RealTimePower *float64 `json:"real_time_power_w,omitempty"`
	CurrentMonth  *float64 `json:"current_month_kwh,omitempty"`
	TotalEnergy   *float64 `json:"total_energy_kwh,omitempty"`
}

// MarshalText implements encoding.TextMarshaler for EnergyUsage
func (e EnergyUsage) MarshalText() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("\n╔════════════════════════════════════════╗\n")
	buf.WriteString("║           ⚡ 能耗信息                  ║\n")
	buf.WriteString("╠════════════════════════════════════════╣\n")

	hasEnergyData := false

	if e.RealTimePower != nil {
		buf.WriteString(fmt.Sprintf("║  实时功率: %.1f W                      \n", *e.RealTimePower))
		hasEnergyData = true
	}
	if e.CurrentMonth != nil {
		buf.WriteString(fmt.Sprintf("║  本月能耗: %.2f kWh                    \n", *e.CurrentMonth))
		hasEnergyData = true
	}
	if e.TotalEnergy != nil {
		buf.WriteString(fmt.Sprintf("║  累计能耗: %.2f kWh                    \n", *e.TotalEnergy))
		hasEnergyData = true
	}

	if !hasEnergyData {
		buf.WriteString("║  ⚠️  无能耗数据                        ║\n")
	}

	buf.WriteString("╚════════════════════════════════════════╝\n")
	return []byte(buf.String()), nil
}

// MarshalJSON implements json.Marshaler for EnergyUsage
// Returns structured JSON object instead of text table
func (e EnergyUsage) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias EnergyUsage
	return json.Marshal(Alias(e))
}

// Capabilities represents device capabilities for output
type Capabilities struct {
	SupportedFeatures   []string `json:"supported_features,omitempty"`
	SupportedModes      []string `json:"supported_modes,omitempty"`
	SupportedFanSpeeds  []string `json:"supported_fan_speeds,omitempty"`
	SupportedSwingModes []string `json:"supported_swing_modes,omitempty"`
	MinTemperature      *int     `json:"min_temperature,omitempty"`
	MaxTemperature      *int     `json:"max_temperature,omitempty"`
}

// MarshalText implements encoding.TextMarshaler for Capabilities
func (c Capabilities) MarshalText() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("\n╔════════════════════════════════════════╗\n")
	buf.WriteString("║         📋 设备能力信息                ║\n")
	buf.WriteString("╠════════════════════════════════════════╣\n")

	hasData := false

	if len(c.SupportedFeatures) > 0 {
		buf.WriteString("║  支持的功能:                           ║\n")
		for _, flag := range c.SupportedFeatures {
			buf.WriteString(fmt.Sprintf("║    • %-32s║\n", flag))
		}
		hasData = true
	}

	if len(c.SupportedModes) > 0 {
		buf.WriteString("║  支持的模式:                           ║\n")
		buf.WriteString(fmt.Sprintf("║    %s                                ║\n", strings.Join(c.SupportedModes, ", ")))
		hasData = true
	}

	if len(c.SupportedFanSpeeds) > 0 {
		buf.WriteString("║  支持的风速:                           ║\n")
		buf.WriteString(fmt.Sprintf("║    %s                                ║\n", strings.Join(c.SupportedFanSpeeds, ", ")))
		hasData = true
	}

	if len(c.SupportedSwingModes) > 0 {
		buf.WriteString("║  支持的摆风:                           ║\n")
		buf.WriteString(fmt.Sprintf("║    %s                              ║\n", strings.Join(c.SupportedSwingModes, ", ")))
		hasData = true
	}

	if c.MinTemperature != nil && c.MaxTemperature != nil {
		buf.WriteString(fmt.Sprintf("║  温度范围: %d°C - %d°C                 ║\n", *c.MinTemperature, *c.MaxTemperature))
		hasData = true
	}

	if !hasData {
		buf.WriteString("║  ⚠️  无能力信息                        ║\n")
	}

	buf.WriteString("╚════════════════════════════════════════╝\n")
	return []byte(buf.String()), nil
}

// MarshalJSON implements json.Marshaler for Capabilities
// Returns structured JSON object instead of text table
func (c Capabilities) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias Capabilities
	return json.Marshal(Alias(c))
}

// DeviceInfo represents a single device's information
type DeviceInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	IP     string `json:"ip"`
	Online bool   `json:"online"`
	Status string `json:"status"` // For JSON output: "在线" or "离线"
}

// DeviceList represents a list of devices
type DeviceList []DeviceInfo

// MarshalText implements encoding.TextMarshaler for DeviceList
// Returns formatted table output
func (dl DeviceList) MarshalText() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("\n📋 设备列表:\n")
	buf.WriteString("─────────────────────────────────────────────────────────────\n")

	for _, d := range dl {
		status := "🔴 离线"
		if d.Online {
			status = "🟢 在线"
		}
		buf.WriteString(fmt.Sprintf("  %s %-12s - %-8s %s (ID: %s)\n", status, d.Name, d.Type, d.IP, d.ID))
	}

	buf.WriteString("─────────────────────────────────────────────────────────────\n")
	return []byte(buf.String()), nil
}

// MarshalJSON implements json.Marshaler for DeviceList
// Returns structured JSON array
func (dl DeviceList) MarshalJSON() ([]byte, error) {
	return json.Marshal([]DeviceInfo(dl))
}

// setupFlags 设置命令行参数
func setupFlags(flags *GlobalFlags) {
	// Global flags
	pflag.StringVarP(&flags.Region, "region", "r", msmart.DefaultCloudRegion, "Cloud region (DE, KR, US)")
	pflag.StringVarP(&flags.DeviceType, "device_type", "d", "AC", "Device type (AC, CC)")
	pflag.IntVarP(&flags.DeviceID, "id", "i", 0, "Device ID for V3 devices")
	pflag.StringVarP(&flags.DeviceToken, "token", "T", "", "Auth token for V3 devices")
	pflag.StringVarP(&flags.DeviceKey, "key", "k", "", "Auth key for V3 devices")
	pflag.BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	pflag.BoolVarP(&flags.JSON, "json", "j", false, "Output in JSON format")

	// Subcommand flags
	pflag.BoolVarP(&flags.AutoMode, "auto", "a", false, "Auto discover device")
	pflag.StringVar(&flags.Account, "account", "", "Midea account")
	pflag.StringVar(&flags.Password, "password", "", "Midea password")
	pflag.BoolVarP(&flags.AutoConnect, "auto-connect", "c", false, "Auto connect and get token")
	pflag.IntVar(&flags.DiscoveryCount, "count", 3, "Discovery packet count")
	pflag.BoolVarP(&flags.ShowEnergy, "energy", "e", false, "Show energy info")
	pflag.BoolVar(&flags.ShowAll, "all", false, "Show all properties")
	pflag.BoolVar(&flags.ShowCapabilities, "capabilities", false, "Show device capabilities")
	pflag.StringVar(&flags.CapabilitiesFile, "capabilities-file", "", "Save capabilities to file")
	pflag.StringVarP(&flags.TempValue, "temp", "t", "", "Temperature (16-30)")
	pflag.StringVarP(&flags.ModeValue, "mode", "m", "", "Mode (cool/heat/auto/dry/fan)")
	pflag.StringVarP(&flags.FanValue, "fan", "f", "", "Fan speed (auto/low/medium/high/silent)")
	pflag.StringVarP(&flags.SwingValue, "swing", "s", "", "Swing mode (off/vertical/horizontal/both)")
	pflag.StringVarP(&flags.PowerValue, "power", "p", "", "Power state (on/off)")
	pflag.StringVarP(&flags.Name, "name", "n", "", "Device name")

	pflag.Usage = printUsage
}

// setupLogger 设置日志输出
// 如果 jsonMode 为 true，日志输出为 JSON 格式到 stderr
// 否则输出为格式化文本到 stdout
func setupLogger(jsonMode bool) {
	var handler slog.Handler
	if jsonMode {
		// JSON 格式输出到 stderr
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// 格式化文本输出到 stdout（使用自定义 Handler）
		handler = NewPrettyTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	slog.SetDefault(slog.New(handler))
}

func main() {
	flags := &GlobalFlags{}
	setupFlags(flags)

	// 先解析命令行参数
	pflag.Parse()

	// 设置日志输出（根据 --json 参数）
	setupLogger(flags.JSON)

	if err := run(flags); err != nil {
		// Print error if not already printed
		slog.Error("错误", "error", err)
		os.Exit(1)
	}
}

// parseCommand 解析命令行参数，返回获取命令参数的函数
// 使用 lambda 表达式封装参数访问
func parseCommand() func(int) string {
	pflag.Parse()
	fullArgs := pflag.Args()
	args := make([]string, 0, len(fullArgs))
	for _, arg := range fullArgs {
		if !strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	return func(i int) string {
		if i >= 0 && i < len(args) {
			return args[i]
		}
		return ""
	}
}

func run(flags *GlobalFlags) error {
	ctx := context.Background()

	// 一次性解析所有参数，返回 lambda 函数
	getCommand := parseCommand()

	if flags.Verbose {
		msmart.Verbose = true
	}

	// Validate device type
	deviceTypeStr := strings.ToUpper(flags.DeviceType)
	if _, ok := deviceTypeMap[deviceTypeStr]; !ok {
		slog.Error("不支持的设备类型", "type", flags.DeviceType, "supported", "AC (空调), CC (商业空调)")
		return fmt.Errorf("unsupported device type: %s", flags.DeviceType)
	}

	// 获取命令
	command := getCommand(0)
	if command == "" {
		printUsage()
		return fmt.Errorf("no command provided")
	}

	configPath := config.DefaultConfigPath()

	switch command {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "version", "--version":
		slog.Info("midea", "version", version)
		return nil
	case "discover":
		return handleDiscover(ctx, configPath, flags, getCommand(1))
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
		return handleStatus(ctx, configPath, flags, deviceTypeStr, getCommand(1))
	case "on":
		return handlePower(ctx, configPath, flags, deviceTypeStr, true, getCommand(1))
	case "off":
		return handlePower(ctx, configPath, flags, deviceTypeStr, false, getCommand(1))
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
		return handleTemp(ctx, configPath, flags, deviceTypeStr, identifier, temp)
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
		return handleMode(ctx, configPath, flags, deviceTypeStr, identifier, mode)
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
		return handleFan(ctx, configPath, flags, deviceTypeStr, identifier, speed)
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
		return handleSwing(ctx, configPath, flags, deviceTypeStr, identifier, swing)
	case "set":
		return handleSet(ctx, configPath, flags, deviceTypeStr, getCommand(1))
	case "query":
		return handleQuery(ctx, configPath, flags, deviceTypeStr, getCommand(1), getCommand(2))
	case "download":
		return handleDownload(ctx, configPath, flags, getCommand(1))
	default:
		slog.Error("未知命令", "command", command)
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
  -i, --id <设备ID>    V3设备ID
  -T, --token <token>  V3设备认证token
  -k, --key <key>      V3设备认证key
  -j, --json           以JSON格式输出

命令:
  help                          显示帮助信息
  version                       显示版本信息

  discover [<host>] [--auto|-a] [--auto-connect|-c] [--count <数量>] [--account <账号> --password <密码>]
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
                                key: 属性名称
                                  temp/temperature/target_temp - 目标温度
                                  indoor_temp/indoor_temperature - 室内温度
                                  outdoor_temp/outdoor_temperature - 室外温度
                                  mode/operational_mode - 运行模式
                                  fan/fan_speed - 风速
                                  swing/swing_mode - 摆风模式
                                  power/power_state - 电源状态
                                  eco - ECO模式
                                  turbo - 强力模式
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

  # 查询设备属性
  midea query 客厅                  # 显示所有属性
  midea query 客厅 temp             # 查询目标温度
  midea query 客厅 indoor_temp      # 查询室内温度
  midea query 客厅 outdoor_temp     # 查询室外温度
  midea query 客厅 mode             # 查询运行模式
  midea query 客厅 fan              # 查询风速
  midea query 客厅 swing            # 查询摆风模式
  midea query 客厅 power            # 查询电源状态
  midea query 客厅 eco              # 查询ECO模式
  midea query 客厅 turbo            # 查询强力模式

  # 下载设备协议和插件
  midea download 192.168.1.60 --account your@email.com --password yourpass

配置文件: 当前目录的 midea.json (优先) 或 ~/.config/midea/config.json
`)
}

// ============================================================================
// Discovery Commands
// ============================================================================

func handleDiscover(ctx context.Context, configPath string, flags *GlobalFlags, targetHost string) error {
	slog.Debug("正在发现设备...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
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
		slog.Debug("目标设备", "host", targetHost)
	}

	// Set account and password if provided
	if flags.Account != "" && flags.Password != "" {
		discoverConfig.Account = flags.Account
		discoverConfig.Password = flags.Password
	}

	devices, err := msmart.Discover(ctx, discoverConfig)

	// Even if there's an error, check if we discovered any devices
	if err != nil && len(devices) == 0 {
		slog.Error("发现设备失败", "error", err)
		return err
	}

	// Log warning if there was an error but we have devices
	if err != nil && len(devices) > 0 {
		slog.Warn("发现设备时有错误", "error", err)
	}

	if len(devices) == 0 {
		slog.Warn("未发现任何设备")
		return nil
	}

	slog.Debug("发现设备", "count", len(devices))

	for _, d := range devices {
		// Get device info
		deviceID := d.GetID()
		deviceType := "未知设备"
		if d.GetType() == msmart.DeviceTypeAirConditioner {
			deviceType = "空调"
		}

		// Get name
		name := d.GetName()

		// Check if device already exists in config
		existingDevice := cfg.GetDevice(deviceID)
		if existingDevice != nil {
			// Keep the existing name
			name = existingDevice.Name
		}

		// Save device to config
		devCfg := config.Device{
			ID:   deviceID,
			Name: name,
			IP:   d.GetIP(),
			Port: d.GetPort(),
			SN:   d.GetSN(),
			Type: int(d.GetType()),
			// Token:   token,
			// Key:     key,
			Version: d.GetVersion(),
			Online:  d.GetOnline(),
		}
		if dev, ok := d.(device.DeviceAuthV3); ok {
			if token, key, localKey, expired := dev.GetKeyInfo(); token != nil && key != nil {
				devCfg.Token = hex.EncodeToString(token)
				devCfg.Key = hex.EncodeToString(key)
				devCfg.LocalKey = hex.EncodeToString(localKey)
				devCfg.LocalKeyExpire = expired.Format(time.RFC3339)
			}
		}
		cfg.AddDevice(devCfg)

		// Print device info
		deviceName := name
		if deviceName == "" {
			deviceName = "(未命名)"
		}

		slog.Info("发现设备", "device", DeviceInfo{
			ID:     deviceID,
			Name:   deviceName,
			Type:   deviceType,
			IP:     d.GetIP(),
			Online: d.GetOnline(),
		})
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		slog.Error("保存配置失败", "error", err)
		return err
	}

	slog.Info("配置已保存", "path", configPath)
	slog.Info("使用 'midea bind <id|ip> -n <名称>' 来为设备命名")
	return nil
}

func handleList(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		return err
	}

	devices := cfg.ListDevices()
	if len(devices) == 0 {
		slog.Info("暂无设备配置")
		slog.Info("使用 'midea discover' 来发现设备")
		return nil
	}

	// Build device list
	deviceList := make(DeviceList, 0, len(devices))
	for _, d := range devices {
		deviceType := "未知设备"
		if d.Type == 0xAC {
			deviceType = "空调"
		}

		name := d.Name
		if name == "" {
			name = "(未命名)"
		}

		deviceList = append(deviceList, DeviceInfo{
			ID:     d.ID,
			Name:   name,
			Type:   deviceType,
			IP:     d.IP,
			Online: d.Online,
		})
	}

	slog.Info("设备列表", "devices", deviceList)
	slog.Info("配置文件", "path", configPath)
	return nil
}

// ============================================================================
// Device Management Commands
// ============================================================================

func handleBind(configPath string, identifier string, name string) error {
	if name == "" {
		slog.Error("请指定名称", "usage", "midea bind <id|sn|ip> -n <名称>")
		return fmt.Errorf("name not specified for bind command")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		return err
	}

	if !cfg.BindName(identifier, name) {
		slog.Error("未找到设备", "identifier", identifier)
		slog.Info("使用 'midea list' 查看设备列表")
		return fmt.Errorf("device not found: %s", identifier)
	}

	if err := cfg.Save(configPath); err != nil {
		slog.Error("保存配置失败", "error", err)
		return err
	}

	slog.Info("已绑定", "identifier", identifier, "name", name)
	return nil
}

func handleUnbind(configPath string, identifier string) error {
	if identifier == "" {
		slog.Error("用法错误", "usage", "midea unbind <name|id>")
		return fmt.Errorf("insufficient arguments for unbind command")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		return err
	}

	if !cfg.RemoveDevice(identifier) {
		slog.Error("未找到设备", "identifier", identifier)
		slog.Info("使用 'midea list' 查看设备列表")
		return fmt.Errorf("device not found: %s", identifier)
	}

	if err := cfg.Save(configPath); err != nil {
		slog.Error("保存配置失败", "error", err)
		return err
	}

	slog.Info("已解绑", "identifier", identifier)
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

func findDeviceConfigByConfig(ctx context.Context, anyId string, configPath string, flags *GlobalFlags) (*config.Config, *config.Device, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		return nil, nil, err
	}
	var devCfg *config.Device
	if anyId != "" {
		devCfg = cfg.GetDevice(anyId)
	}
	if devCfg == nil && flags.DeviceID != 0 {
		devCfg = cfg.GetDevice(strconv.Itoa(flags.DeviceID))
	}
	if devCfg == nil && flags.Name != "" {
		devCfg = cfg.GetDevice(flags.Name)
	}
	if devCfg == nil {
		return nil, nil, fmt.Errorf("未找到设备: %s (使用 'midea list' 查看设备列表 或使用 --auto 自动发现设备)", anyId)
	}
	return cfg, devCfg, nil
}
func connectDeviceByConfig(ctx context.Context, anyId string, configPath string, flags *GlobalFlags) (device.Device, error) {
	cfg, devCfg, err := findDeviceConfigByConfig(ctx, anyId, configPath, flags)
	if err != nil {
		return nil, err
	}

	// Print connection info
	slog.Debug("目标设备", "name", devCfg.Name, "ip", devCfg.IP)
	slog.Debug("正在连接...")
	dev, err := connectDevice(ctx, devCfg, flags)
	if err == nil {
		// Save to config
		cfg.AddDevice(*devCfg)
		if err := cfg.Save(configPath); err != nil {
			slog.Warn("保存配置失败", "error", err)
		}
	}
	return dev, err
}
func connectDevice(ctx context.Context, devCfg *config.Device, flags *GlobalFlags) (device.Device, error) {

	token, key, localKey, expired, err := devCfg.GetValidKeys()
	if err != nil {
		return nil, err
	}

	var dev device.Device

	if flags.AutoConnect {
		opt := msmart.DiscoverConfig{
			Target:           devCfg.IP,
			Region:           "",
			Account:          flags.Account,
			Password:         flags.Password,
			AutoConnect:      true,
			ExistingToken:    device.Token(token),
			ExistingKey:      device.Key(key),
			ExistingLocalKey: device.LocalKey(localKey),
			// ExistingLocalKeyExpired: &time.Time{},
		}
		if localKey != nil && expired.IsZero() && time.Now().Before(expired) {
			opt.ExistingLocalKeyExpired = &expired
		}
		dev, err = msmart.DiscoverSingle(ctx, opt.Target, &opt)
		if err != nil {
			return nil, err
		}

	} else {
		var (
			opts []msmart.DeviceOption
		)
		if devCfg.IP != "" {
			opts = append(opts, msmart.WithDeviceAddr(devCfg.IP, devCfg.Port))
		}
		if devCfg.ID != "" {
			opts = append(opts, msmart.WithDeviceID(devCfg.ID))
		}

		if token != nil && key != nil {
			opts = append(opts, msmart.WithTokenKey(token, key))
		}
		if localKey != nil && expired.IsZero() && time.Now().Before(expired) {
			opts = append(opts, msmart.WithLocalKey(localKey, expired))
		}
		deviceType := device.DeviceType(devCfg.Type)
		switch strings.ToUpper(flags.DeviceType) {
		case DeviceTypeCC:
			deviceType = msmart.DeviceTypeCommercialAC
		case DeviceTypeAC:
			deviceType = msmart.DeviceTypeAirConditioner
		}

		dev = device.NewDeviceFromType(deviceType, opts...)
		if dev == nil {
			return nil, fmt.Errorf("unsupported device type: %v", deviceType)
		}
	}

	if dev, ok := dev.(device.DeviceAuthV3); ok {
		if localKey != nil && expired.IsZero() && time.Now().Before(expired) {
			slog.Debug("使用缓存的 LocalKey 已认证")
		} else if token != nil && key != nil && !dev.IsAuthenticated() {
			slog.Debug("正在认证...")
			ok, err := dev.AuthenticateFromPreset(ctx)
			if !ok {
				return nil, errors.New("authentication failed")
			}
			if err != nil {
				return nil, err
			}
			slog.Debug("认证成功")
		} else {
			return nil, fmt.Errorf("V3设备需要token和key进行认证 (使用 '--auto' 参数自动获取token/key,或使用 'midea discover --auto-connect' 重新发现设备)")
		}
	}

	devCfg.ID = dev.GetID()
	// devCfg.Name = dev.GetName()
	devCfg.IP = dev.GetIP()
	devCfg.Port = dev.GetPort()
	devCfg.SN = dev.GetSN()
	devCfg.Type = int(dev.GetType())
	devCfg.Online = dev.GetOnline()

	if dev, ok := dev.(device.DeviceAuthV3); ok {
		if token, key, localKey, expired := dev.GetKeyInfo(); token != nil && key != nil {
			devCfg.Token = hex.EncodeToString(token)
			devCfg.Key = hex.EncodeToString(key)
			devCfg.LocalKey = hex.EncodeToString(localKey)
			devCfg.LocalKeyExpire = expired.Format(time.RFC3339)
		}
	}

	return dev, nil
}

func handleStatus(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string) error {
	var acDevice ac.AC
	var err error

	device, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Get capabilities if requested
	if flags.ShowCapabilities || flags.CapabilitiesFile != "" {
		if err := device.GetCapabilities(ctx); err != nil {
			slog.Warn("获取设备能力失败", "error", err)
		} else {
			// If a file path is specified, write to file
			if flags.CapabilitiesFile != "" {
				if err := writeCapabilitiesToYAML(device, flags.CapabilitiesFile); err != nil {
					slog.Error("写入能力信息到文件失败", "error", err)
				} else {
					slog.Info("设备能力已写入", "file", flags.CapabilitiesFile)
				}
			} else {
				// Display capabilities to screen
				printCapabilities(device)
			}
		}
	}
	acDevice, ok := device.(ac.AC)
	if !ok {
		slog.Warn("设备不是AC类型", "type", fmt.Sprintf("%T", device))
		// 继续执行，但不打印AC特定状态
	}

	// Enable energy usage requests if --energy flag is set
	if acDevice != nil && flags.ShowEnergy {
		acDevice.SetEnableEnergyUsageRequests(true)
	}

	// Refresh state
	if err := device.Refresh(ctx); err != nil {
		slog.Error("查询失败", "error", err)
		return err
	}

	if acDevice != nil {
		// Print state
		printACState(acDevice)

		// Display energy usage if requested
		if flags.ShowEnergy {
			printEnergyUsage(acDevice)
		}
	}
	return nil
}

func printCapabilities(acDevice device.Device) {
	// Get capabilities dictionary
	caps := acDevice.CapabilitiesDict()

	// Create Capabilities struct
	capabilities := Capabilities{}

	if caps != nil {
		if flags, ok := caps["supported_features"].([]string); ok {
			capabilities.SupportedFeatures = flags
		}
		if modes, ok := caps["supported_modes"].([]string); ok {
			capabilities.SupportedModes = modes
		}
		if fans, ok := caps["supported_fan_speeds"].([]string); ok {
			capabilities.SupportedFanSpeeds = fans
		}
		if swings, ok := caps["supported_swing_modes"].([]string); ok {
			capabilities.SupportedSwingModes = swings
		}
		if minTemp, ok := caps["min_temperature"].(int); ok {
			capabilities.MinTemperature = &minTemp
		}
		if maxTemp, ok := caps["max_temperature"].(int); ok {
			capabilities.MaxTemperature = &maxTemp
		}
	}

	// Output using slog (will automatically call MarshalText/MarshalJSON)
	slog.Info("设备能力信息", "capabilities", capabilities)
}

// writeCapabilitiesToYAML writes device capabilities to a YAML file
func writeCapabilitiesToYAML(acDevice device.Device, filename string) error {
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

func printACState(acDevice xc.XCDevice) {
	// Get mode name
	modeNames := map[xc.OperationalMode]string{
		xc.OperationalModeCool:    "cool",
		xc.OperationalModeHeat:    "heat",
		xc.OperationalModeAuto:    "auto",
		xc.OperationalModeDry:     "dry",
		xc.OperationalModeFanOnly: "fan_only",
	}
	modeName := modeNames[acDevice.OperationalMode()]

	// Get fan speed name
	fanSpeed := acDevice.FanSpeed()
	var fanName = SpeedNames[fanSpeed]
	// Get swing mode name
	swingName := SwingNames[acDevice.SwingMode()]

	// Create state struct
	state := ACState{
		Power:              acDevice.PowerState(),
		TargetTemperature:  acDevice.TargetTemperature(),
		IndoorTemperature:  acDevice.IndoorTemperature(),
		OutdoorTemperature: acDevice.OutdoorTemperature(),
		Mode:               modeName,
		FanSpeed:           fanName,
		SwingMode:          swingName,
		Eco:                acDevice.Eco(),
		Turbo:              acDevice.Turbo(),
	}

	// Output using slog (will automatically call MarshalText/MarshalJSON)
	slog.Info("空调状态", "state", state)
}
func valPtrMust[T comparable](v T) *T {
	var zero T
	if zero == v {
		return nil
	}
	return &v
}

func printEnergyUsage(acDevice ac.AC) {
	// Create energy usage struct
	energy := EnergyUsage{
		RealTimePower: acDevice.GetRealTimePowerUsage(ac.EnergyDataFormatBCD),
		CurrentMonth:  acDevice.GetCurrentEnergyUsage(ac.EnergyDataFormatBCD),
		TotalEnergy:   acDevice.GetTotalEnergyUsage(ac.EnergyDataFormatBCD),
	}

	// Output using slog (will automatically call MarshalText/MarshalJSON)
	slog.Info("能耗信息", "energy", energy)
}

func handlePower(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, on bool, identifier string) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if acDevice, ok := dev.(device.SimpleDevice); ok {
		// Set power state
		acDevice.SetPowerState(on)
	} else {
		slog.Error("设备不支持电源控制", "type", reflect.TypeOf(dev))
	}

	// Apply changes
	if err := dev.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	action := "已开机"
	if !on {
		action = "已关机"
	}
	slog.Info(dev.GetName()+" "+action, "success", true)
	return nil
}

func handleTemp(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, temp float64) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if acDevice, ok := dev.(xc.XCDevice); ok {

		// Set temperature
		acDevice.SetTargetTemperature(temp)
	} else {
		slog.Error("设备不支持温度控制", "type", reflect.TypeOf(dev))
	}

	// Apply changes
	if err := dev.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	slog.Info("温度已设置", "device", dev.GetName(), "temp", temp)
	return nil
}

func handleMode(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, mode xc.OperationalMode) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}
	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if acDevice, ok := dev.(xc.XCDevice); ok {
		// Set power state
		acDevice.SetOperationalMode(mode)
	} else {
		slog.Error("设备不支持模式控制", "type", reflect.TypeOf(dev))
	}

	// Apply changes
	if err := dev.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	slog.Info("模式已设置", "device", dev.GetName(), "mode", ModeNames[mode])
	return nil
}

func handleFan(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, speed xc.FanSpeed) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if acDevice, ok := dev.(xc.XCDevice); ok {
		// Set fan speed
		acDevice.SetFanSpeed(speed)
	} else {
		slog.Error("设备不支持风速控制", "type", reflect.TypeOf(dev))
	}
	// Apply changes
	if err := dev.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	slog.Info("风速已设置", "device", dev.GetName(), "speed", SpeedNames[speed])
	return nil
}

func handleSwing(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, swing xc.SwingMode) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if acDevice, ok := dev.(xc.XCDevice); ok {
		// Set swing mode
		acDevice.SetSwingMode(swing)
	} else {
		slog.Error("设备不支持摆风控制", "type", reflect.TypeOf(dev))
	}

	// Apply changes
	if err := dev.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	slog.Info("摆风已设置", "device", dev.GetName(), "swing", SwingNames[swing])
	return nil
}

// handleSet handles the set command for multi-parameter control
func handleSet(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}
	
	// Debug: output actual type
	slog.Info("调试: 设备信息", "type", reflect.TypeOf(dev), "value", reflect.ValueOf(dev), "isNil", dev == nil)
	
	acDevice, ok := dev.(xc.XCDevice)
	if !ok {
		slog.Error("设备不支持空调控制", "type", reflect.TypeOf(dev))
		return fmt.Errorf("device does not support AC control")
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
		slog.Error("未指定任何更改")
		return fmt.Errorf("no changes specified")
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Apply changes
	if err := acDevice.Apply(ctx); err != nil {
		slog.Error("控制失败", "error", err)
		return err
	}

	slog.Info("已设置", "device", dev.GetName(), "changes", strings.Join(changes, ", "))
	return nil
}
func handleQuery(ctx context.Context, configPath string, flags *GlobalFlags, deviceTypeStr string, identifier string, key string) error {
	dev, err := connectDeviceByConfig(ctx, identifier, configPath, flags)
	if err != nil {
		return err
	}

	// Type assert to AC interface
	acDevice, ok := dev.(ac.AC)
	if !ok {
		return fmt.Errorf("device does not support AC interface")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Refresh state
	if err := dev.Refresh(ctx); err != nil {
		slog.Error("查询失败", "error", err)
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

func printSpecificAttribute(acDevice ac.AC, key string) error {
	switch strings.ToLower(key) {
	case "temp", "temperature", "target_temp":
		slog.Info("目标温度", "value", fmt.Sprintf("%.0f°C", acDevice.TargetTemperature()))

	case "indoor_temp", "indoor_temperature":
		if temp := acDevice.IndoorTemperature(); temp != nil && *temp > 0 {
			slog.Info("室内温度", "value", fmt.Sprintf("%.1f°C", *temp))
		} else {
			slog.Warn("室内温度不可用")
		}

	case "outdoor_temp", "outdoor_temperature":
		if temp := acDevice.OutdoorTemperature(); temp != nil && *temp > 0 {
			slog.Info("室外温度", "value", fmt.Sprintf("%.1f°C", *temp))
		} else {
			slog.Warn("室外温度不可用")
		}

	case "mode", "operational_mode":
		slog.Info("运行模式", "value", ModeNames[acDevice.OperationalMode()])

	case "fan", "fan_speed":
		fanSpeed := acDevice.FanSpeed()
		fanName := SpeedNames[fanSpeed]
		slog.Info("风速", "value", fanName)

	case "swing", "swing_mode":
		slog.Info("摆风模式", "value", SwingNames[acDevice.SwingMode()])

	case "power", "power_state":
		if acDevice.PowerState() {
			slog.Info("电源状态", "value", "开启")
		} else {
			slog.Info("电源状态", "value", "关闭")
		}

	case "eco":
		if acDevice.Eco() {
			slog.Info("ECO模式", "value", "开启")
		} else {
			slog.Info("ECO模式", "value", "关闭")
		}

	case "turbo":
		if acDevice.Turbo() {
			slog.Info("强力模式", "value", "开启")
		} else {
			slog.Info("强力模式", "value", "关闭")
		}

	default:
		slog.Error("未知属性", "key", key)
		slog.Info("支持的属性:")
		slog.Info("  temp, temperature, target_temp       - 目标温度")
		slog.Info("  indoor_temp, indoor_temperature      - 室内温度")
		slog.Info("  outdoor_temp, outdoor_temperature    - 室外温度")
		slog.Info("  mode, operational_mode               - 运行模式")
		slog.Info("  fan, fan_speed                       - 风速")
		slog.Info("  swing, swing_mode                    - 摆风模式")
		slog.Info("  power, power_state                   - 电源状态")
		slog.Info("  eco                                  - ECO模式")
		slog.Info("  turbo                                - 强力模式")
		return fmt.Errorf("unknown attribute: %s", key)
	}
	return nil
}

// handleDownload handles the download command for downloading device protocol and plugin
func handleDownload(ctx context.Context, configPath string, flags *GlobalFlags, host string) error {
	if host == "" {
		slog.Error("用法: midea download <host> [--account <账号> --password <密码>]")
		return fmt.Errorf("insufficient arguments for download command")
	}

	slog.Debug("正在发现设备", "host", host)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Discover the device (no auto-connect, we just need SN)
	discoverConfig := &msmart.DiscoverConfig{
		Target:           host,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 3,
		AutoConnect:      false, // Don't connect, just discover
	}

	devices, err := msmart.Discover(ctx, discoverConfig)
	if err != nil {
		slog.Error("发现设备失败", "error", err)
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

	if sn == "" {
		return fmt.Errorf("设备没有 SN,无法下载协议")
	}

	slog.Debug("发现设备", "type", fmt.Sprintf("%02X", deviceType), "sn", sn)

	// Create cloud client
	slog.Debug("正在连接云端...")

	var cloud *msmart.SmartHomeCloud
	var accountPtr, passwordPtr *string

	if flags.Account != "" && flags.Password != "" {
		accountPtr = &flags.Account
		passwordPtr = &flags.Password
	}

	cloud, err = msmart.NewSmartHomeCloud(flags.Region, accountPtr, passwordPtr, false, nil)
	if err != nil {
		slog.Error("创建云端客户端失败", "error", err)
		return err
	}

	// Login to cloud
	if err := cloud.Login(false); err != nil {
		slog.Error("云端登录失败", "error", err)
		return err
	}

	slog.Debug("云端登录成功")

	// Download Lua protocol
	slog.Debug("正在下载 Lua 协议...")
	luaName, luaContent, err := cloud.GetProtocolLua(deviceType, sn)
	if err != nil {
		slog.Error("下载 Lua 协议失败", "error", err)
		return err
	}

	// Save Lua file
	if err := os.WriteFile(luaName, []byte(luaContent), 0644); err != nil {
		slog.Error("保存 Lua 文件失败", "error", err)
		return err
	}
	slog.Debug("Lua 协议已保存", "file", luaName)

	// Download plugin
	slog.Debug("正在下载插件...")
	pluginName, pluginData, err := cloud.GetPlugin(deviceType, sn)
	if err != nil {
		slog.Error("下载插件失败", "error", err)
		return err
	}

	// Save plugin file
	if err := os.WriteFile(pluginName, pluginData, 0644); err != nil {
		slog.Error("保存插件文件失败", "error", err)
		return err
	}
	slog.Info("插件已保存", "file", pluginName)

	slog.Debug("下载完成!")
	return nil
}
