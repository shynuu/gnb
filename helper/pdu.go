package helper

import "net"

// PDUSession helper
type PDUSession struct {
	IPv4   string
	Socket *net.UDPConn
}

// PDUSessionList helper, store all the current running PDU session
var PDUSessionList []*PDUSession
