package udp2tcplib

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

func (udp2tcp *UDP2Tcp) DialUDP2Tcp(ip string, port int) (conn UDP2TcpConn, err error) {

	serverAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}

	udp2tcp.Type = Client
	udp2tcp.conn, err = net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		panic(err)
	}

	conn = UDP2TcpConn{}
	conn.SSTHRESH = 1000
	conn.CWND = 512
	conn.window = make([]uint32, 1)
	conn.readChan = make(chan UDP2TcpPacket)
	conn.writeChan = make(chan []byte)
	conn.socketChan = make(chan UDP2TcpPacket)
	conn.fileSize = make(chan int)
	conn.file = make(map[int][]byte)

	go conn.handleClientConn(*udp2tcp)

	return conn, err
}

func (conn UDP2TcpConn) handleClientConn(udp2tcp UDP2Tcp) {
	defer udp2tcp.conn.Close()

	done := make(chan bool)
	wg := sync.WaitGroup{}

	wg.Add(3)
	// HandShake

	buffer := make([]byte, 524)
	fileLen := <-conn.fileSize
	pkt := buildPacket(12345, 0, 0, false, true, false, nil)
	data := serializePacket(pkt)

	n, err := udp2tcp.conn.Write(data)
	fmt.Println(n)
	if err != nil {
		log.Println(err)
	}
	n, addr, err := udp2tcp.conn.ReadFromUDP(buffer)
	if err != nil {
		log.Println(err)
	}
	serverData := dump(buffer, n, addr)

	if serverData.isValid() {

		conn.SeqNum = serverData.SeqNum
		conn.AckNum = serverData.AckNum + 1
		conn.ConnID = serverData.ConnID
		conn.SYN = false
		conn.ACK = true
		conn.FIN = false
		aux := conn.AckNum
		conn.AckNum = conn.SeqNum
		conn.SeqNum = aux

		log.Println("Cient HandShake Success")

	} else {
		return
	}

	// Read From udp Client handle
	go func(wg *sync.WaitGroup) {
		close := false
		for {
			if close == true {
				break
			}
			select {
			case <-done:
				close = true
				break
			default:
				buffer := make([]byte, 524)
				n, addr, _ := udp2tcp.conn.ReadFromUDP(buffer)

				if n < 11 {
					continue
				}
				data := dump(buffer, n, addr)

				if data.isValid() {

					if conn.CWND <= conn.SSTHRESH {
						conn.CWND += 512

					}
					if conn.CWND >= conn.SSTHRESH {
						conn.CWND += (512 * 512) / conn.CWND
					}
					conn.socketChan <- data
				}
			}

		}
		log.Println("Done 1")
		wg.Done()
	}(&wg)

	var timeout int = 0

	go func(wg *sync.WaitGroup) {
		for {
			pkt := <-conn.socketChan
			fmt.Println("Recive", pkt.AckNum, pkt.SeqNum, pkt.ACK, pkt.SYN, pkt.FIN)
			index := conn.isInWindow(pkt.AckNum - 1)
			if pkt.FIN == true && pkt.ACK == true && conn.SSTHRESH == -1 {

				break
			} else if conn.SSTHRESH == -1 {
				pkt = buildPacket(pkt.SeqNum, 0, pkt.ConnID, true, false, false, nil)
				data := serializePacket(pkt)

				conn.writeChan <- data
			}
			if int(pkt.AckNum) == fileLen {
				conn.SSTHRESH = -1

				pkt = buildPacket(pkt.SeqNum, 0, pkt.ConnID, false, false, true, nil)
				data := serializePacket(pkt)
				fmt.Println(pkt.AckNum, fileLen)
				conn.writeChan <- data
				if err != nil {
					log.Println(err)
				}

			}

			if index >= 0 {
				timeout = 0
				conn.TCPHeader = pkt.TCPHeader
				aux := conn.SeqNum
				conn.SeqNum = conn.AckNum

				conn.window = append(conn.window[:index], conn.window[index+1:]...)

				if conn.ACK == false {
					conn.AckNum = 0
				} else {
					conn.AckNum = aux
				}

			} else {
				timeout++
				if timeout == 3 {

					conn.SSTHRESH = conn.CWND
					conn.CWND = 512
					timeout = 0
					conn.AckNum = conn.window[0]
					conn.window = make([]uint32, 1)

				}
			}
			fmt.Println(len(conn.window))
			if (int(pkt.SeqNum) == fileLen && len(conn.window) == 0) || conn.SSTHRESH == -1 {
				conn.SSTHRESH = -1
				fmt.Println("Send close request")
				pkt = buildPacket(pkt.SeqNum, 0, pkt.ConnID, false, false, true, nil)
				data := serializePacket(pkt)
				fmt.Println(pkt.AckNum, fileLen)
				go func() { conn.writeChan <- data }()
				if err != nil {
					log.Println(err)
				}

			}
		}
		log.Println("Done 2")
		wg.Done()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		close := false
		lenFile := 123456
		for {

			if close == true {
				break
			}
			payload := <-conn.writeChan

			if conn.SSTHRESH == -1 {
				fmt.Println("okd")
				udp2tcp.conn.Write(payload)
				close = true
			} else {
				lenFile += len(payload)
				pkt = buildPacket(0, uint32(lenFile), conn.ConnID, conn.ACK, conn.SYN, conn.FIN, payload)
				conn.window = append(conn.window, uint32(fileLen))
				data := serializePacket(pkt)
				fmt.Println("Sent", pkt.AckNum, pkt.SeqNum, pkt.ACK, pkt.SYN, pkt.FIN)

				_, err := udp2tcp.conn.Write(data)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		log.Println("Done 3")
		wg.Done()
		done <- true
	}(&wg)

	wg.Wait()
	os.Exit(0)
}

func (conn *UDP2TcpConn) isInWindow(seqNum uint32) int {
	for i, seq := range conn.window {
		if seq == seqNum {
			return i
		}
	}
	return -1
}
