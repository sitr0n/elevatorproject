package main

import (
	"encoding/json"
        "fmt"
        "net"
	"time"
	"log"
)



type Packet struct {
        ID int
	Response string
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func listen(receive chan Packet) {

        ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.158:10001")
    	CheckError(err)

        connection, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

        defer connection.Close()

        var message Packet

        for {
		inputBytes := make([]byte, 4096)
                length, _, _ := connection.ReadFromUDP(inputBytes)
                err = json.Unmarshal(inputBytes[:length], &message)
		if err != nil {
			log.Print(err)
			continue
		}
		receive <- message
		msg := <- receive
		fmt.Println(msg.ID)
		fmt.Println(msg.Response)

            }
}



func broadcast(send chan Packet) {
       

	ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.146:10001")
  	CheckError(err)

	ServerAddr2,err := net.ResolveUDPAddr("udp","129.241.187.159:10001")
  	CheckError(err)
 
  	LocalAddr, err := net.ResolveUDPAddr("udp", "129.241.187.158:0")
 	CheckError(err)
 
	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	connection2, err := net.DialUDP("udp", LocalAddr, ServerAddr2)
	CheckError(err)

        defer connection.Close()
	defer connection2.Close()
        
        
	
	p := Packet{ID: 111, Response: "PC 158 reporting in"}
	
        for {

		send <- p
                message := <-send
		jsonRequest, err := json.Marshal(message)
		if err != nil {
			log.Print("Marshal Register information failed.")
			log.Fatal(err)
		}
		connection.Write(jsonRequest)
		connection2.Write(jsonRequest)

		time.Sleep(time.Second*2)

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
