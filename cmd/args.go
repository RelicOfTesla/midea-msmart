// midea - CLI tool for controlling Midea smart air conditioners
// Parameter parsing utilities
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
	"github.com/spf13/pflag"
)

// ========== 参数值映射 ==========

var (
	// ModeMap 模式字符串到 OperationalMode 的映射
	ModeMap = map[string]ac.OperationalMode{
		"cool": ac.OperationalModeCool,
		"heat": ac.OperationalModeHeat,
		"auto": ac.OperationalModeAuto,
		"dry":  ac.OperationalModeDry,
		"fan":  ac.OperationalModeFanOnly,
	}

	// ModeNames OperationalMode 到中文名称的映射
	ModeNames = map[ac.OperationalMode]string{
		ac.OperationalModeCool:    "制冷",
		ac.OperationalModeHeat:    "制热",
		ac.OperationalModeAuto:    "自动",
		ac.OperationalModeDry:     "除湿",
		ac.OperationalModeFanOnly: "送风",
	}

	// SpeedMap 风速字符串到 FanSpeed 的映射
	SpeedMap = map[string]ac.FanSpeed{
		"auto":   ac.FanSpeedAuto,
		"low":    ac.FanSpeedLow,
		"medium": ac.FanSpeedMedium,
		"high":   ac.FanSpeedHigh,
		"silent": ac.FanSpeedSilent,
	}

	// SpeedNames FanSpeed 到中文名称的映射
	SpeedNames = map[ac.FanSpeed]string{
		ac.FanSpeedAuto:   "自动",
		ac.FanSpeedLow:    "低",
		ac.FanSpeedMedium: "中",
		ac.FanSpeedHigh:   "高",
		ac.FanSpeedSilent: "静音",
	}

	// SwingMap 摆风字符串到 SwingMode 的映射
	SwingMap = map[string]ac.SwingMode{
		"off":        ac.SwingModeOff,
		"vertical":   ac.SwingModeVertical,
		"horizontal": ac.SwingModeHorizontal,
		"both":       ac.SwingModeBoth,
	}

	// SwingNames SwingMode 到中文名称的映射
	SwingNames = map[ac.SwingMode]string{
		ac.SwingModeOff:        "关闭",
		ac.SwingModeVertical:   "上下摆风",
		ac.SwingModeHorizontal: "左右摆风",
		ac.SwingModeBoth:       "全方位摆风",
	}
)

// ========== 辅助函数 ==========

// ========== 参数解析函数 ==========

// ParseMode 解析模式字符串
func ParseMode(s string) (ac.OperationalMode, error) {
	mode, ok := ModeMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid mode: %s (valid: cool/heat/auto/dry/fan)", s)
	}
	return mode, nil
}

// ParseFanSpeed 解析风速字符串
func ParseFanSpeed(s string) (ac.FanSpeed, error) {
	speed, ok := SpeedMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid fan speed: %s (valid: auto/low/medium/high/silent)", s)
	}
	return speed, nil
}

// ParseSwingMode 解析摆风模式字符串
func ParseSwingMode(s string) (ac.SwingMode, error) {
	swing, ok := SwingMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid swing mode: %s (valid: off/vertical/horizontal/both)", s)
	}
	return swing, nil
}

// ParseTemp 解析温度字符串
func ParseTemp(s string) (float64, error) {
	temp, err := strconv.ParseFloat(s, 64)
	if err != nil || temp < 16 || temp > 30 {
		return 0, fmt.Errorf("invalid temperature: %s (range: 16-30°C)", s)
	}
	return temp, nil
}

// ========== 输出辅助函数 ==========

// PrintError 统一的错误输出
func PrintError(format string, args ...interface{}) {
	fmt.Printf("❌ "+format+"\n", args...)
}

// ========== 命令参数解析函数 ==========

// parseDiscoverArgs 解析 discover 命令参数
func parseDiscoverArgs(args []string) (targetHost string, autoConnect bool, account string, password string, discoveryCount int) {
	discoveryCount = 3 // Default
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--auto-connect", "-a":
			autoConnect = true
		case "--account":
			if i+1 < len(args) {
				account = args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		case "--count":
			if i+1 < len(args) {
				if count, err := strconv.Atoi(args[i+1]); err == nil && count > 0 {
					discoveryCount = count
				}
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && targetHost == "" {
				targetHost = args[i]
			}
		}
	}
	return
}

