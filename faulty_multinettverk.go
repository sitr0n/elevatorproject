package main

import (
        "bytes"
        "encoding/gob"
        "fmt"
        "net"
	"time"
)



type Packet struct {
        ID int
	Response string
        //Content []byte
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func listen(receive chan Packet) {

        ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.143:10001")
    	CheckError(err)

        connection, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

        defer connection.Close()

        var message Packet

        for {
                inputBytes := make([]byte, 4096)
                length, _, _ := connection.ReadFromUDP(inputBytes)
                buffer := bytes.NewBuffer(inputBytes[:length])
                decoder := gob.NewDecoder(buffer)
                decoder.Decode(&message)
                
                receive <- message
		msg := <- receive
		fmt.Println(msg.ID)
		fmt.Println(msg.Response)
                 
            }
}



func broadcast(send chan Packet) {
       

	ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.141:10001")
  	CheckError(err)

	ServerAddr2,err := net.ResolveUDPAddr("udp","129.241.187.146:10001")
  	CheckError(err)
 
  	LocalAddr, err := net.ResolveUDPAddr("udp", "129.241.187.143:0")
 	CheckError(err)
 
	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	connection2, err := net.DialUDP("udp", LocalAddr, ServerAddr2)
	CheckError(err)

        defer connection.Close()
	defer connection2.Close()
        
        var buffer bytes.Buffer
        encoder := gob.NewEncoder(&buffer)
	
	p := Packet{ID: 789, Response: "PC 143 reporting in"}
	
        for {
		
		send <- p
                message := <-send
                encoder.Encode(message)
                connection.Write(buffer.Bytes())
		connection2.Write(buffer.Bytes())
                buffer.Reset()
		time.Sleep(time.Second*5)
        }
}

func main(){

	exit := make(chan bool)
	receive := make(chan Packet, 100)
        send := make(chan Packet, 100)
	go broadcast(send)
        go listen(receive)
        
	<- exit
}
