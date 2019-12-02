package udp2tcplib

import "fmt"

type TCPHeader struct {
	AckNum, SeqNum uint32
	ConnID         uint16
	ACK, SYN, FIN  bool
}

func (h *TCPHeader) String() string {
	if h == nil {
		return "<nil_>"
	}
	return fmt.Sprintf("AckNum : %d SeqNum : %d ConnID : %d Flags: { ACK : %v , SYN : %v , FIN : %v }", h.AckNum, h.SeqNum, h.ConnID, h.ACK, h.SYN, h.FIN)
}
