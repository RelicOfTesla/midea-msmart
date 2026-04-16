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
midea status <name|id> [--auto] [--json|-j] [--capabilities [FILE]] [--energy]
```

- `--auto`: 自动发现设备并获取 token
- `--json`, `-j`: 以 JSON 格式输出状态（便于脚本处理）
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
- `--auto`: 自动发现设备并获取 token

### 高级命令

```
midea - 美的空调控制 CLI v1.0.0

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
  midea status 客厅 --json           # 以 JSON 格式输出状态（便于脚本处理）
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
```

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
      "key": "yyy",
      "local_key": "01a74aca...",
      "local_key_expire": "2026-04-17T01:34:12Z"
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

### LocalKey 存储与复用

对于 V3 设备，CLI 会自动缓存 `local_key` 和过期时间：

- **存储**：在认证成功后，自动保存 `local_key` 和过期时间到配置文件
- **复用**：下次连接时，如果 `local_key` 有效且未过期，直接使用缓存的 key，避免重复认证
- **过期处理**：如果 `local_key` 已过期，会自动重新认证并更新缓存

这样可以显著减少 V3 设备的连接时间。

## 🛠️ 开发


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

## 🙏 致谢

本项目是 [msmart-ng](https://github.com/mill1000/midea-msmart) Python 库的 Go 语言实现。

感谢以下项目和资源：
- [msmart-ng](https://github.com/mill1000/midea-msmart) - 原始 Python 实现

