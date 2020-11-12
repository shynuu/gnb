package procedure

import (
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
	"time"
)

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
