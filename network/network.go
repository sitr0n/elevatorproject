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
	Alive 		bool
	send		chan interface{}
	Orderchan	chan def.Order
	Ackchan		chan bool
	State 		def.Elevator
}

var _localip string

func Init(first_remote interface{}, second_remote interface{}, r *[def.ELEVATORS]Remote) {
	_localip = get_localip()
	r[0].address = ip_address(first_remote)
	r[1].address = ip_address(second_remote)
	
	ch_ack := make(chan bool)
	ch_order := make(chan def.Order)
	
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].id = i
		r[i].Alive = false
		r[i].send = make(chan interface{})
		r[i].Orderchan = ch_order
		r[i].Ackchan = ch_ack
		
		connect_remote(&r[i])
		
		go remote_listener(&r[i], r[i].Orderchan, r[i].Ackchan)
		go remote_broadcaster(r[i].output, r[i].send)
	}
	go send_ping(r)
}



func remote_listener(r *Remote, add_order chan <- def.Order, ch_ack chan <- bool) {
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
		wd_kick <- true
		if (r.Alive == false) {
			go watchdog(r, wd_kick)
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
		time.Sleep(100*time.Millisecond)
		for i := 0; i < def.ELEVATORS; i++ {
			remote[i].send <- false
		}
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

func (r Remote) Send_ack() {
	r.send <- true
}

func (r Remote) Send_ping() {
	r.send <- false
}

func remote_broadcaster(connection *net.UDPConn, message <- chan interface{}) {
	fmt.Println("Starting remote bcaster")
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

func watchdog(r *Remote, kick <- chan bool) {
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

func connect_remote(r *Remote) {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + def.PORT[r.id])
	def.Check(err)
	in_connection, _ := net.ListenUDP("udp", listen_addr)
	def.Check(err)
	defer in_connection.Close()
	
	local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	def.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", string(r.address) + def.PORT[r.id])
	def.Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	def.Check(err)
	defer out_connection.Close()
	
	r.input = in_connection
	r.output = out_connection
	fmt.Println("Device ", r.id , " connected!")
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

