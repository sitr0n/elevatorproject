package bcast

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
)

import state "../state"

const stasjon22 string = "129.241.187.56:0"

const localIP string = stasjon22

func Broadcast(data *state.Elevator) {
       
	ServerAddr,err := net.ResolveUDPAddr("udp","129.241.187.57:10001")
  	state.Check(err)
 
  	LocalAddr, err := net.ResolveUDPAddr("udp", localIP)
 	state.Check(err)
 
	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	state.Check(err)

        defer connection.Close()
        
        
        
        var buffer bytes.Buffer
        encoder := gob.NewEncoder(&buffer)
	
        //message := <-state
        encoder.Encode(data)
        connection.Write(buffer.Bytes())
        fmt.Println("Broadcasting: ", data)
        buffer.Reset()
}
