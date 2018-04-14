package network

import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	"os"
	"strconv"
)

import unsafe "unsafe"
import def "../def"


type Remote struct {
	id		int
	input		*net.UDPConn
	output 		*net.UDPConn
	address  	string
	Alive 		bool
	send		chan interface{}
	Orderchan	chan def.Order
	Ackchan		chan bool
	State 		def.Elevator
}

var _localip string

func Init(remote_address []string, r *[def.ELEVATORS]Remote) {
	_localip = get_localip()

	ch_ack := make(chan bool)
	ch_order := make(chan def.Order)
	
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].address = ip_address(remote_address[i])
		r[i].id = i
		r[i].Alive = false
		r[i].send = make(chan interface{})
		r[i].Orderchan = ch_order
		r[i].Ackchan = ch_ack
		
		r[i].connect_remote()
		
		go r[i].remote_listener(r[i].Orderchan, r[i].Ackchan)
		go r[i].remote_broadcaster(r[i].send)
	}
	go send_ping(r)
}


func Await_ack(remote *[def.ELEVATORS]Remote) bool {
	timeout := make(chan bool)
	timer_cancel := make(chan bool)
	go timeout_timer(timer_cancel, timeout)
	select {
	case <- remote[0].Ackchan:
		timer_cancel <- true
		return true
	
	case <- timeout:
		return false
	}
}

func Broadcast_state(r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].send <- r[i].State
	}
}

func Broadcast_order(order def.Order, r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].send <- order
	}
}

func (r Remote) Get_state() def.Elevator {
	return r.State
}

func (r Remote) Send_order(order def.Order) {
	r.send <- order
}

func (r Remote) Send_state() {
	r.send <- r.State
}

func Send_ack(r [def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++{
		r[i].send <- true
	}
}

func (r Remote) Send_ping() {
	r.send <- false
}

func (r Remote) remote_listener(add_order chan <- def.Order, ch_ack chan <- bool) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + def.PORT[r.id])
	def.Check(err)
	connection, _ := net.ListenUDP("udp", listen_addr)
	def.Check(err)
	defer connection.Close()
	
	
	r.input = connection
	
	fmt.Println("Device ", r.id , " connected!\n", "Input: ", r.input)

	var elevator def.Elevator
	var order def.Order
	var ack bool = false
	const STATE_SIZE = int(unsafe.Sizeof(elevator))
	const ORDER_SIZE = int(unsafe.Sizeof(order))
	const ACK_SIZE = int(unsafe.Sizeof(ack))
	wd_kick := make(chan bool)
	inputBytes := make([]byte, 4096)
	
	fmt.Println("Starting remote", r.id, "listener!\n", "Input: ", r.input)
	for {
		fmt.Println("this works...")
		length, _, _ := r.input.ReadFromUDP(inputBytes)
		wd_kick <- true
		fmt.Println("Received something!\n\n")
		if (r.Alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection established!")
		}
		
		switch length {
		case ACK_SIZE:
			err := json.Unmarshal(inputBytes[:length], &ack)
			def.Check(err)
			if (ack == true) {
				ch_ack <- ack
			}
			break
			
		case ORDER_SIZE: 
			err := json.Unmarshal(inputBytes[:length], &order)
			def.Check(err)
			add_order <- order
			break
			
		case STATE_SIZE:
			err := json.Unmarshal(inputBytes[:length], &elevator)
			def.Check(err)
			
			r.State = elevator
			//ch_state <- state
			break
		
		default:
			fmt.Println("Oops! Received something unexpected from remote", r.id)
		}
	}
}

func send_ping(remote *[def.ELEVATORS]Remote) {
	for {
		time.Sleep(200*time.Millisecond)
		for i := 0; i < def.ELEVATORS; i++ {
			remote[i].send <- false
		}
	}
}

func timeout_timer(cancel <- chan bool, timeout chan <- bool) {
	for i := 0; i < 10; i++ {
		time.Sleep(500*time.Millisecond)
		select {
		case <- cancel:
			return

		default:
		}
	}
	timeout <- true
}


func (r Remote) remote_broadcaster(message <- chan interface{}) {
	fmt.Println("Starting remote bcaster")
	local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	def.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", string(r.address) + def.PORT[r.id])
	def.Check(err)
	connection, err := net.DialUDP("udp", local_addr, target_addr)
	def.Check(err)
	defer connection.Close()

	r.output = connection

	for {
		select {
		case msg := <- message:
			encoded, err := json.Marshal(msg)
			def.Check(err)
			
			connection.Write(encoded)
			//fmt.Println("Wrote: ", msg)
		}
	}
}

func (r Remote) watchdog(kick <- chan bool) {
	r.Alive = true
	for i := 0; i < 10; i++ {
		time.Sleep(50*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	r.Alive = false
}

func (r Remote) connect_remote() {
	
}

func ip_address(adr string) string {
	var is_int = true
	for _, char := range adr {
		if (char == '.') {
			is_int = false
		}
	}
	if (is_int == true) {
		index, err := strconv.Atoi(adr)
		if (err != nil) {
			fmt.Println("Argument is invalid, try another.")
			os.Exit(2)
		}
		if (index > def.WORKSPACES || index < 1) {
			fmt.Println("An argument is out of bounds. Please try another number or target IP address.")
			os.Exit(2)
		}
		return def.WORKSPACE[index]
	} else {
		return adr
	}
}

/*
func ip_address(adr interface{}) def.IP {
	switch a := adr.(type) {
	case def.IP:
		return a
	case int:
		if (a > 23 || a < 0) {
			fmt.Println("Workspace index is out of bounds. Please abort process and try another argument!")
			for {
			}
		} else {
			return def.WORKSPACE[a]
		}
	default:
		fmt.Println("Wrong data type passed to network.Init. Try string or workspace number.")
		return "0"
	}
}
*/

func get_localip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	def.Check(err)
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

