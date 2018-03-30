package network

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
	"../state"
)

var _alive bool = false

const stasjon17 string = "129.241.187.145:10001"
const stasjon20 string = "129.241.187.155:10001"
const stasjon22 string = "129.241.187.56:10001"
const stasjon23 string = "129.241.187.57:10001"

const targetIP string = stasjon17

func Listener() {
	localip := find_localip()
	ch_wd_reset := make(chan bool)
	ch_wd_timeout := make(chan bool)
	
        target_addr, err := net.ResolveUDPAddr("udp", localip + ":10001")
    	state.Check(err)

        connection, _ := net.ListenUDP("udp", target_addr)
        state.Check(err)
        defer connection.Close()

        var message state.Elevator

        for {
		
                inputBytes := make([]byte, 4096)
                length, _, _ := connection.ReadFromUDP(inputBytes)
                buffer := bytes.NewBuffer(inputBytes[:length])
                decoder := gob.NewDecoder(buffer)
                decoder.Decode(&message)
                
		ch_wd_reset <- true
		if (get_alive() == false) {
			fmt.Println("A new elevator has joined!")
			set_alive(true)
			go watchdog(ch_wd_reset, ch_wd_timeout)
		}
		
		fmt.Println("Received: ", message)
		buffer.Reset()
	}
}

func Communication_handler(bcast <- chan state.Elevator, listen <- chan bool) {
	localip := find_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10001")
    	state.Check(err)
        in_connection, _ := net.ListenUDP("udp", listen_addr)
        state.Check(err)
        defer in_connection.Close()


        local_addr, err := net.ResolveUDPAddr("udp", localip + ":0")
	state.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", targetIP)
	state.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	state.Check(err)
	defer out_connection.Close()

	fmt.Println("Communication started!")
        var message state.Elevator

        for {
        	select {
        	case data := <- bcast:
        		var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)
			encoder.Encode(data)
			out_connection.Write(buffer.Bytes())
			fmt.Println("Broadcasting: ", data)
			buffer.Reset()

		case <- listen:
			inputBytes := make([]byte, 4096)
			fmt.Println("Starts listening....")
	                length, _, _ := in_connection.ReadFromUDP(inputBytes)
	                buffer := bytes.NewBuffer(inputBytes[:length])
	                decoder := gob.NewDecoder(buffer)
	                decoder.Decode(&message)
			
			fmt.Println("Received: ", message)
			buffer.Reset()
	        }
        }
}

func Broadcast_state(data *state.Elevator) {
	localip := find_localip()

	LocalAddr, err := net.ResolveUDPAddr("udp", localip + ":0")
	state.Check(err)

	ServerAddr,err := net.ResolveUDPAddr("udp", targetIP)
	state.Check(err)

	connection, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	state.Check(err)

	defer connection.Close()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	
	encoder.Encode(data)
	connection.Write(buffer.Bytes())
	fmt.Println("Broadcasting: ", *data)
	//buffer.Reset()
}

func watchdog(reset <- chan bool, timeout chan <- bool) {
	fmt.Println("Watchdog activated!\n")
	for i := 0; i < 10; i++ {
		select {
			case <- reset:
				i = 0
			default:
		}
		time.Sleep(1000*time.Millisecond)
	}
	fmt.Println("Connection lost.")
	timeout <- true
	set_alive(false)
}

func find_localip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	state.Check(err)
	defer conn.Close()

	mask := conn.LocalAddr().String()

	var local string = ""
	for _, char := range mask {
		if (char == ':') {
			break
		}
		local += string(char)
	}
	return local
}

func set_alive (b bool) {
	_alive = b
}

func get_alive () bool {
	return _alive
}
