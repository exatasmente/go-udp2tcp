package udp2tcplib

import (
	"encoding/binary"
	"net"
	"sync"
)

//UDP2TcpPacket Representação do pacote enviado pela rede
type UDP2TcpPacket struct {
	TCPHeader
	Data []byte
}

//UDP2TcpType variavél de Enumeração
type UDP2TcpType int

// Constantes
const (
	Client     = UDP2TcpType(1)
	Server     = UDP2TcpType(2)
	MAXSEQNUM  = 102400
	MAXTIMEOUT = 10
	MINTIMEOUT = 1
)

//UDP2TcpSocket Representação do Socket para a biblioteca
type UDP2TcpSocket struct {
	conn *net.UDPConn
	Type UDP2TcpType
}

//UDP2TcpConnState Estrutura para controle do estado do conexão
type UDP2TcpConnState struct {
	mux        sync.Mutex
	readChan   chan UDP2TcpPacket
	writeChan  chan []byte
	socketChan chan UDP2TcpPacket
	fileSize   chan int
	SSTHRESH   int
	CWND       int
	window     []uint32
}

//UDP2TcpConn representação da conexão com o servidor
type UDP2TcpConn struct {
	TCPHeader
	UDP2TcpConnState
	addr *net.UDPAddr
	file map[int][]byte
}

//UDP2Tcp Estrutura base para controle da biblioteca
type UDP2Tcp struct {
	UDP2TcpSocket
}

// Pacote inicial
var packet UDP2TcpPacket = buildPacket(4321, 123456, 0, true, true, false, nil)

//isValid função para verificação se o pacote está com as flags definidas
func (packet UDP2TcpPacket) isValid() bool {

	return bool2Binary(packet) > 0
}

//hasPayload verificação se o pacote possui dados (payload)
func (packet UDP2TcpPacket) hasPayload() bool {
	if len(packet.Data) > 0 {
		return true
	}
	return false
}

//Write Função de alto nivél para envio do arquivo pelo cliente
func (conn UDP2TcpConn) Write(file []byte) {
	// Variavéis para contole
	var start int = 0
	var end int = 512
	close := false
	lenFile := 0

	// envia o tamanho do arquivo + o valor de sequência inicial para o canal fileSize
	conn.fileSize <- len(file) + 123456

	// Loop
	for {

		if close == true {
			break
		}

		// verifica a capacidade da janela de transferência
		conn.mux.Lock()
		lenWindow := cap(conn.window)
		conn.mux.Unlock()

		// variavél para armazenar parte do arquivo a ser enviado
		var payload []byte

		// caso não seja possivél enviar
		if lenWindow == 0 {
			continue
		}
		conn.mux.Lock()
		// preenchendo a janela
		for lenWindow > 0 && lenFile < len(file) {
			lenWindow--
			if end <= len(file) {
				lenFile += 512
				payload = file[start:end]
				start = end
				end += 512

			} else {
				lenFile = len(file)
				payload = file[start:]

			}

			// envia o dado para o canal de comunicação com o socket
			conn.writeChan <- payload

		}
		conn.mux.Unlock()
	}

}

//Read Le o arquvo enviado pelo cliente
func (conn UDP2TcpConn) Read() []byte {
	data := <-conn.readChan
	return data.Data
}

// Tranforma o dado enviado pela rede []byte para a representação do pacote no sistema UDP2TcpPacket
func dump(buf []byte, rlen int, addr *net.UDPAddr) UDP2TcpPacket {

	seqNum := binary.BigEndian.Uint32(buf[0:])
	ackNum := binary.BigEndian.Uint32(buf[4:])
	connID := binary.BigEndian.Uint16(buf[8:])
	ACK, SYN, FIN := binaryToBool(buf[10])
	packet = UDP2TcpPacket{
		TCPHeader: TCPHeader{SeqNum: seqNum, AckNum: ackNum, ConnID: connID, ACK: ACK, SYN: SYN, FIN: FIN},
		Data:      buf[11:rlen],
	}

	return packet
}

// Constrói um pacote
func buildPacket(seqNum, ackNum uint32, connID uint16, ACK, SYN, FIN bool, data []byte) UDP2TcpPacket {
	tcpPacket := UDP2TcpPacket{
		TCPHeader: TCPHeader{seqNum, ackNum, connID, ACK, SYN, FIN},
		Data:      data,
	}

	return tcpPacket
}

// Serializa o pacote para ser enviado pela rede
func serializePacket(tcpPacket UDP2TcpPacket) []byte {
	bytePacket := make([]byte, 12+len(tcpPacket.Data))
	binary.BigEndian.PutUint32(bytePacket[0:], tcpPacket.SeqNum)
	binary.BigEndian.PutUint32(bytePacket[4:], tcpPacket.AckNum)
	binary.BigEndian.PutUint16(bytePacket[8:], tcpPacket.ConnID)
	bytePacket[10] = bool2Binary(tcpPacket)

	if len(tcpPacket.Data) > 0 {
		copy(bytePacket[11:], tcpPacket.Data)
	}
	return bytePacket
}

// Converte as flags em inteiro
func bool2Binary(packet UDP2TcpPacket) byte {
	var flags byte = 0
	if packet.ACK {
		flags += 1
	}
	if packet.SYN {
		flags += 2

	}
	if packet.FIN {
		flags += 4
	}
	return flags

}

// converte o inteiro nas flags
func binaryToBool(flags byte) (bool, bool, bool) {
	var ACK, SYN, FIN bool

	switch flags {
	case 1:
		ACK = true
		SYN = false
		FIN = false
		break
	case 2:
		ACK = false
		SYN = true
		FIN = false
		break
	case 3:
		ACK = true
		SYN = true
		FIN = false
		break
	case 4:
		ACK = false
		SYN = false
		FIN = true
		break
	case 5:
		ACK = true
		SYN = false
		FIN = true
		break
	case 6:
		ACK = false
		SYN = true
		FIN = true
		break
	default:
		ACK = false
		SYN = false
		FIN = false
		break
	}

	return ACK, SYN, FIN
}
