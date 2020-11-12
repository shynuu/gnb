package procedures

import (
	"free5gc/lib/nas"
	"free5gc/lib/ngap/ngapType"
	"free5gc/src/gnb/uee"
)

func GetNasPdu(ue *uee.RanUeContext, msg *ngapType.DownlinkNASTransport) (m *nas.Message) {
	for _, ie := range msg.ProtocolIEs.List {
		if ie.Id.Value == ngapType.ProtocolIEIDNASPDU {
			pkg := []byte(ie.Value.NASPDU.Value)
			m, err := NASDecode(ue, nas.GetSecurityHeaderType(pkg), pkg)
			if err != nil {
				return nil
			}
			return m
		}
	}
	return nil
}
