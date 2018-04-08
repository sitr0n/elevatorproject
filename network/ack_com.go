package network

import (
	"fmt"
	"net"
	"time"
	"../state"
	
)

const ack_stasjon10 string = "129.241.187.158:10002"
const ack_stasjon11 string = "129.241.187.159:10002"
const ackIP1 string = ack_stasjon10
const ackIP2 string = ack_stasjon11


func Ack_listener1(ack_wd1_reset chan<- bool) { 
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10002")
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

func Ack_broadcast() {
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
