package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"redes/udp2tcplib"
	"strconv"
)

// Variáreis Golbais
//  ip endereço ip do servidor
// 	port porta para comunicação
//  caminho para o arquivo que será enviado

var ip string
var port int
var filePath string

func main() {

	// udp2tcplb é o pacote responsável por implementar o protocolo confiavél sobre a rede
	// A estrutura principal é a UDP2Tcp
	tcp := new(udp2tcplib.UDP2Tcp)

	// Verificação dos parâmetros passados para a aplicação
	if len(os.Args[1:]) != 3 {
		fmt.Println("Parâmetros Inválidos digite os parêmetros na forma :<servidor> <porta> <diretório>")
		os.Exit(0)
	}
	port, _ = strconv.Atoi(os.Args[2])
	ip, filePath = os.Args[1], os.Args[3]

	// Realiza a ligação com o servior ataves do ip e porta
	conn, err := tcp.DialUDP2Tcp(ip, port)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	handleConn(conn)

}

func handleConn(conn udp2tcplib.UDP2TcpConn) {

	// Lê o arquivo e envia para o servidor
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// chamada para enviar o arquivo para o servidor
	conn.Write(input)

}
