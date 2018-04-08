package main

import (//"bytes"
	//"encoding/json"
	"fmt"
	"net"
	"time"
	//"../state"
	//"log"
)

//var _alive bool = false

const ack_stasjon10 string = "129.241.187.158:10002"
const ack_stasjon11 string = "129.241.187.159:10002"


const ackIP1 string = ack_stasjon11
const ackIP2 string = ack_stasjon11

func Check(e error) {
	if e != nil {
		fmt.Println("Connection error")
	}
}

func Ack_listener(reset chan<- bool) { 
	localip := get_localip()

	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":10002")
	Check(err)
	input, _ := net.ListenUDP("udp", listen_addr)
	Check(err)
	defer input.Close()

	buffer := make([]byte, 1024)
	i := 0
	for {
		n,addr,err := input.ReadFromUDP(buffer)
		Check(err)
		fmt.Println("Received ",string(buffer[0:n]), " from ",addr)
		if (string(buffer[0:n]) == "Hello") {
			fmt.Println("inside correct loop")
			res := "reset"
			buffer = []byte(res)
			fmt.Println(string(buffer[0:n])," ", i)
			i++
			reset <- true
			
		} 

	}
	
}


func Ack_broadcast() {
	localip := get_localip()

	local_addr, err := net.ResolveUDPAddr("udp", localip + ":0")
	Check(err)
	target_addr,err := net.ResolveUDPAddr("udp", ackIP1)
	Check(err)
	//target_addr2,err := net.ResolveUDPAddr("udp", ackIP2)
	//Check(err)
	out_connection, err := net.DialUDP("udp", local_addr, target_addr)
	Check(err)
	//out_connection2, err := net.DialUDP("udp", local_addr, target_addr2)
	//Check(err)
	defer out_connection.Close()
	//defer out_connection2.Close()
	
	for {
		msg := "Hello"
		buf := []byte(msg)
		_, err := out_connection.Write(buf)
		Check(err)
		//out_connection2.Write(buf)
		
		time.Sleep(500 * time.Millisecond)
        }
}

func get_localip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	Check(err)
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

func watchdog(reset <- chan bool) {
	//fmt.Println("Watchdog activated!\n")
	//set_alive(true)
	for i := 0; i < 10; i++ {
		select {
		case <- reset:
			i = 0

		default:
		}
		time.Sleep(500*time.Millisecond)
	}
	//set_alive(false)
	fmt.Println("Connection lost.")
}

func main() {

	reset := make(chan bool)
	exit := make(chan bool)
	go Ack_broadcast()
	go Ack_listener(reset)
	go watchdog(reset)

	<- exit
}











