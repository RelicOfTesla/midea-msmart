# msmart

A Go library for local control of Midea (and associated brands) smart air conditioners. Designed for ease of integration with minimal dependencies.

> This is a Go translation of [msmart-ng](https://github.com/mill1000/midea-msmart) Python library.
> 
> **Note:** This project is developed using vibe coding.

## Supported Devices

This library supports air conditioners from Midea and several associated brands that use the following Android apps or their iOS equivalents:

* Artic King (com.arcticking.ac)
* Midea Air (com.midea.aircondition.obm)
* NetHome Plus (com.midea.aircondition)
* SmartHome/MSmartHome (com.midea.ai.overseas)
* Toshiba AC NA (com.midea.toshiba)
* 美的美居 (com.midea.ai.appliances)

__Note: Only air conditioners (type 0xAC and 0xCC) are supported.__

## Note On Cloud Usage

This library works locally. No internet connection is required to control your device.

_However_, for newer "V3" devices, the Midea Cloud is used to acquire a token & key for device authentication. Once retrieved and saved, no further cloud connection is required. Devices are not linked to the library's built-in accounts and concerned users may supply their own account credentials if they prefer.

## Features

### Simple API

```go
package main

import (
    "context"
    "fmt"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
)

func main() {
    // Build device (ip, port, deviceID, options...)
    device := ac.NewAirConditioner(
        "10.100.1.140",  // IP address
        6444,            // Port
        15393162840672,  // Device ID
        msmart.WithName("Living Room AC"),
    )

    // Get capabilities
    ctx := context.Background()
    if err := device.GetCapabilities(ctx); err != nil {
        panic(err)
    }

    // Get current state
    if err := device.Refresh(ctx); err != nil {
        panic(err)
    }

    power := device.PowerState()
    if power != nil {
        fmt.Printf("Power: %v\n", *power)
    }
    fmt.Printf("Temperature: %.1f°C\n", device.TargetTemperature())
}
```

### Device Discovery

Automatically discover devices on the local network or an individual device by IP or hostname.

```go
package main

import (
    "context"
    "fmt"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

func main() {
    ctx := context.Background()

    // Discover all devices on the network (pass nil for default config)
    devices, err := msmart.Discover(ctx, nil)
    if err != nil {
        panic(err)
    }

    for _, device := range devices {
        fmt.Printf("Found device: %s (%s)\n", device.GetName(), device.GetIP())
    }

    // Discover a single device by IP
    device, err := msmart.DiscoverSingle(ctx, "10.100.1.140", nil)
    if err != nil {
        panic(err)
    }

    if device != nil {
        fmt.Printf("Device: %s\n", device.GetName())
    }
}
```

__Note: V3 devices are automatically authenticated via the NetHome Plus cloud.__

### Minimal Dependencies

Built using Go standard library with minimal external dependencies.

### Code Quality

- Fully typed for clarity
- Context support for cancellation and timeout
- Error handling following Go conventions
- Well-structured codebase

## Installing

```shell
go get github.com/RelicOfTesla/midea-msmart/msmart
```

## Usage

### Device Discovery

Use the `Discover` function to find devices on your local network:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

func main() {
    ctx := context.Background()

    // Discover all devices (pass nil for default config)
    devices, err := msmart.Discover(ctx, nil)
    if err != nil {
        panic(err)
    }

    for _, device := range devices {
        data := map[string]interface{}{
            "ip":       device.GetIP(),
            "port":     device.GetPort(),
            "id":       device.GetID(),
            "online":   device.IsOnline(),
            "supported": device.IsSupported(),
            "type":     device.GetType(),
            "name":     device.GetName(),
            "sn":       device.GetSN(),
        }

        jsonData, _ := json.MarshalIndent(data, "", "  ")
        fmt.Println(string(jsonData))
    }
}
```

Ensure the device type is `0xAC` (172) or `0xCC` (204) and the `supported` property is true.

Save the device ID, IP address, and port. Version 3 devices will also require the `token` and `key` fields to control the device.

#### Warning: V3 Device Users

For V3 devices, it's highly recommended to save your token and key values in a secure place. In the event that the cloud becomes unavailable, having these on hand will allow you to continue controlling your device locally.

### Controlling Devices

```go
package main

import (
    "context"
    "encoding/hex"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
)

func main() {
    ctx := context.Background()

    // For V3 devices, you need token and key from discovery or cloud
    token, _ := hex.DecodeString("YOUR_TOKEN_HEX_STRING")
    key, _ := hex.DecodeString("YOUR_KEY_HEX_STRING")

    // Create device (ip, port, deviceID, options...)
    device := ac.NewAirConditioner(
        "10.100.1.140",  // IP address
        6444,            // Port
        15393162840672,  // Device ID
        msmart.WithVersion(3),  // For V3 devices
    )

    // Authenticate V3 device (required before control)
    if err := device.Authenticate(token, key); err != nil {
        panic(err)
    }

    // Get current state
    if err := device.Refresh(ctx); err != nil {
        panic(err)
    }

    // Control the device
    device.SetPowerState(true)
    device.SetTargetTemperature(20.5)
    device.SetOperationalMode(ac.OperationalModeCool)
    device.SetFanSpeed(ac.FanSpeedHigh)

    // Apply changes
    if err := device.Apply(ctx); err != nil {
        panic(err)
    }
}
```

### Querying Device State

```go
package main

import (
    "context"
    "fmt"
    msmart "github.com/RelicOfTesla/midea-msmart/msmart"
    "github.com/RelicOfTesla/midea-msmart/msmart/device/ac"
)

func main() {
    ctx := context.Background()

    // Create device (ip, port, deviceID, options...)
    device := ac.NewAirConditioner(
        "10.100.1.140",  // IP address
        6444,            // Port
        15393162840672,  // Device ID
        msmart.WithVersion(3),  // For V3 devices
    )

    // Refresh device state
    if err := device.Refresh(ctx); err != nil {
        panic(err)
    }

    // Print current state
    if power := device.PowerState(); power != nil {
        fmt.Printf("Power: %v\n", *power)
    }
    fmt.Printf("Target Temperature: %.1f°C\n", device.TargetTemperature())
    if indoor := device.IndoorTemperature(); indoor != nil {
        fmt.Printf("Indoor Temperature: %.1f°C\n", *indoor)
    }
    fmt.Printf("Operational Mode: %v\n", device.OperationalMode())
    fmt.Printf("Fan Speed: %v\n", device.FanSpeed())
}
```

## Troubleshooting

* If devices are not being discovered, ensure your devices are on the same subnet as your computer.
* If a cloud connection cannot be made, try using credentials from a different region.

## Gratitude

This project is a Go translation of [mill1000/midea-msmart](https://github.com/mill1000/midea-msmart), which builds upon the work of:

* [mac-zhou/midea-msmart](https://github.com/mac-zhou/midea-msmart)
* [dudanov/MideaUART](https://github.com/dudanov/MideaUART)
* [NeoAcheron/midea-ac-py](https://github.com/NeoAcheron/midea-ac-py)
* [andersonshatch/midea-ac-py](https://github.com/andersonshatch/midea-ac-py)
* [yitsushi/midea-air-condition](https://github.com/yitsushi/midea-air-condition)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
