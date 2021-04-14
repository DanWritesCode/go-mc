# Go-MC
![Version](https://img.shields.io/badge/Minecraft-1.16.5-blue.svg)
![Protocol](https://img.shields.io/badge/Protocol-754-blue.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Tnze/go-mc.svg)](https://pkg.go.dev/github.com/Tnze/go-mc)
[![Go Report Card](https://goreportcard.com/badge/github.com/Tnze/go-mc)](https://goreportcard.com/report/github.com/Tnze/go-mc)
[![Build Status](https://travis-ci.org/Tnze/go-mc.svg?branch=master)](https://travis-ci.org/Tnze/go-mc)

Requires Go version: 1.16

There's some library in Go support you to create your Minecraft client or server.  

### Edit by Dan: This fork adds support for joining Minecraft servers through proxies. It also refactors some classes to use the ioutil package.

- [x] Chat Message (Support Json or old `§`)
- [x] NBT (Based on reflection)
- [x] Yggdrasil
- [x] Realms Server
- [x] RCON protocol (Server & Client)
- [x] Saves decoding & encoding
- [x] Minecraft network protocol
- [x] Robot player framework

> `1.13.2` version is at [gomcbot](https://github.com/Tnze/gomcbot).

## Getting start
After you install golang:  
To get the latest version: `go get github.com/Tnze/go-mc@master`  
To get old versions (e.g. 1.14.3): `go get github.com/Tnze/go-mc@v1.14.3`

First, you might have a try of the simple examples. It's a good start.

### Run Examples

- Run `go run github.com/Tnze/go-mc/cmd/mcping localhost` to ping and list the localhost mc server.  
- Run `go run github.com/Tnze/go-mc/cmd/daze` to join the local server at *localhost:25565* as Steve on the offline mode.

### Basic Usage

One of the most useful functions of this lib is that it implements the network communication protocol of minecraft. It allows you to construct, send, receive, and parse network packets. All of them are encapsulated in `go-mc/net` and `go-mc/net/packet`.

```go
import "github.com/Tnze/go-mc/net"
import pk "github.com/Tnze/go-mc/net/packet"
```

It's very easy to create a packet. For example, after any client connected the server, it sends a [Handshake Packet](https://wiki.vg/Protocol#Handshake). You can create this package with the following code:


```go
p := pk.Marshal(
    0x00,                       // Handshake packet ID
    pk.VarInt(ProtocolVersion), // Protocol version
    pk.String("localhost"),     // Server's address
    pk.UnsignedShort(25565),    // Server's port
    pk.Byte(1),                 // 1 for status ping, 2 for login
)
```

Then you can send it to server using `conn.WritePacket(p)`. The `conn` is a `net.Conn` which is returned by `net.Dial()`. And don't forget to handle the error.^_^

Receiving packet is quite easy too. To read a packet, call `p.Scan()` like this:


```go
var (
    x, y, z    pk.Double
    yaw, pitch pk.Float
    flags      pk.Byte
    TeleportID pk.VarInt
)

err := p.Scan(&x, &y, &z, &yaw, &pitch, &flags, &TeleportID)
if err != nil {
    return err
}
```

### Advanced usage

Sometimes you are handling packet like this:

| **Field Name** |     Field Type      | **Notes**                                 |
| :------------: | :-----------------: | :---------------------------------------- |
|  World Count   |       VarInt        | Size of the following array.              |
|  World Names   | Array of Identifier | Identifiers for all worlds on the server. |

That is, the first field is an integer type and the second field is an array (a `[]string` in this case). The integer represents the length of array.

Traditionally, you can use the following method to read such a field:

```go
r := bytes.Reader(p.Data)
// Read WorldCount
var WorldCount pk.VarInt
if err := WorldCount.ReadFrom(r); err != nil {
    return err
}
// Read WorldNames
WorldNames := make([]pk.Identifier, WorldCount)
for i := 0; i < int(WorldCount); i++ {
    if err := WorldNames[i].ReadFrom(r); err != nil {
        return err
    }
}
```

But this is tediously long an not compatible with `p.Scan()` method.

In the latest version, two new types is added: `pk.Ary` and `pk.Opt`. Dedicated to handling "Array of ...." and "Optional ...." fields.

```go
var WorldCount pk.VarInt
var WorldNames = []pk.Identifier{}
if err := p.Scan(&WorldCount, pk.Ary{&WorldCount, &WorldNames}); err != nil {
    return err
}
```



---

As the `go-mc/net` package implements the minecraft network protocol, there is no update between the versions at this level. So net package actually supports any version. It's just that the ID and content of the package are different between different versions.

Originally it's all right to write a bot with only `go-mc/net` package, but considering that the process of handshake, login and encryption is not difficult but complicated, I have implemented it in `go-mc/bot` package, which is **not cross-versions**. You may use it directly or as a reference for your own implementation.

Now, go and have a look at the examples!
