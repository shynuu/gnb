package forge

import (
	"encoding/hex"
	"fmt"
	"free5gc/src/gnb/helper"
	"free5gc/src/gnb/uee"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Ping function
func Ping(destination string, pduSession *helper.PDUSession) (err error) {

	// Send the Dummy Packet with the ICMP Request
	tt, b, err := forgeICMP(pduSession.IPv4, destination)
	if err != nil {
		err = fmt.Errorf("Error sending the Packet")
		return
	}
	sendICMP(pduSession.Socket, tt, b)
	time.Sleep(1 * time.Second)
	return
}

func sendICMP(upfConn *net.UDPConn, tt []byte, b []byte) (err error) {

	_, err = upfConn.Write(append(tt, b...))
	if err != nil {
		err = fmt.Errorf("Error sending the Packet")
		return
	}
	return

}

func forgeICMP(source string, destination string) (tt []byte, b []byte, err error) {

	// Forge the GTP-U Header
	gtpHdr, err := hex.DecodeString("32ff00340000000100000000")
	if err != nil {
		err = fmt.Errorf("Error forging GTP Header")
		return
	}

	// Convert the ICMP payload data to Binary (Byte Array)
	icmpData, err := hex.DecodeString("8c870d0000000000101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")
	if err != nil {
		err = fmt.Errorf("Error forging ICMP Data")
		return
	}

	// Forge ICMP IP Header
	ipv4hdr := ipv4.Header{
		Version:  4,
		Len:      20,
		Protocol: 1,
		Flags:    0,
		TotalLen: 48,
		TTL:      64,
		Src:      net.ParseIP(source).To4(),
		Dst:      net.ParseIP(destination).To4(),
		ID:       1,
	}
	checksum := uee.CalculateIpv4HeaderChecksum(&ipv4hdr)
	ipv4hdr.Checksum = int(checksum)

	// Encoding header into Binary values (Byte Array)
	v4HdrBuf, err := ipv4hdr.Marshal()
	if err != nil {
		err = fmt.Errorf("Error encoding IP header")
		return
	}
	// Concatenate the GTP-U and the IP Header
	tt = append(gtpHdr, v4HdrBuf...)

	// Forge the ICMP Payload
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: 12394, Seq: 1,
			Data: icmpData,
		},
	}

	// Encoding the ICMP Packet into Binary values (Byte Array)
	b, err = m.Marshal(nil)
	if err != nil {
		err = fmt.Errorf("Error encoding ICMP header")
		return
	}
	b[2] = 0xaf
	b[3] = 0x88
	return
}
