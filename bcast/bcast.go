package bcast

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
)

import state "../state"

func Broadcast(data *state.Elevator) {
       
	ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.57:10001")
  	state.Check(err)
 
  	LocalAddr, err := net.ResolveUDPAddr("udp", "129.241.187.56:0")
 	state.Check(err)
 
	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	state.Check(err)

        defer connection.Close()
        
        
        
        var buffer bytes.Buffer
        encoder := gob.NewEncoder(&buffer)
	
        //message := <-state
        encoder.Encode(data)
        connection.Write(buffer.Bytes())
        fmt.Println("Broadcasting: ")
        buffer.Reset()
}
