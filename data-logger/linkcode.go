package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
)

/**
 * Values for the .flags bit field
 *
 * 1....... = LAN Transport mode
 * ..cc.... = Encoding
 * ....ff.. = Framing
 * ......ee = Encryption
 */
const (
	FlagTLSNone  = 0x00
	FlagTLSKudzu = 0x01
	FlagTLSAny   = 0x02

	FlagFramingRaw      = 0x00
	FlagFramingHTTPRaw  = 0x04
	FlagFramingHTTPJSON = 0x08

	FlagEncodingCayenne = 0x10
	FlagEncodingJSON    = 0x20

	FlagTransportLAN = 0x80
)

type ConnectionInfo struct {
	Port      int
	Addresses []net.IP
	Path      string
	LanMode   bool
}

func CreateConnectionInfo(port int, path string) *ConnectionInfo {
	var inv ConnectionInfo
	inv.Port = port
	inv.Path = path
	inv.LanMode = true

	// Find all the public IP addresses
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("Could not enumerate interfaces: %+v\n", err.Error()))
		return &inv
	}
	for _, i := range ifaces {
		if (i.Flags & net.FlagUp) == 0 {
			continue
		}
		if (i.Flags & net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			fmt.Print(fmt.Errorf("Could not enumerate addresses of %s: %+v\n", i.Name, err.Error()))
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if ipv4 := v.IP.To4(); ipv4 != nil {
					inv.Addresses = append(inv.Addresses, ipv4)
				}
			}
		}
	}

	return &inv
}

func (inv *ConnectionInfo) Serialize() []byte {
	iHdr := 0
	iHdrLen := 4
	iPort := iHdr + iHdrLen
	iPortLen := 2
	iAddrs := iPort + iPortLen
	iAddrsLen := len(inv.Addresses) * 4
	iPath := iAddrs + iAddrsLen
	iPathLen := len(inv.Path)
	ret := make([]byte, iPath+iPathLen)

	ret[iHdr] = 0x01 // Protocol version
	ret[iHdr+1] = byte(len(inv.Addresses))
	ret[iHdr+2] = byte(len(inv.Path))
	ret[iHdr+3] = FlagTLSNone | FlagFramingHTTPRaw | FlagEncodingJSON | FlagTransportLAN

	binary.LittleEndian.PutUint16(ret[iPort:], uint16(inv.Port))
	for i, addr := range inv.Addresses {
		copy(ret[iAddrs+i*4:], addr)
	}
	copy(ret[iPath:], []byte(inv.Path))

	return ret
}

func (inv *ConnectionInfo) String() string {
	return base64.StdEncoding.EncodeToString(inv.Serialize())
}
