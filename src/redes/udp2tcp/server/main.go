package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"redes/udp2tcplib"
	"strconv"
)

var fileDir string

func main() {
	tcp := new(udp2tcplib.UDP2Tcp)
	var port int
	var fileDir string
	if len(os.Args[1:]) == 2 {
		port, _ = strconv.Atoi(os.Args[1])
		fileDir = os.Args[2]
	} else {
		fmt.Println("Parâmetros Inválidos digite os parêmetros na forma : <porta> <diretório>")
		os.Exit(1)
	}

	tcp.ListenUDP2TCP("", port)
	CreateDirIfNotExist(fileDir)
	for {

		conn, err := tcp.Accept()

		if err != nil {
			panic(err)

		}
		if conn != nil {
			go handleConn(conn, fileDir)
		}
		//fmt.Println(data)

	}
}

func handleConn(conn *udp2tcplib.UDP2TcpConn, fileDir string) {
	var close bool = false
	for {
		if close == true {
			fmt.Println("Server closer ")
		}

		data := conn.Read()
		if data != nil {
			err := ioutil.WriteFile(fileDir+"/"+strconv.Itoa(int(conn.ConnID))+".file", data, 0655)
			if err != nil {
				panic(err)
			}
			close = true
		}

	}
}
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
