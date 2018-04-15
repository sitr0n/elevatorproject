package network

import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	"os"
	"strconv"
	//"log"
	//"runtime"
)

//import unsafe "unsafe"
import def "../def"


type Remote struct {
	id		int
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
		
		go r[i].remote_listener()
		go r[i].remote_broadcaster()
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

func (r *Remote) Get_state() def.Elevator {
	return r.State
}

func (r *Remote) Send_order(order def.Order) {
	r.send <- order
}

func (r *Remote) Send_state() {
	r.send <- r.State
}

func Send_ack(r [def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].send <- true
	}
}

func (r *Remote) Send_ping() {
	r.send <- false
}

func (r *Remote) Set_alive(a bool) {
	r.Alive = a
}

func (r *Remote) remote_listener() {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + def.PORT[r.id])
	def.Check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	def.Check(err)
	defer in_connection.Close()
	
	var elevator def.Elevator
	var order def.Order
	var ack bool = false
	const STATE_SIZE = 44
	const ACK_SIZE = 4
	const PING_SIZE = 5
	
	wd_kick := make(chan bool, 100)
	
	fmt.Println("Starting remote", r.id, "listener!\n")
	for {
		buffer := make([]byte, 1024)
		length, _, _ := in_connection.ReadFromUDP(buffer)
		if (r.Alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection with remote", r.id, "established!")
		}
		wd_kick <- true
		
		switch length {
		case ACK_SIZE:
			err := json.Unmarshal(buffer[:length], &ack)
			def.Check(err)
			r.Ackchan <- true
			break
			
		case PING_SIZE:
			break
			
		case STATE_SIZE:
			err := json.Unmarshal(buffer[:length], &elevator)
			def.Check(err)
			
			r.State = elevator
			
			fmt.Println("STATE: ", r.State)
			break
		
		default:
			err := json.Unmarshal(buffer[:length], &order)
			def.Check(err)
			r.Orderchan <- order
		}
	}
}

func send_ping(remote *[def.ELEVATORS]Remote) {
	for {
		time.Sleep(time.Second)
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


func (r *Remote) remote_broadcaster() {
	fmt.Println("Starting remote bcaster")
	//local_addr, err := net.ResolveUDPAddr("udp", _localip + ":0")
	//def.Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", r.address + def.PORT[r.id])
	def.Check(err)
	out_connection, err := net.DialUDP("udp", nil, target_addr)
	def.Check(err)
	defer out_connection.Close()


	for {
		select {
		case msg := <- r.send:
			encoded, err := json.Marshal(msg)
			def.Check(err)
			
			out_connection.Write(encoded)
			fmt.Println("Wrote: ", msg, "to", r.address + def.PORT[r.id])
		}
	}
}

func flush_channel(c <- chan interface{}) {
	for i := 0; i < 100; i++ {
		select {
		case <- c:
		default:
		}
	}
}

func (r *Remote) watchdog(kick <- chan bool) {
	r.Set_alive(true)
	fmt.Println("Watchdog is UP")
	for i := 0; i < 10; i++ {
		time.Sleep(5000*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	r.Set_alive(false)
	fmt.Println("Connection with remote", r.id, "lost.")
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

