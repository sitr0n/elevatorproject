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
	
	connect_remote(&_remote[0], &_remote[1])

	
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

func connect_remote(r1 *remote, r2 *remote) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + PORT1)
	state.Check(err)
	listen_addr2, err := net.ResolveUDPAddr("udp", _localip + PORT2)
	state.Check(err)
	in_connection, _ := net.ListenUDP("udp", listen_addr)
	state.Check(err)
	in_connection2, _ := net.ListenUDP("udp", listen_addr2)
	state.Check(err)
	defer in_connection.Close()
	defer in_connection2.Close()
	
	local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	state.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", string(r1.address) + PORT1)
	state.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	state.Check(err)
	target_addr2,err := net.ResolveUDPAddr("udp", string(r2.address) + PORT2)
	state.Check(err)
	out_connection2, err := net.DialUDP("udp", local_addr, target_addr2)
	state.Check(err)
	defer out_connection.Close()
	defer out_connection2.Close()
	
	r1.input = in_connection
	r2.input = in_connection2
	r1.output = out_connection
	r2.output = out_connection2
	fmt.Println("Device ", r1.id , " connected to ", r1.address, PORT1)
	fmt.Println("Device ", r2.id , " connected to", r2.address, PORT2)
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
