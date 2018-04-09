package network

import (//"bytes"
	"encoding/json"
	"fmt"
	"net"
	"time"
	"../state"
	"log"
)

var _alive bool = false

const stasjon13 string = "129.241.187.152:10001"
const stasjon14 string = "129.241.187.142:10001"
const stasjon17 string = "129.241.187.145:10001"
const stasjon20 string = "129.241.187.155:10001"
const stasjon22 string = "129.241.187.56:10001"
const stasjon23 string = "129.241.187.57:10001"
const stasjon10 string = "129.241.187.158:10001"
const stasjon11 string = "129.241.187.159:10001"

const targetIP string = stasjon11
const (
	REMOTE_1   int	= 1
	REMOTE_2	= 2
)

//TODO: make remote button event listener

func Broadcast_state(bcast <- chan state.Elevator) {
	localip := get_localip()

	local_addr, err := net.ResolveUDPAddr("udp", localip + ":0")
	state.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", targetIP)
	state.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	state.Check(err)
	defer out_connection.Close()
	
	for {
		select {
		case data := <- bcast:
			send_state(data, out_connection)
			//send_state(data, second_out_connection)
		}
	}
}

func Poll_remote_state(output *state.Elevator) {
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10001")
	state.Check(err)
	input, _ := net.ListenUDP("udp", listen_addr)
	state.Check(err)
	defer input.Close()
	
	wd_reset := make(chan bool)

	for {
		if (is_alive() == false) {
			go watchdog(wd_reset)
			fmt.Println("Connection established!")
		}
		wd_reset <- true
		
		*output = read_state(input)
		fmt.Println("Received: ", *output)
	}
}

func Poll_remote_state2(output *state.Elevator) {
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10002")
	state.Check(err)
	input, _ := net.ListenUDP("udp", listen_addr)
	state.Check(err)
	defer input.Close()
	
	wd_reset := make(chan bool)

	for {
		if (is_alive() == false) {
			go watchdog(wd_reset)
			fmt.Println("Connection established!")
		}
		wd_reset <- true
		
		*output = read_state(input)
		fmt.Println("Received: ", *output)
	}
}


func send_state(state state.Elevator, connection *net.UDPConn) {
	jsonRequest, err := json.Marshal(state)
		if err != nil {
			log.Print("Marshal Register information failed.")
			log.Fatal(err)
		}
		connection.Write(jsonRequest)
	fmt.Println("Broadcasting: ", state)
	
}

func read_state(connection *net.UDPConn) state.Elevator {
	var message state.Elevator
	inputBytes := make([]byte, 4096)
	fmt.Println("Starts listening....")
	length, _, _ := connection.ReadFromUDP(inputBytes)
                err := json.Unmarshal(inputBytes[:length], &message)
		if err != nil {
			log.Print(err)
			//continue
		}
	
	return message
}

func watchdog(wd_reset <- chan bool) {
	//fmt.Println("Watchdog activated!\n")
	set_alive(true)
	for i := 0; i < 10; i++ {
		time.Sleep(500*time.Millisecond)
		select {
		case <- wd_reset:
			i = 0

		default:
		}
	}
	set_alive(false)
	fmt.Println("Connection lost with elevator 2.")
}


func get_localip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	state.Check(err)
	defer conn.Close()

	ip_with_port := conn.LocalAddr().String()

	var ip string = ""
	for _, char := range ip_with_port {
		if (char == ':') {
			break
		}
		ip += string(char)
	}
	return ip
}

func set_alive(b bool) {
	_alive = b
}

func is_alive() bool {
	return _alive
}