// parseBindArgs 解析 bind 命令参数
func parseBindArgs(args []string) (identifier string, name string) {
	if len(args) >= 1 {
		identifier = args[0]
	}
	for i := 1; i < len(args); i++ {
		if args[i] == "-n" && i+1 < len(args) {
			name = args[i+1]
			break
		}
	}
	return
}

// parseUnbindArgs 解析 unbind 命令参数
func parseUnbindArgs(args []string) (identifier string) {
	if len(args) >= 1 {
		identifier = args[0]
	}
	return
}

// parseStatusArgs 解析 status 命令参数（使用 flag.FlagSet）
func parseStatusArgs(args []string) (identifier string, autoMode bool, showCapabilities bool, capabilitiesFile string, showEnergy bool) {
	fs := pflag.NewFlagSet("status", pflag.ContinueOnError)
	// 忽略未知 flags（全局 flags）
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")
	fs.BoolVarP(&showCapabilities, "capabilities", "c", false, "Show capabilities")
	fs.StringVarP(&capabilitiesFile, "capabilities-file", "", "", "Save capabilities to file")
	fs.BoolVarP(&showEnergy, "energy", "e", false, "Show energy info")

	fs.Parse(args)

	if fs.NArg() > 0 {
		identifier = fs.Arg(0)
	}
	return
}

