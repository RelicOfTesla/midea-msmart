# midea - 美的空调控制 CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个用于控制美的空调的命令行工具，支持局域网设备发现、状态查询、远程控制等功能。

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

# 使用美的账号密码发现设备
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

## 📖 命令详解

### 全局选项

```
-v, --verbose           显示详细调试日志
--region <地区>         云端服务地区 (DE, KR, US), 默认: US
--device_type <类型>    设备类型: AC (空调), CC (商业空调), 默认: AC
```

### 设备管理命令

#### discover - 发现设备

```bash
midea discover [<host>] [--auto-connect|-a] [--count <数量>] [--account <账号> --password <密码>]
```

- `<host>`: 可选，指定目标设备 IP（发现单个设备）
- `--auto-connect`: 自动连接并获取 V3 设备的 token
- `--count`: 广播包数量（默认: 3）
- `--account/--password`: 美的账号密码（V3 设备认证需要）

#### list - 列出设备

```bash
midea list
```

列出已保存的所有设备。

#### bind - 绑定设备

```bash
midea bind <id|sn|ip> -n <名称>
```

为设备绑定一个别名，方便后续使用。

#### unbind - 解绑设备

```bash
midea unbind <name|id>
```

解绑指定的设备。

### 状态查询命令

#### status - 查询设备状态

```bash
midea status <name|id> [--auto] [--capabilities [FILE]] [--energy]
```

- `--auto`: 自动发现设备并获取 token
- `--capabilities`: 显示设备能力信息
- `--capabilities FILE`: 将设备能力写入 YAML 文件
- `--energy`: 显示能耗信息

### 控制命令

#### on - 开机

```bash
midea on <name|id>
```

#### off - 关机

```bash
midea off <name|id>
```

#### temp - 设置温度

```bash
midea temp <name|id> <温度>
```

温度范围: 16-30°C

#### mode - 设置运行模式

```bash
midea mode <name|id> <模式>
```

模式选项:
- `cool` - 制冷
- `heat` - 制热
- `auto` - 自动
- `dry` - 除湿
- `fan` - 送风

#### fan - 设置风速

```bash
midea fan <name|id> <风速>
```

风速选项:
- `auto` - 自动
- `low` - 低速
- `medium` - 中速
- `high` - 高速
- `silent` - 静音

#### swing - 设置摆风模式

```bash
midea swing <name|id> <模式>
```

摆风模式:
- `off` - 关闭
- `vertical` - 上下摆风
- `horizontal` - 左右摆风
- `both` - 全方位摆风

#### set - 多参数设置

```bash
midea set <name|id> [选项]
```

选项:
- `--temp <温度>` - 设置温度
- `--mode <模式>` - 设置运行模式
- `--fan <风速>` - 设置风速
- `--swing <模式>` - 设置摆风
- `--power <on|off>` - 设置电源

示例:
```bash
# 一次设置多个属性
midea set 客厅 --temp 26 --mode cool --fan high

# 开机并设置温度
midea set 客厅 --power on --temp 24
```

#### query - 查询设备属性

```bash
midea query <name|id> [key] [--all] [--auto]
```

- `key`: 属性名称（如: temp, mode, fan, swing, power）
- `--all`: 显示所有属性（默认）
- `--auto`: 自动发现设备并获取 token

### 高级命令

#### download - 下载设备协议和插件

```bash
midea download <host> [--account <账号> --password <密码>]
```

下载设备的 Lua 协议和插件（需要美的账号密码）。

## ⚙️ 配置

配置文件位置：
- 当前目录的 `midea.json`（优先）
- `~/.config/midea/config.json`（默认）

配置文件格式：

```json
{
  "devices": {
    "客厅": {
      "id": "192.168.1.60",
      "name": "客厅空调",
      "type": "AC",
      "token": "xxx",
      "key": "yyy"
    },
    "卧室": {
      "id": "192.168.1.61",
      "name": "卧室空调",
      "type": "AC",
      "token": "xxx",
      "key": "yyy"
    }
  }
}
```

## 🛠️ 开发

### 项目结构

```
midea-msmart/
├── cmd/                    # 命令行工具
│   └── main.go            # 主入口
├── msmart/                 # 核心库
│   ├── cloud.go           # 云端 API
│   ├── const.go           # 常量定义
│   ├── crc8.go            # CRC8 校验
│   ├── device.go          # 设备基类
│   ├── frame.go           # 协议帧
│   ├── lan.go             # 局域网通信
│   ├── utils.go           # 工具函数
│   └── device/            # 设备类型实现
│       ├── ac/            # 空调设备
│       └── cc/            # 商业空调
├── go.mod                  # Go 模块定义
├── go.sum                  # 依赖校验
└── README.md              # 本文件
```

### 从源码运行

```bash
# 直接运行
go run ./cmd [命令]

# 编译
go build -o midea ./cmd
```

### 测试

```bash
# 运行测试
go test ./...

# 运行特定测试
go test ./msmart/... -v
```

## 📝 示例

### 自动化脚本

```bash
#!/bin/bash
# 每天早上 7 点自动打开客厅空调，设置为制冷模式，温度 24°C

midea on 客厅
midea mode 客厅 cool
midea temp 客厅 24
```

### 配合 cron 定时任务

```cron
# 每天 7:00 开启客厅空调
0 7 * * * /usr/local/bin/midea on 客厅

# 每天 23:00 关闭客厅空调
0 23 * * * /usr/local/bin/midea off 客厅
```

### 查询设备能力

```bash
# 查询设备能力并保存到文件
midea status 客厅 --capabilities capabilities.yaml

# 查看设备能力
cat capabilities.yaml
```

## 🔧 故障排除

### 设备发现失败

1. 确保设备和电脑在同一个局域网
2. 检查防火墙设置，确保 UDP 端口 6445 未被阻止
3. 尝试使用 `--verbose` 模式查看详细日志

### V3 设备认证失败

1. 尝试使用 `--auto-connect` / `--auto` 参数号和密码
2. 确保使用正确的美的账
3. 确保设备已绑定到账号

### 连接超时

1. 检查设备是否在线
2. 检查网络连接
3. 尝试使用 `-v` 参数查看详细日志

## 🙏 致谢

本项目是 [msmart-ng](https://github.com/mill1000/midea-msmart) Python 库的 Go 语言实现。

感谢以下项目和资源：
- [msmart-ng](https://github.com/mill1000/midea-msmart) - 原始 Python 实现
- [midea-ac-lib](https://github.com/regevbr/midea-ac-lib) - Midea AC 协议参考

