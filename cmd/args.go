// midea - CLI tool for controlling Midea smart air conditioners
// Parameter parsing utilities
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RelicOfTesla/midea-msmart/msmart/device/xc"
)

// ========== 参数值映射 ==========

var (
	// ModeMap 模式字符串到 OperationalMode 的映射
	ModeMap = map[string]xc.OperationalMode{
		"cool": xc.OperationalModeCool,
		"heat": xc.OperationalModeHeat,
		"auto": xc.OperationalModeAuto,
		"dry":  xc.OperationalModeDry,
		"fan":  xc.OperationalModeFanOnly,
	}

	// ModeNames OperationalMode 到中文名称的映射
	ModeNames = map[xc.OperationalMode]string{
		xc.OperationalModeCool:    "制冷",
		xc.OperationalModeHeat:    "制热",
		xc.OperationalModeAuto:    "自动",
		xc.OperationalModeDry:     "除湿",
		xc.OperationalModeFanOnly: "送风",
	}

	// SpeedMap 风速字符串到 FanSpeed 的映射
	SpeedMap = map[string]xc.FanSpeed{
		"auto":   xc.FanSpeedAuto,
		"low":    xc.FanSpeedLow,
		"medium": xc.FanSpeedMedium,
		"high":   xc.FanSpeedHigh,
		"silent": xc.FanSpeedSilent,
	}

	// SpeedNames FanSpeed 到中文名称的映射
	SpeedNames = map[xc.FanSpeed]string{
		xc.FanSpeedAuto:   "自动",
		xc.FanSpeedLow:    "低",
		xc.FanSpeedMedium: "中",
		xc.FanSpeedHigh:   "高",
		xc.FanSpeedSilent: "静音",
	}

	// SwingMap 摆风字符串到 SwingMode 的映射
	SwingMap = map[string]xc.SwingMode{
		"off":        xc.SwingModeOff,
		"vertical":   xc.SwingModeVertical,
		"horizontal": xc.SwingModeHorizontal,
		"both":       xc.SwingModeBoth,
	}

	// SwingNames SwingMode 到中文名称的映射
	SwingNames = map[xc.SwingMode]string{
		xc.SwingModeOff:        "关闭",
		xc.SwingModeVertical:   "上下摆风",
		xc.SwingModeHorizontal: "左右摆风",
		xc.SwingModeBoth:       "全方位摆风",
	}
)

// ========== 参数解析辅助函数 ==========

// ParseMode 解析模式字符串
func ParseMode(s string) (xc.OperationalMode, error) {
	mode, ok := ModeMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid mode: %s (valid: cool/heat/auto/dry/fan)", s)
	}
	return mode, nil
}

// ParseFanSpeed 解析风速字符串
func ParseFanSpeed(s string) (xc.FanSpeed, error) {
	speed, ok := SpeedMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("invalid fan speed: %s (valid: auto/low/medium/high/silent)", s)
	}
	return speed, nil
}

// ParseSwingMode 解析摆风模式字符串
func ParseSwingMode(s string) (xc.SwingMode, error) {
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