// parsePowerArgs 解析 on/off 命令参数（使用 flag.FlagSet）
func parsePowerArgs(args []string) (identifier string, autoMode bool) {

	fs := pflag.NewFlagSet("power", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")

	fs.Parse(args)

	if fs.NArg() > 0 {
		identifier = fs.Arg(0)
	}
	return
}

// parseTempArgs 解析 temp 命令参数（使用 flag.FlagSet）
func parseTempArgs(args []string) (identifier string, temp float64, autoMode bool, err error) {

	fs := pflag.NewFlagSet("temp", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")

	fs.Parse(args)

	if fs.NArg() < 2 {
		PrintError("用法: midea temp <name|id> <温度> [--auto]")
		err = fmt.Errorf("temp requires identifier and temperature")
		return
	}

	identifier = fs.Arg(0)
	temp, err = ParseTemp(fs.Arg(1))
	if err != nil {
		PrintError("无效的温度: %s (范围: 16-30°C)", fs.Arg(1))
		return
	}
	return
}

// parseModeArgs 解析 mode 命令参数（使用 flag.FlagSet）
func parseModeArgs(args []string) (identifier string, mode ac.OperationalMode, autoMode bool, err error) {

	fs := pflag.NewFlagSet("mode", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")

	fs.Parse(args)

	if fs.NArg() < 2 {
		PrintError("用法: midea mode <name|id> <模式> [--auto]")
		err = fmt.Errorf("mode requires identifier and mode value")
		return
	}

	identifier = fs.Arg(0)
	mode, err = ParseMode(fs.Arg(1))
	if err != nil {
		PrintError("无效的模式: %s", fs.Arg(1))
		return
	}
	return
}

// parseFanArgs 解析 fan 命令参数（使用 flag.FlagSet）
func parseFanArgs(args []string) (identifier string, speed ac.FanSpeed, autoMode bool, err error) {

	fs := pflag.NewFlagSet("fan", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")

	fs.Parse(args)

	if fs.NArg() < 2 {
		PrintError("用法: midea fan <name|id> <风速> [--auto]")
		err = fmt.Errorf("fan requires identifier and fan speed")
		return
	}

	identifier = fs.Arg(0)
	speed, err = ParseFanSpeed(fs.Arg(1))
	if err != nil {
		PrintError("无效的风速: %s", fs.Arg(1))
		return
	}
	return
}

// parseSwingArgs 解析 swing 命令参数（使用 flag.FlagSet）
func parseSwingArgs(args []string) (identifier string, swing ac.SwingMode, autoMode bool, err error) {

	fs := pflag.NewFlagSet("swing", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")

	fs.Parse(args)

	if fs.NArg() < 2 {
		PrintError("用法: midea swing <name|id> <模式> [--auto]")
		err = fmt.Errorf("swing requires identifier and swing mode")
		return
	}

	identifier = fs.Arg(0)
	swing, err = ParseSwingMode(fs.Arg(1))
	if err != nil {
		PrintError("无效的摆风模式: %s", fs.Arg(1))
		return
	}
	return
}

// parseSetArgs 解析 set 命令参数（使用 flag.FlagSet）
func parseSetArgs(args []string) (identifier string, autoMode bool, tempValue *float64, modeValue *ac.OperationalMode, speedValue *ac.FanSpeed, swingValue *ac.SwingMode, powerValue *bool, err error) {
	var tempStr, modeStr, fanStr, swingStr, powerStr string


	fs := pflag.NewFlagSet("set", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")
	fs.StringVarP(&tempStr, "temp", "t", "", "Temperature (16-30)")
	fs.StringVarP(&modeStr, "mode", "m", "", "Mode (cool/heat/auto/dry/fan)")
	fs.StringVarP(&fanStr, "fan", "f", "", "Fan speed (auto/low/medium/high/silent)")
	fs.StringVarP(&swingStr, "swing", "s", "", "Swing mode (off/vertical/horizontal/both)")
	fs.StringVarP(&powerStr, "power", "p", "", "Power state (on/off)")

	fs.Parse(args)

	if fs.NArg() < 1 {
		PrintError("用法: midea set <name|id> [选项] [--auto]")
		err = fmt.Errorf("set requires identifier")
		return
	}

	identifier = fs.Arg(0)

	// Parse temp
	if tempStr != "" {
		temp, parseErr := ParseTemp(tempStr)
		if parseErr != nil {
			PrintError("无效的温度: %s (范围: 16-30°C)", tempStr)
			err = parseErr
			return
		}
		tempValue = &temp
	}

	// Parse mode
	if modeStr != "" {
		mode, parseErr := ParseMode(modeStr)
		if parseErr != nil {
			PrintError("无效的模式: %s", modeStr)
			err = parseErr
			return
		}
		modeValue = &mode
	}

	// Parse fan
	if fanStr != "" {
		speed, parseErr := ParseFanSpeed(fanStr)
		if parseErr != nil {
			PrintError("无效的风速: %s", fanStr)
			err = parseErr
			return
		}
		speedValue = &speed
	}

	// Parse swing
	if swingStr != "" {
		swing, parseErr := ParseSwingMode(swingStr)
		if parseErr != nil {
			PrintError("无效的摆风模式: %s", swingStr)
			err = parseErr
			return
		}
		swingValue = &swing
	}

	// Parse power
	if powerStr != "" {
		power := strings.ToLower(powerStr)
		if power != "on" && power != "off" {
			PrintError("无效的电源状态: %s (应为 on 或 off)", powerStr)
			err = fmt.Errorf("invalid power state: %s", powerStr)
			return
		}
		isOn := power == "on"
		powerValue = &isOn
	}

	// Check if at least one option is specified
	if tempValue == nil && modeValue == nil && speedValue == nil && swingValue == nil && powerValue == nil {
		PrintError("未指定任何更改")
		err = fmt.Errorf("no changes specified")
		return
	}

	return
}

// parseQueryArgs 解析 query 命令参数（使用 flag.FlagSet）
func parseQueryArgs(args []string) (identifier string, key string, showAll bool, autoMode bool) {
	showAll = true


	fs := pflag.NewFlagSet("query", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.BoolVarP(&autoMode, "auto", "a", false, "Auto discover device")
	fs.BoolVar(&showAll, "all", true, "Show all properties")

	fs.Parse(args)

	if fs.NArg() > 0 {
		identifier = fs.Arg(0)
	}
	if fs.NArg() > 1 {
		key = fs.Arg(1)
		showAll = false
	}
	return
}

// parseDownloadArgs 解析 download 命令参数
func parseDownloadArgs(args []string) (targetHost string, account string, password string) {
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--account":
			if i+1 < len(args) {
				account = args[i+1]
				i++
			}
		case "--password":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && targetHost == "" {
				targetHost = args[i]
			}
		}
	}
	return
}
