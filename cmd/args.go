// midea - CLI tool for controlling Midea smart air conditioners
// Parameter parsing utilities
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
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

// ========== 参数解析辅助函数 ==========

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
