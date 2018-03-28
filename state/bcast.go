package state

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	//"time"
)

//import state "../state"

const stasjon22 string = "129.241.187.56:0"
const stasjon23 string = "129.241.187.57:10001"

const localIP string =	stasjon22
const targetIP string = stasjon23

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
