package network

import (
	"fmt"
	"net"
	"time"
	"encoding/json"
)

import unsafe "unsafe"
import def "../def"


type Remote struct {
	id		int
	input		*net.UDPConn
	output 		*net.UDPConn
	address  	def.IP
	alive 		bool
	Send		chan interface{}
	Receive		chan interface{}
	State 		def.Elevator
}

var _localip string

func Init(first_remote interface{}, second_remote interface{}, r *[def.ELEVATORS]Remote, ch_order chan <- def.Order, ch_ack chan <- bool) {
	_localip = get_localip()
	r[0].address = ip_address(first_remote)
	r[1].address = ip_address(second_remote)
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].id = i
		r[i].alive = false
		r[i].Send = make(chan interface{})
		r[i].Receive = make(chan interface{})
		
		connect_remote(&r[i])
		
		go r[i].remote_listener(ch_order, ch_ack)
		go remote_broadcaster(r[i].output, r[i].Send)
	}
}



func (r Remote) remote_listener(ch_order chan <- def.Order, ch_ack chan <- bool) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + def.PORT[r.id])
	def.Check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	def.Check(err)
	defer in_connection.Close()
	
	var elevator def.Elevator
	var order def.Order
	var ack bool = false
	const STATE_SIZE = int(unsafe.Sizeof(elevator))
	const ORDER_SIZE = int(unsafe.Sizeof(order))
	const ACK_SIZE = int(unsafe.Sizeof(ack))
	
	wd_kick := make(chan bool)
	inputBytes := make([]byte, 4096)
	
	fmt.Println("Starting remote", r.id, "listener!")
	for {
		length, _, _ := r.input.ReadFromUDP(inputBytes)
		if (r.alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection with remote", r.id, "established!")
		}
		wd_kick <- true
		
		switch length {
		case ACK_SIZE:
			err := json.Unmarshal(inputBytes[:length], &ack)
			def.Check(err)
			ch_ack <- ack
			break
			
		case ORDER_SIZE: 
			err := json.Unmarshal(inputBytes[:length], &order)
			def.Check(err)
			ch_order <- order
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

func remote_broadcaster(connection *net.UDPConn, message <- chan interface{}) {
	fmt.Println("Starting remote bcaster")
	for {
		select {
		case msg := <- message:
			encoded, err := json.Marshal(msg)
			def.Check(err)
			
			connection.Write(encoded)
			fmt.Println("Wrote: ", msg)
		}
	}
}

func (r Remote) watchdog(kick <- chan bool) {
	r.alive = true
	for i := 0; i < 10; i++ {
		time.Sleep(50*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	r.alive = false
}

func connect_remote(r *Remote) {
	
	
	local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	def.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", string(r.address) + def.PORT[r.id])
	def.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	def.Check(err)
	defer out_connection.Close()
	
}


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

