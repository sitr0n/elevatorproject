package network
import def "../def"
import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	"os"
	"strconv"
)

const(
	PING 		= false
	Ack_order 	= 0
	Ack_state 	= 1
	Ack_order_accept= 2

	_PING_PERIOD 	= 1000
)

type ack struct {
	received	int
}
type Remote struct {
	id		int
	address  	string
	Alive 		bool
	send		chan interface{}
	Orderchan	chan def.Order
	ackaccept	chan bool
	ackorder	chan bool
	ackstate	chan bool
	Reconnected	chan bool
	state 		def.Elevator
}
var _localip string

func Init(remote_address []string, r *[def.ELEVATORS]Remote) {
	_localip = get_localip()
	ch_order := make(chan def.Order, 100)
	
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].address = ip_address(remote_address[i])
		r[i].id = i
		r[i].Alive = false
		r[i].send = make(chan interface{}, 100)
		r[i].Orderchan = ch_order
		r[i].ackaccept = make(chan bool, 100)
		r[i].ackorder = make(chan bool, 100)
		r[i].ackstate = make(chan bool, 100)
		r[i].Reconnected = make (chan bool, 100)
		
		go r[i].remote_listener()
		go r[i].remote_broadcaster()
	}
	go ping_remotes(r)
}


func (r *Remote) Await_ack(expecting int) bool {
	timeout := make(chan bool)
	timer_cancel := make(chan bool)
	go timeout_timer(timer_cancel, timeout)
	
	switch expecting {
	case Ack_order:
		select {
		case <- r.ackorder:
			timer_cancel <- true
			return true
		
		case <- timeout:
			return false
		}
	case Ack_state:
		select {
		case <- r.ackstate:
			timer_cancel <- true
			return true
		
		case <- timeout:
			return false
		}
	case Ack_order_accept:
		select {
		case <- r.ackaccept:
			timer_cancel <- true
			return true
		
		case <- timeout:
			return false
		}
	default:
	}
	return false
}

func Broadcast_state(e *def.Elevator, r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		go r[i].Send_state(e)
	}
}

func Broadcast_order(order def.Order, r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		go r[i].Send_order(order)
	}
}

func (r *Remote) Get_state() def.Elevator {
	return r.state
}

func (r *Remote) Send_order(order def.Order) {
	for {
		r.send <- order
		received := r.Await_ack(Ack_order)
		if (received == true || r.Alive == false) {
			break
		}
	}
}

func (r *Remote) Send_state(state *def.Elevator) {
	for {
		r.send <- state
		fmt.Println("Sending state:", state)
		received := r.Await_ack(Ack_state)
		if (received == true || r.Alive == false) {
			break
		}
	}
}

func Send_state_to_all(e *def.Elevator, r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		go r[i].Send_state(e)
	}
}

func (r *Remote) Send_ack(msg int) {
	switch msg {
	case 0:
		fmt.Println("Sending state received")
	case 1:
		fmt.Println("Sending order received")
	case 2:
		fmt.Println("Sending order taken!!")
	}
	var response ack = ack{msg}
	r.send <- response
}

func Send_ack(msg int, r *[def.ELEVATORS]Remote) {
	for i := 0; i < def.ELEVATORS; i++ {
		r[i].Send_ack(msg)
	}
}

func (r *Remote) remote_listener() {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + def.PORT[r.id])
	def.Check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	def.Check(err)
	defer in_connection.Close()
	
	var state def.Elevator
	var order def.Order
	var ack ack
	const PING_SIZE = 5
	const ACK_SIZE = 2
	const STATE_SIZE1 = 79
	const STATE_SIZE2 = 80
	const STATE_SIZE3 = 81
	
	wd_kick := make(chan bool, 100)
	for {
		buffer := make([]byte, 1024)
		length, _, _ := in_connection.ReadFromUDP(buffer)
		if (r.Alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection with remote", r.id, "established!")
			r.Reconnected <- true
		}
		wd_kick <- true
		
		switch length {
		case PING_SIZE:
		
		case ACK_SIZE:
			err := json.Unmarshal(buffer[:length], &ack)
			def.Check(err)
			switch ack.received {
			case Ack_order:
				r.ackorder <- true
			case Ack_state:
				fmt.Println("Remote got my state")
				r.ackstate <- true
			case Ack_order_accept:
				fmt.Println("Remote", r.id, "has confirmed the order!\n")
				r.ackaccept <- true
			default:
			}
			
		case STATE_SIZE1:
			err := json.Unmarshal(buffer[:length], &state)
			def.Check(err)
			r.state = state
			fmt.Println("Received STATE:", state)
			r.Send_ack(Ack_state)
	
		case STATE_SIZE2:
			err := json.Unmarshal(buffer[:length], &state)
			def.Check(err)
			r.state = state
			fmt.Println("Received STATE:", state)
			r.Send_ack(Ack_state)
			
		case STATE_SIZE3:
			err := json.Unmarshal(buffer[:length], &state)
			def.Check(err)
			r.state = state
			fmt.Println("Received STATE:", state)
			r.Send_ack(Ack_state)
		
		default:
			fmt.Println("Received size:", length)
			err := json.Unmarshal(buffer[:length], &order)
			def.Check(err)
			r.Orderchan <- order
			r.Send_ack(Ack_order)
		}
	}
}

func (r *Remote) remote_broadcaster() {
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
		}
	}
}

func ping_remotes(remote *[def.ELEVATORS]Remote) {
	for {
		time.Sleep(time.Duration(_PING_PERIOD)*time.Millisecond)
		for i := 0; i < def.ELEVATORS; i++ {
			remote[i].send <- PING
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

func flush_channel(c <- chan interface{}) {
	for i := 0; i < 100; i++ {
		select {
		case <- c:
		default:
		}
	}
}

func (r *Remote) watchdog(kick <- chan bool) {
	r.Alive = true
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(_PING_PERIOD)*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	r.Alive = false
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
	case string:
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

