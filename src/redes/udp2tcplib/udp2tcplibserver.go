package udp2tcplib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var clients map[uint16]*UDP2TcpConn = make(map[uint16]*UDP2TcpConn)
var connID uint16 = 1

func readFile(fname string) (nums int, err error) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(b), "\n")

	var num int

	for _, l := range lines {

		if len(l) == 0 {
			continue
		}

		n, err := strconv.Atoi(l)
		if err != nil {
			return 0, err
		}
		num = n
	}

	return num, nil
}

func (udp2tcp *UDP2Tcp) ListenUDP2TCP(ip string, port int) {
	var err error
	serverAddr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	num, err := readFile("server.dat")
	udp2tcp.Type = Server
	// listen to incoming udp packets
	udp2tcp.conn, err = net.ListenUDP("udp", &serverAddr)
	connID = uint16(num)
	if err != nil {
		log.Fatal(err)
	}

}

func (udp2tcp *UDP2Tcp) Accept() (conn *UDP2TcpConn, err error) {
	buffer := make([]byte, 524)

	n, addr, err := udp2tcp.conn.ReadFromUDP(buffer)

	if err != nil {
		panic(err)
	}
	packet := dump(buffer, n, addr)
	newConn := false
	fmt.Println("Recive", packet.SeqNum, packet.AckNum, packet.ConnID, packet.ACK, packet.SYN, packet.FIN)
	if packet.isValid() {

		//fmt.Println(packet)
		if packet.ConnID == 0 && packet.SYN == true {
			// New Conn Handshake
			conn = new(UDP2TcpConn)
			conn.readChan = make(chan UDP2TcpPacket, 1)
			conn.socketChan = make(chan UDP2TcpPacket, 1)
			conn.writeChan = make(chan []byte)
			packet.ConnID = connID
			conn.addr = addr
			connID++
			conn.TCPHeader = packet.TCPHeader
			newConn = true
			ioutil.WriteFile("server.dat", []byte(strconv.Itoa(int(connID))), 0655)
			go udp2tcp.handleServerConn(conn, addr)

		} else if packet.ACK == true || packet.FIN == true {
			conn = clients[packet.ConnID]
			conn.socketChan <- packet
			newConn = false
		}

	}
	if newConn {
		return conn, err
	}
	return nil, err
}

func (udp2tcp *UDP2Tcp) handleServerConn(conn *UDP2TcpConn, addr *net.UDPAddr) {
	clients[conn.ConnID] = conn
	conn.mux.Lock()
	defer delete(clients, conn.ConnID)
	conn.mux.Unlock()

	wg := sync.WaitGroup{}

	wg.Add(3)
	var fileLen int = 0
	done := make(chan bool)

	timer := time.AfterFunc(MAXTIMEOUT*time.Second,
		func() {
			log.Println("Connection Timeout")

			pkt := buildPacket(4321, 123456, conn.ConnID, true, true, false, nil)
			pkt.Data = []byte("ERROR")

			conn.readChan <- pkt

		})

	// HandShake
	go func(wg *sync.WaitGroup) {
		pkt := buildPacket(4321, 123456, conn.ConnID, true, true, false, nil)
		data := serializePacket(pkt)
		fileLen = 123456
		conn.writeChan <- data
		log.Println("Sever HandShake Ok")
		wg.Done()
		timer.Reset(MAXTIMEOUT * time.Second)
	}(&wg)

	go func(wg *sync.WaitGroup) {
		clientFile := make(map[uint32][]byte)
		for {
			pkt := <-conn.socketChan
			timer.Reset(MAXTIMEOUT * time.Second)
			fileLen += len(pkt.Data)
			conn.mux.Lock()
			pkt.AckNum = uint32(fileLen)
			conn.mux.Unlock()
			if pkt.FIN == true {

				pkt := buildPacket(pkt.SeqNum, pkt.AckNum, pkt.ConnID, true, false, true, nil)
				data := serializePacket(pkt)
				fmt.Println("Sent", pkt.AckNum, pkt.SeqNum, pkt.ACK, pkt.SYN, pkt.FIN)
				conn.mux.Lock()
				udp2tcp.conn.WriteTo(data, conn.addr)
				conn.mux.Unlock()
				var file []byte

				keys := make([]int, 0, len(clientFile))
				for k := range clientFile {
					keys = append(keys, int(k))
				}
				sort.Ints(keys)
				for _, k := range keys {
					file = append(file, clientFile[uint32(k)]...)
				}
				pkt.Data = bytes.Trim(file, "\x00")
				log.Println("Close Conection")

				conn.readChan <- pkt
				break
			} else {
				clientFile[pkt.SeqNum] = pkt.Data
				pkt := buildPacket(pkt.SeqNum, pkt.AckNum, pkt.ConnID, true, false, false, nil)
				fmt.Println(pkt)
				response := serializePacket(pkt)
				conn.writeChan <- response

			}
		}
		done <- true
		wg.Done()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		close := false
		for {
			if close == true {
				break
			}
			select {
			case <-done:
				close = true

			case data := <-conn.writeChan:
				timer.Reset(MAXTIMEOUT * time.Second)
				_, err := udp2tcp.conn.WriteTo(data, conn.addr)
				if err != nil {
					panic(err)
				}

			}
		}
		wg.Done()
	}(&wg)

	wg.Wait()

}
