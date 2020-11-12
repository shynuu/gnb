package forge

import (
	"encoding/hex"
	"fmt"
	"free5gc/lib/nas"
	"free5gc/lib/nas/nasMessage"
	"free5gc/lib/nas/nasTestpacket"
	"free5gc/lib/nas/nasType"
	"free5gc/lib/nas/security"
	"free5gc/lib/ngap"
	"free5gc/lib/openapi/models"
	"free5gc/src/gnb/context"
	"free5gc/src/gnb/helper"
	"free5gc/src/gnb/interfaces"
	"free5gc/src/gnb/procedures"
	"free5gc/src/gnb/uee"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func ipv4HeaderChecksum(hdr *ipv4.Header) uint32 {
	var Checksum uint32
	Checksum += uint32((hdr.Version<<4|(20>>2&0x0f))<<8 | hdr.TOS)
	Checksum += uint32(hdr.TotalLen)
	Checksum += uint32(hdr.ID)
	Checksum += uint32((hdr.FragOff & 0x1fff) | (int(hdr.Flags) << 13))
	Checksum += uint32((hdr.TTL << 8) | (hdr.Protocol))

	src := hdr.Src.To4()
	Checksum += uint32(src[0])<<8 | uint32(src[1])
	Checksum += uint32(src[2])<<8 | uint32(src[3])
	dst := hdr.Dst.To4()
	Checksum += uint32(dst[0])<<8 | uint32(dst[1])
	Checksum += uint32(dst[2])<<8 | uint32(dst[3])
	return ^(Checksum&0xffff0000>>16 + Checksum&0xffff)
}

// PDUSessionEstablish function
func PDUSessionEstablish(userEquipment *context.UE) (err error) {

	pduSession, err := PDUSessionEstablishment(userEquipment)
	if err != nil {
		err = fmt.Errorf("Error Connecting to UPF")
		return
	}
	helper.PDUSessionList = append(helper.PDUSessionList, pduSession)
	return
}

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

// PDUSessionEstablishment function
func PDUSessionEstablishment(userEquipment *context.UE) (pduSession *helper.PDUSession, err error) {

	var n int
	var sendMsg []byte
	var recvMsg = make([]byte, 2048)

	amfN3IpAddr := context.RAN_Self().AmfInterface.IPv4Addr
	amfN3Port := context.RAN_Self().AmfInterface.Port
	ranN3IpAddr := context.RAN_Self().NGRANInterface.IPv4Addr
	ranN3Port := context.RAN_Self().NGRANInterface.Port

	// RAN connect to AMF
	amfConn, err := interfaces.ConnectToAmf(amfN3IpAddr, ranN3IpAddr, amfN3Port, ranN3Port)
	if err != nil {
		err = fmt.Errorf("Error Connecting to AMF")
		return
	}

	ranGTPIpAddr := context.RAN_Self().GTPInterface.IPv4Addr
	ranGTPPort := context.RAN_Self().GTPInterface.Port
	upfGTPIpAddr := context.RAN_Self().UpfInterface.IPv4Addr
	upfGTPPort := context.RAN_Self().UpfInterface.Port

	// RAN connect to UPF
	upfConn, err := interfaces.ConnectToUpf(ranGTPIpAddr, upfGTPIpAddr, ranGTPPort, upfGTPPort)
	if err != nil {
		err = fmt.Errorf("Error Connecting to UPF")
		return
	}

	// send NGSetupRequest Msg
	sendMsg, err = procedures.GetNGSetupRequest([]byte("\x00\x01\x02"), 24, "free5gc")
	if err != nil {
		err = fmt.Errorf("Error")
		return
	}

	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error")
		return
	}

	// receive NGSetupResponse Msg
	n, err = amfConn.Read(recvMsg)
	if err != nil {
		err = fmt.Errorf("Error")
		return
	}

	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		err = fmt.Errorf("Error Decoded ngap")
		return
	}

	// New UE
	// ue := test.NewRanUeContext("imsi-2089300007487", 1, security.AlgCiphering128NEA2, security.AlgIntegrity128NIA2)
	ue := uee.NewRanUeContext(userEquipment.Supi, 1, security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2)
	ue.AmfUeNgapId = 1
	ue.AuthenticationSubs = uee.GetAuthSubscription(context.RAN_Self().Security.K, context.RAN_Self().Security.OPC, context.RAN_Self().Security.OP)

	// send InitialUeMessage(Registration Request)(imsi-2089300007487)
	mobileIdentity5GS := nasType.MobileIdentity5GS{
		Len:    12, // suci
		Buffer: []uint8{0x01, 0x02, 0xf8, 0x39, 0xf0, 0xff, 0x00, 0x00, 0x00, 0x00, 0x47, 0x78},
	}

	ueSecurityCapability := ue.GetUESecurityCapability()
	registrationRequest := nasTestpacket.GetRegistrationRequest(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, ueSecurityCapability, nil, nil, nil)
	sendMsg, err = procedures.GetInitialUEMessage(ue.RanUeNgapId, registrationRequest, "")
	if err != nil {
		err = fmt.Errorf("Error getData")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error Sending Message")
		return
	}

	// receive NAS Authentication Request Msg
	n, err = amfConn.Read(recvMsg)
	if err != nil {
		err = fmt.Errorf("Error Receiving Data")
		return
	}
	ngapPdu, err := ngap.Decoder(recvMsg[:n])
	if err != nil {
		err = fmt.Errorf("Error Decoding NGAP")
		return
	}

	// Calculate for RES*
	nasPdu := procedures.GetNasPdu(ue, ngapPdu.InitiatingMessage.Value.DownlinkNASTransport)
	if err != nil {
		err = fmt.Errorf("Error getting NasPDU")
		return
	}
	rand := nasPdu.AuthenticationRequest.GetRANDValue()
	resStat := ue.DeriveRESstarAndSetKey(ue.AuthenticationSubs, rand[:], context.RAN_Self().Security.NetworkName)

	// send NAS Authentication Response
	pdu := nasTestpacket.GetAuthenticationResponse(resStat, "")
	sendMsg, err = procedures.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		err = fmt.Errorf("Error getting GetUplinkNASTransport")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error sending NAS Authentication")
		return
	}

	// receive NAS Security Mode Command Msg
	n, err = amfConn.Read(recvMsg)
	if err != nil {
		err = fmt.Errorf("Error Receiving NAS Security")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		err = fmt.Errorf("Error Decoding NAS Security")
		return
	}

	// send NAS Security Mode Complete Msg
	registrationRequestWith5GMM := nasTestpacket.GetRegistrationRequest(nasMessage.RegistrationType5GSInitialRegistration, mobileIdentity5GS, nil, ueSecurityCapability, ue.Get5GMMCapability(), nil, nil)
	pdu = nasTestpacket.GetSecurityModeComplete(registrationRequestWith5GMM)
	pdu, err = procedures.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext, true, true)
	if err != nil {
		err = fmt.Errorf("Error Getting NAS Security Mode Complete")
		return
	}
	sendMsg, err = procedures.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		err = fmt.Errorf("Error Getting NAS Security Mode Complete Message")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error Sending NAS Security Mode Complete")
		return
	}

	// receive ngap Initial Context Setup Request Msg
	n, err = amfConn.Read(recvMsg)
	if err != nil {
		err = fmt.Errorf("Error Receiving ngap Initial Context Setup Request Msg")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		err = fmt.Errorf("Error Decoding ngap Initial Context Setup Request Msg")
		return
	}

	// send ngap Initial Context Setup Response Msg
	sendMsg, err = procedures.GetInitialContextSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId)
	if err != nil {
		err = fmt.Errorf("Error getting ngap Initial Context Setup Response Msg")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error sending ngap Initial Context Setup Response Msg")
		return
	}

	// send NAS Registration Complete Msg
	pdu = nasTestpacket.GetRegistrationComplete(nil)
	pdu, err = procedures.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		err = fmt.Errorf("Error Encoding NasPDU with Security")
		return
	}
	sendMsg, err = procedures.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		err = fmt.Errorf("Error GetUplinkNASTransport NAS Registration Complete Msg")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error sending NAS Registration Complete Msg")
		return
	}

	time.Sleep(100 * time.Millisecond)

	// send GetPduSessionEstablishmentRequest Msg
	// Slice Parameters
	sNssai := models.Snssai{
		Sst: context.RAN_Self().Snssai.Sst,
		Sd:  context.RAN_Self().Snssai.Sd,
	}
	pdu = nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(10, nasMessage.ULNASTransportRequestTypeInitialRequest, "internet", &sNssai)
	pdu, err = procedures.EncodeNasPduWithSecurity(ue, pdu, nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true, false)
	if err != nil {
		err = fmt.Errorf("Error Getting PDU")
		return
	}
	sendMsg, err = procedures.GetUplinkNASTransport(ue.AmfUeNgapId, ue.RanUeNgapId, pdu)
	if err != nil {
		err = fmt.Errorf("Error GetUplinkNASTransport PDU")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error sending GetPduSessionEstablishmentRequest Msg")
		return
	}

	// receive 12. NGAP-PDU Session Resource Setup Request(DL nas transport((NAS msg-PDU session setup Accept)))
	n, err = amfConn.Read(recvMsg)
	if err != nil {
		err = fmt.Errorf("Error receiving GAP-PDU Session Resource Setup Request")
		return
	}
	_, err = ngap.Decoder(recvMsg[:n])
	if err != nil {
		err = fmt.Errorf("Error decoding GAP-PDU Session Resource Setup Request")
		return
	}

	// send 14. NGAP-PDU Session Resource Setup Response
	sendMsg, err = procedures.GetPDUSessionResourceSetupResponse(ue.AmfUeNgapId, ue.RanUeNgapId, ranGTPIpAddr)
	if err != nil {
		err = fmt.Errorf("Error getting NGAP-PDU Session Resource Setup Response")
		return
	}
	_, err = amfConn.Write(sendMsg)
	if err != nil {
		err = fmt.Errorf("Error sending NGAP-PDU Session Resource Setup Response")
		return
	}

	amfConn.Close()

	pduSession = &helper.PDUSession{Socket: upfConn, IPv4: userEquipment.IPv4}

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
	checksum := ipv4HeaderChecksum(&ipv4hdr)
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
