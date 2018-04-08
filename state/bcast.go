package state

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	//"time"
)

//import state "../state"

const stasjon17 string = "129.241.187.145:10001"
const stasjon20 string = "129.241.187.155:0"
const stasjon22 string = "129.241.187.56:0"
const stasjon23 string = "129.241.187.57:10001"
const stasjon10 string = "129.241.187.158:10001"
const stasjon11 string = "129.241.187.159:10001"

const localIP string =	stasjon10
const targetIP string = stasjon11

func Broadcast(data *Elevator) {
	ServerAddr,err := net.ResolveUDPAddr("udp", targetIP)
	Check(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", localIP)
	Check(err)

	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	Check(err)

	defer connection.Close()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	
	encoder.Encode(data)
	connection.Write(buffer.Bytes())
	fmt.Println("Broadcasting: ", *data)
	//buffer.Reset()
}
