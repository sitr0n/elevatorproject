package main

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
	"./state"
)


func listen() {

	//receive := make(chan state.Elevator)
        ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.57:10001")
    	state.Check(err)

        connection, err := net.ListenUDP("udp", ServerAddr)
        defer connection.Close()

        var message state.Elevator

        for {
		
                inputBytes := make([]byte, 4096)
		fmt.Println("Waiting for data...")
                length, _, _ := connection.ReadFromUDP(inputBytes)
                buffer := bytes.NewBuffer(inputBytes[:length])
                decoder := gob.NewDecoder(buffer)
                decoder.Decode(&message)
                
                //receive <- message
		//msg := <- receive
		fmt.Println("Found: ", message)
		time.Sleep(2*time.Second)
                 
	}
}

func main () {

	listen()
}


