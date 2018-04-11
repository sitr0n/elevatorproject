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

const ack_stasjon10 string = "129.241.187.158"
const ack_stasjon11 string = "129.241.187.159"
const ackIP1 string = ack_stasjon10
const ackIP2 string = ack_stasjon11

const (
	REMOTE_1   int	= 1
	REMOTE_2	= 2
)
type ip string
type remote struct {
	id		int
	input		*net.UDPConn
	output 		*net.UDPConn
	address  	ip
	alive 		bool
	elevator 	state.Elevator
}
var _remote [2]remote
var _localip string

func Init(first_remote interface{}, second_remote interface{}) {
	_localip = get_localip()
	_remote[0].id = 0
	_remote[0].address = ip_address(first_remote)
	_remote[0].alive = false
	_remote[1].id = 1
	_remote[1].address = ip_address(second_remote)
	_remote[1].alive = false
	
	connect_remote(&_remote[0])
	connect_remote(&_remote[1])
	
	//go Ack_listener1()
	//go Ack_listener2()
	//go Ack_broadcaster()
}

func connect_remote(r *remote) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + ":10001")
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

func remote_listener(r *remote, response chan <- interface{}) {
	var state state.Elevator
	var order order.Order
	var ack bool
	wd_kick := make(chan bool)
	
	inputBytes := make([]byte, 4096)
	
	fmt.Println("Starting remote listener ")
	for {
		length, _, _ := r.input.ReadFromUDP(inputBytes)
		wd_kick <- true
		if (r.alive == false) {
			go _watchdog(r.id, wd_kick)
			fmt.Println("Connection established!")
		}
		
		if (length > 10) {
			err := json.Unmarshal(inputBytes[:length], &state)
			checker(err)
			response <- state
		} else if (length > 5) {
			err := json.Unmarshal(inputBytes[:length], &order)
			checker(err)
			response <- order
		} else if (length > 1) {
			err := json.Unmarshal(inputBytes[:length], &ack)
			checker(err)
			response <- ack
		}
	}
}
func checker(e error) {
	if e != nil {
		log.Print(e)
		//continue
	}
}


func _watchdog(index int, kick <- chan bool) {
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

func Ack_listener1(ack_wd1_reset chan<- bool) {
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + PORT)
	state.Check_conn(err)
	input, _ := net.ListenUDP("udp", listen_addr)
	state.Check_conn(err)
	defer input.Close()

	buffer := make([]byte, 1024)

	for {
		n,addr,err := input.ReadFromUDP(buffer)
		state.Check_conn(err)
		fmt.Println("Received ",string(buffer[0:n]), " from ",addr)
		if (string(buffer[0:n]) == "Hello") {
			res := "reset"
			buffer = []byte(res)
			ack_wd1_reset <- true	
		} 
	}
}
	


func Ack_listener2(ack_wd2_reset chan<- bool) { 
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10003")
	state.Check_conn(err)
	input, _ := net.ListenUDP("udp", listen_addr)
	state.Check_conn(err)
	defer input.Close()

	buffer := make([]byte, 1024)

	for {
		n,addr,err := input.ReadFromUDP(buffer)
		state.Check_conn(err)
		fmt.Println("Received ",string(buffer[0:n]), " from ",addr)
		if (string(buffer[0:n]) == "Hello") {
			res := "reset"
			buffer = []byte(res)
			ack_wd2_reset <- true
			
		} 

	}
	
}
//Acknowledge is sent evey 500 millisec to two other stations.
//TODO: make ack-listeners less hardcoded?

func Ack_broadcaster() {
	localip := get_localip()

	local_addr, err := net.ResolveUDPAddr("udp", localip + ":0")
	state.Check_conn(err)
	target_addr,err := net.ResolveUDPAddr("udp", ackIP1)
	state.Check_conn(err)
	target_addr2,err := net.ResolveUDPAddr("udp", ackIP2)
	state.Check_conn(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	state.Check_conn(err)
	out_connection2, err := net.DialUDP("udp", local_addr, target_addr2)
	state.Check_conn(err)
	defer out_connection.Close()
	defer out_connection2.Close()
	
	for {
		msg := "Hello"
		buf := []byte(msg)
		_, err := out_connection.Write(buf)
		state.Check_conn(err)
		out_connection2.Write(buf)
		time.Sleep(500 * time.Millisecond)
        }
}

func Ack_watchdog1(ack_wd1_reset <- chan bool) {
	//fmt.Println("Watchdog activated!\n")
	set_alive(true)
	for i := 0; i < 10; i++ {
		select {
		case <- ack_wd1_reset:
			i = 0

		default:
		}
		time.Sleep(500*time.Millisecond)
	}
	set_alive(false)
	fmt.Println("Connection lost.")
}

func Ack_watchdog2(ack_wd2_reset <- chan bool) {
	//fmt.Println("Watchdog activated!\n")
	set_alive(true)
	for i := 0; i < 10; i++ {
		select {
		case <- ack_wd2_reset:
			i = 0

		default:
		}
		time.Sleep(500*time.Millisecond)
	}
	set_alive(false)
	fmt.Println("Connection lost.")
}
