package main

import (
        "bytes"
        "encoding/gob"
        "fmt"
        "net"
	//"time"
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

/*
func Init() {
        receive := make(chan Packet, 10)
        send := make(chan Packet, 10)
        go listen(receive)
        go broadcast(send)
        
}
*/
func listen(receive chan Packet) {
        ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.143:10001")
    	CheckError(err)

        connection, err := net.ListenUDP("udp", ServerAddr)
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
 
  	LocalAddr, err := net.ResolveUDPAddr("udp", "129.241.187.143:0")
 	CheckError(err)
 
	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

        defer connection.Close()
        
        var buffer bytes.Buffer
        encoder := gob.NewEncoder(&buffer)
	
	p := Packet{ID: 12345, Response: "It does maafaka"}
	send <- p
	/*var p1 Packet = p
	i := 0
	p1.ID += i
	i++*/
        for {
		
		//send <- p
                message := <-send
                encoder.Encode(message)
                connection.Write(buffer.Bytes())
                buffer.Reset()
		//time.Sleep(time.Second)
        }
}

func main(){

	exit := make(chan bool)
	receive := make(chan Packet, 10)
        send := make(chan Packet, 10)
	go broadcast(send)
        go listen(receive)
        
	<- exit
}
