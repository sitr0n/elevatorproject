package network

import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	"log"
)
import state "../state"
import order "../driver"
import unsafe "unsafe"


type ip string
type remote struct {
	id		int
	input		*net.UDPConn
	output 		*net.UDPConn
	address  	ip
	alive 		bool
	state 		state.Elevator
}

var _remote [2]remote
var _localip string

func Get_remote(index int) remote {
	return _remote[index]
}

func Init(first_remote interface{}, second_remote interface{}, ch_bcast <- chan interface{}, ch_order chan <- order.Order, ch_ack chan <- bool) {
	_localip = get_localip()
	_remote[0].id = 0
	_remote[0].address = ip_address(first_remote)
	_remote[0].alive = false
	_remote[1].id = 1
	_remote[1].address = ip_address(second_remote)
	_remote[1].alive = false
	
	connect_remote(&_remote[0])
	connect_remote(&_remote[1])
	
	//ch_list := make(chan interface{})
	
	go remote_listener(&_remote[0], ch_order, ch_ack)
	go remote_broadcaster(_remote[0].output, ch_bcast)
	
	go remote_listener(&_remote[1], ch_order, ch_ack)
	go remote_broadcaster(_remote[1].output, ch_bcast)
}



func remote_listener(r *remote, ch_order chan <- order.Order, ch_ack chan <- bool) {
	var state state.Elevator
	var order order.Order
	var ack bool = false
	const STATE_SIZE = int(unsafe.Sizeof(state))
	const ORDER_SIZE = int(unsafe.Sizeof(order))
	const ACK_SIZE = int(unsafe.Sizeof(ack))
	wd_kick := make(chan bool)
	inputBytes := make([]byte, 4096)
	
	fmt.Println("Starting remote", r.id, "listener!")
	for {
		length, _, _ := r.input.ReadFromUDP(inputBytes)
		wd_kick <- true
		if (r.alive == false) {
			go watchdog(r.id, wd_kick)
			fmt.Println("Connection established!")
		}
		
		switch length {
		case ACK_SIZE:
			err := json.Unmarshal(inputBytes[:length], &ack)
			checker(err)
			ch_ack <- ack
			break
			
		case ORDER_SIZE: 
			err := json.Unmarshal(inputBytes[:length], &order)
			checker(err)
			ch_order <- order
			break
			
		case STATE_SIZE:
			err := json.Unmarshal(inputBytes[:length], &state)
			checker(err)
			
			r.state = state
			//ch_state <- state
			break
		
		default:
			fmt.Println("Oops! Received something unexpected from remote...")
		}
	}
}

func remote_broadcaster(connection *net.UDPConn, message <- chan interface{}) {
	fmt.Println("Starting remote bcaster")
	for {
		select {
		case msg := <- message:
			encoded, err := json.Marshal(msg)
			checker(err)
			
			connection.Write(encoded)
		}
	}
}

func watchdog(index int, kick <- chan bool) {
	_remote[index].alive = true
	for i := 0; i < 10; i++ {
		time.Sleep(50*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	_remote[index].alive = false
}

func connect_remote(r *remote) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + PORT)
	state.Check(err)
	in_connection, _ := net.ListenUDP("udp", listen_addr)
	state.Check(err)
	defer in_connection.Close()
	
	local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	state.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", string(r.address) + PORT)
	state.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	state.Check(err)
	defer out_connection.Close()
	
	r.input = in_connection
	r.output = out_connection
	fmt.Println("Device ", r.id , " connected!")
}

func ip_address(adr interface{}) ip {
	switch a := adr.(type) {
	case ip:
		return a
	case int:
		if (a > 23 || a < 0) {
			fmt.Println("Workspace index is out of bounds. Please abort process and try another argument!")
			for {
			}
		} else {
			return WORKSPACE[a]
		}
	default:
		fmt.Println("Wrong data type passed to network.Init. Try string or workspace number.")
		return "0"
	}
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

func checker(e error) {
	if e != nil {
		log.Print(e)
		//continue
	}
}