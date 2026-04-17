# midea - 美的空调控制 CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个用于控制美的空调的命令行工具，支持局域网设备发现、状态查询、远程控制等功能。

**⚠️ 这是有 AI vibe coding 的项目 - 由 AI 辅助开发和维护。**

## Supported Devices

This library supports air conditioners from Midea and several associated brands that use the following Android apps or their iOS equivalents:

* Artic King (com.arcticking.ac)
* Midea Air (com.midea.aircondition.obm)
* NetHome Plus (com.midea.aircondition)
* SmartHome/MSmartHome (com.midea.ai.overseas)
* Toshiba AC NA (com.midea.toshiba)
* 美的美居 (com.midea.ai.appliances)

__Note: Only air conditioners (type 0xAC and 0xCC) are supported.__

## ✨ 功能特性

- 🔍 **设备发现** - 自动发现局域网内的美的空调设备
- 📋 **设备管理** - 列出、绑定、解绑设备
- 📊 **状态查询** - 查询设备状态、能耗信息、设备能力
- 🎛️ **远程控制** - 开关机、设置温度、模式、风速、摆风
- 🔐 **V3 设备支持** - 支持 V3 设备的自动认证和 token 获取
- 💾 **配置管理** - 本地配置文件，支持多设备管理
- 🌍 **多地区支持** - 支持中国、美国、德国、韩国等地区
- 📖 **详细日志** - 支持 verbose 模式，方便调试

## 📦 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/RelicOfTesla/midea-msmart.git
cd midea-msmart

# 编译
go build -o midea ./cmd

# 移动到 PATH（可选）
sudo mv midea /usr/local/bin/
```


## 🚀 快速开始

### 1. 发现设备

```bash
# 发现局域网内的所有设备
midea discover

# 发现指定 IP 的设备
midea discover 192.168.1.60

# 发现设备并自动获取 V3 设备的 token
midea discover --auto-connect

# 使用美的账号密码发现设备(可选)
midea discover --auto-connect --account your@email.com --password yourpass
```

### 2. 绑定设备

```bash
# 绑定设备并命名为"客厅"
midea bind 192.168.1.60 -n 客厅

# 列出已绑定的设备
midea list
```

### 3. 控制设备

```bash
# 查询状态
midea status 客厅

# 开机
midea on 客厅

# 设置温度
midea temp 客厅 26

# 设置模式
midea mode 客厅 cool

# 设置风速
midea fan 客厅 high

# 设置摆风
midea swing 客厅 vertical
```

## 📖 命令参考

### 常用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `discover` | 发现设备 | `midea discover` |
| `list` | 列出设备 | `midea list` |
| `bind` | 绑定设备 | `midea bind 192.168.1.60 -n 客厅` |
| `status` | 查询状态 | `midea status 客厅` |
| `on/off` | 开关机 | `midea on 客厅` |
| `temp` | 设置温度 | `midea temp 客厅 26` |
| `mode` | 设置模式 | `midea mode 客厅 cool` |
| `fan` | 设置风速 | `midea fan 客厅 high` |
| `swing` | 设置摆风 | `midea swing 客厅 vertical` |
| `set` | 多参数设置 | `midea set 客厅 --temp 26 --mode cool` |
| `query` | 查询属性 | `midea query 客厅 temp` |

### 参数说明

**模式**: `cool`(制冷), `heat`(制热), `auto`(自动), `dry`(除湿), `fan`(送风)

**风速**: `auto`(自动), `low`(低速), `medium`(中速), `high`(高速), `silent`(静音)

**摆风**: `off`(关闭), `vertical`(上下), `horizontal`(左右), `both`(全方位)

**温度范围**: 16-30°C

### 全局选项

```bash
-v, --verbose           # 显示详细日志
--region <地区>         # 云端服务地区 (DE, KR, US), 默认: US
--device_type <类型>    # 设备类型: AC (空调), CC (商业空调), 默认: AC
```

## ⚙️ 配置

配置文件位置：
- 当前目录的 `midea.json`（优先）
- `~/.config/midea/config.json`（默认）

## 📝 自动化示例

```bash
# cron 定时任务
0 7 * * * /usr/local/bin/midea on 客厅    # 每天 7:00 开启
0 23 * * * /usr/local/bin/midea off 客厅  # 每天 23:00 关闭
```


## 🔧 故障排除

### 设备发现失败

1. 确保设备和电脑在同一个局域网
2. 检查防火墙设置，确保 UDP 端口 6445 未被阻止
3. 尝试使用 `--verbose` 模式查看详细日志

### V3 设备认证失败

1. 尝试使用 `--auto` / `--auto-connect` 参数
2. 确保使用正确的美的账号和密码
3. 确保设备已绑定到账号

### 连接超时

1. 检查设备是否在线
2. 检查网络连接
3. 尝试使用 `-v` 参数配合AI查看详细日志

## 📚 作为库使用

```go
package main

import (
    "context"
    "encoding/hex"
    "fmt"
    "log"
    "time"

    "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    token, _ := hex.DecodeString("your_token_hex")
    key, _ := hex.DecodeString("your_key_hex")

    device := ac.NewAirConditioner(
        "192.168.1.60",
        6444,
        "123456789",
        msmart.WithVersion(3),
        msmart.WithTokenKey(token, key),
    )

    if err := device.Refresh(ctx); err != nil {
        log.Fatal(err)
    }

    if device.PowerState() {
        fmt.Printf("温度: %.1f°C\n", device.TargetTemperature())
    }

    device.SetTargetTemperature(26)
    if err := device.Apply(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## 🙏 致谢

本项目是 [msmart-ng](https://github.com/mill1000/midea-msmart) Python 库的 Go 语言实现。

感谢以下项目和资源：
- [msmart-ng](https://github.com/mill1000/midea-msmart) - 原始 Python 实现

