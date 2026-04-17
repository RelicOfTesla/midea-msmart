# msmart

Go 语言实现的美的（及关联品牌）智能空调本地控制库，设计简洁，依赖极少。

> 这是 [msmart-ng](https://github.com/mill1000/midea-msmart) Python 库的 Go 语言移植版本。

**⚠️ 这是有 AI vibe coding 的项目 - 由 AI 辅助开发和维护。**

## 支持的设备

支持使用以下 Android 应用（或 iOS 等效版本）的美的及关联品牌空调：

* Artic King (com.arcticking.ac)
* Midea Air (com.midea.aircondition.obm)
* NetHome Plus (com.midea.aircondition)
* SmartHome/MSmartHome (com.midea.ai.overseas)
* Toshiba AC NA (com.midea.toshiba)
* 美的美居 (com.midea.ai.appliances)

**注意：仅支持空调设备（类型 0xAC 和 0xCC）。**

## 关于云服务

本库通过本地网络控制设备，无需互联网连接。

但对于较新的 "V3" 设备，需要从美的云端获取 token 和 key 用于设备认证。获取并保存后，无需再次连接云端。设备不会绑定到库的内置账号，用户也可使用自己的账号凭证。

## 安装

```shell
go get github.com/RelicOfTesla/midea-msmart/msmart
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
)

func main() {
    // 创建设备（IP, 端口, 设备ID, 选项...）
    device := ac.NewAirConditioner(
        "10.100.1.140",  // IP 地址
        6444,            // 端口
        "15393162840672",  // 设备 ID（字符串）
        msmart.WithName("客厅空调"),
    )

    ctx := context.Background()
    
    // 获取设备能力
    if err := device.GetCapabilities(ctx); err != nil {
        panic(err)
    }

    // 刷新设备状态
    if err := device.Refresh(ctx); err != nil {
        panic(err)
    }

    fmt.Printf("电源: %v\n", device.PowerState())
    fmt.Printf("温度: %.1f°C\n", device.TargetTemperature())
}
```

## 设备发现

自动发现本地网络中的设备：

```go
ctx := context.Background()

// 发现所有设备
devices, err := msmart.Discover(ctx, nil)
if err != nil {
    panic(err)
}

for _, device := range devices {
    fmt.Printf("发现设备: %s (%s)\n", device.GetName(), device.GetIP())
}

// 发现单个设备（按 IP）
device, err := msmart.DiscoverSingle(ctx, "10.100.1.140", nil)
```

**注意：V3 设备会自动通过 NetHome Plus 云端认证。**

## 控制设备

```go
package main

import (
    "context"
    "encoding/hex"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/xc"
)

func main() {
    ctx := context.Background()

    // V3 设备需要 token 和 key
    token, _ := hex.DecodeString("YOUR_TOKEN_HEX_STRING")
    key, _ := hex.DecodeString("YOUR_KEY_HEX_STRING")

    device := ac.NewAirConditioner(
        "10.100.1.140",
        6444,
        "15393162840672",
        msmart.WithVersion(3),
    )

    // V3 设备需要认证
    if err := device.Authenticate(ctx, token, key); err != nil {
        panic(err)
    }

    // 刷新状态
    if err := device.Refresh(ctx); err != nil {
        panic(err)
    }

    // 控制设备
    device.SetPowerState(true)
    device.SetTargetTemperature(20.5)
    device.SetOperationalMode(xc.OperationalModeCool)
    device.SetFanSpeed(xc.FanSpeedHigh)

    // 应用更改
    if err := device.Apply(ctx); err != nil {
        panic(err)
    }
}
```

## 查询设备状态

```go
device := ac.NewAirConditioner("10.100.1.140", 6444, "15393162840672", msmart.WithVersion(3))

if err := device.Refresh(ctx); err != nil {
    panic(err)
}

fmt.Printf("电源: %v\n", device.PowerState())
fmt.Printf("目标温度: %.1f°C\n", device.TargetTemperature())
if indoor := device.IndoorTemperature(); indoor != nil {
    fmt.Printf("室内温度: %.1f°C\n", *indoor)
}
fmt.Printf("运行模式: %v\n", device.OperationalMode())
fmt.Printf("风速: %v\n", device.FanSpeed())
```

## 故障排除

* **设备无法发现**：确保设备与电脑在同一子网
* **无法连接云端**：尝试使用其他地区的账号凭证

## 许可证

MIT License
