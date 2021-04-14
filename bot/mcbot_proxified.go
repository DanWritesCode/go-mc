// Package bot implements a simple Minecraft client that can join a server
// or just ping it for getting information.
//
// Runnable example could be found at examples/ .
package bot

import (
	"net"
	"strconv"
	"strings"
	"encoding/base64"
	//"io"
	//"io/ioutil"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	mcnet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
)



// JoinServer connect a Minecraft server for playing the game.
// Using roughly the same way to parse address as minecraft.
func (c *Client) JoinServerWithProxy(addr string, proxy string, proxyLogin string) (err error) {
	return c.joinWithProxy(&net.Dialer{}, addr, proxy, proxyLogin)
}

func (c *Client) joinWithProxy(d *net.Dialer, addr string, proxy string, proxyLogin string) error {
	const Handshake = 0x00
	addrSrv, err := parseAddress(d.Resolver, addr)
	if err != nil {
		return LoginErr{"resolved address", err}
	}

	// Split Host and Port
	host, portStr, err := net.SplitHostPort(addrSrv)
	if err != nil {
		return LoginErr{"split address", err}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return LoginErr{"parse port", err}
	}

	// Dial connection
	iConn, err := net.Dial("tcp", proxy)
	if err != nil {
		return LoginErr{"connect proxy server", err}
	}

	data := "CONNECT "+addrSrv+" HTTP/1.1\r\nHost: " + addrSrv + "\r\n"
	if proxyLogin != "" {
		proxyLogin = base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(proxyLogin)))
		data = data + "Proxy-Authorization: Basic "+proxyLogin+"\r\n"
	}
	data = data + "\r\n"
	_, err = iConn.Write([]byte(data))
	if err != nil {
		return err
	}

	// read and ignore value
	iConn.Read(make([]byte, 2048))
	//io.Copy(ioutil.Discard, iConn)

	c.Conn = mcnet.WrapConn(iConn)

	// Handshake
	err = c.Conn.WritePacket(pk.Marshal(
		Handshake,
		pk.VarInt(ProtocolVersion), // Protocol version
		pk.String(host),            // Host
		pk.UnsignedShort(port),     // Port
		pk.Byte(2),
	))
	if err != nil {
		return LoginErr{"handshake", err}
	}

	// Login Start
	err = c.Conn.WritePacket(pk.Marshal(
		packetid.LoginStart,
		pk.String(c.Auth.Name),
	))
	if err != nil {
		return LoginErr{"login start", err}
	}

	for {
		//Receive Packet
		var p pk.Packet
		if err = c.Conn.ReadPacket(&p); err != nil {
			return LoginErr{"receive packet", err}
		}

		//Handle Packet
		switch p.ID {
		case packetid.Disconnect: //Disconnect
			var reason chat.Message
			err = p.Scan(&reason)
			if err != nil {
				return LoginErr{"disconnect", err}
			}
			return LoginErr{"disconnect", DisconnectErr(reason)}

		case packetid.EncryptionBeginClientbound: //Encryption Request
			if err := handleEncryptionRequest(c, p); err != nil {
				return LoginErr{"encryption", err}
			}

		case packetid.Success: //Login Success
			err := p.Scan(
				(*pk.UUID)(&c.UUID),
				(*pk.String)(&c.Name),
			)
			if err != nil {
				return LoginErr{"login success", err}
			}
			return nil

		case packetid.Compress: //Set Compression
			var threshold pk.VarInt
			if err := p.Scan(&threshold); err != nil {
				return LoginErr{"compression", err}
			}
			c.Conn.SetThreshold(int(threshold))

		case packetid.LoginPluginRequest: //Login Plugin Request
			// TODO: Handle login plugin request
		}
	}
}
