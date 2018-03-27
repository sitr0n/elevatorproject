package network

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
	"../state"
)

var _alive bool = false
const stasjon22 string = "129.241.187.56:10001"
const stasjon23 string = "129.241.187.57:10001"

const localIP string =	stasjon22
const targetIP string = stasjon23

func Listener() {

	ch_wd_reset := make(chan bool)
	ch_wd_timeout := make(chan bool)
	
        target_addr, err := net.ResolveUDPAddr("udp", localIP)
    	state.Check(err)

        connection, _ := net.ListenUDP("udp", target_addr)
        state.Check(err)
        defer connection.Close()

        var message state.Elevator

        for {
		
                inputBytes := make([]byte, 4096)
                length, _, _ := connection.ReadFromUDP(inputBytes)
                buffer := bytes.NewBuffer(inputBytes[:length])
                decoder := gob.NewDecoder(buffer)
                decoder.Decode(&message)
                
		ch_wd_reset <- true
		if (get_alive() == false) {
			fmt.Println("A new elevator has joined!")
			set_alive(true)
			go watchdog(ch_wd_reset, ch_wd_timeout)
		}
		
		fmt.Println("Received: ", message)
		buffer.Reset()
	}
}


func watchdog(reset <- chan bool, timeout chan <- bool) {
	fmt.Println("Watchdog activated!\n")
	for i := 0; i < 10; i++ {
		select {
			case <- reset:
				i = 0
			default:
		}
		time.Sleep(1000*time.Millisecond)
	}
	fmt.Println("Connection lost.")
	timeout <- true
	set_alive(false)
}

func set_alive (b bool) {
	_alive = b
}

func get_alive () bool {
	return _alive
}
