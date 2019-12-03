# Feito com &hearts; + ☕ em GO

# UDP2TCP
## Exsitem duas aplicações 

### Servidor 

### Cliente

## E uma biblioteca 

## UDP2TCPLib

# imports
> fmt : biblioteca para chamadas de output (print)
> encoding/binary : biblioteca para manipulação de bits e conversão de dados
> sync : biblioteca para controle de concorrência (mutex), e condições de corrida
  
  
# Para compilar
>  A aplicação consiste em 2 programas diferentes e uma biblioteca 
>  para realizar a compilação é necessário ter a linguagem go instalada
> 
>  ## comandos para montar o ambiente de desenvolvimento
>  export GOLANG=<Caminho absoluto para a pasta udp2ctp>
>  exemplo : export GOLANG=C:/Users/luiz/redes/udp2tcp
>  
>  ## Comandos em go para compilar e montar o pacote udp2tcplib
>  
>  go build redes/udp2tcplib
>  go install redes/udp2tcplib
>  
>  ### Comandos para compilar a aplicação cliente e servidor
>  go install redes/udp2tcp/client
>  go install redes/udp2tcp/server


# Para executar

>## Cliente 
>
>./client.exe <ip do servidor> <porta do servidor> <caminho para o arquivo>
>  
> ## Servidor
> ./server.exe <porta do servidor> <diretório para salvar>
>
  
##  Estruturas:
```go
//UDP2TcpPacket Representação do pacote enviado pela rede
type UDP2TcpPacket struct {
	TCPHeader
	Data []byte
}

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

```  

# Esquema das Estruturas:



# Funções e Procedimentos Utilizadas de bibliotecas importadas

## Procedimentos de Inicialização:

## Funções de Instânciação:

## Funcões de Carremagento de Disco:

## Procedimeontos de Remoção (Destrução)

## Porcedimentos/Funções e Recuperação de Definição (Get  e Set):

## Procedimentos/Funções Orientados à Eventos:

# Funções e Procedimentos das Bibliotecas Pessoais:

