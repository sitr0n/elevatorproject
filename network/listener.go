package network

import ("bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
	"../state"
)
var alive bool = false
const stasjon22 string = "129.241.187.56:0"
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
		fmt.Println("Waiting for data...")
                length, _, _ := connection.ReadFromUDP(inputBytes)
                buffer := bytes.NewBuffer(inputBytes[:length])
                decoder := gob.NewDecoder(buffer)
                decoder.Decode(&message)
                
		
		if (!alive) {
			fmt.Println("A new elevator has joined!")
			alive = true
			go watchdog(ch_wd_reset, ch_wd_timeout, &alive)
		} else {
			ch_wd_reset <- true
		}
		fmt.Println("Found: ", message)
	}
}

func watchdog(reset <- chan bool, timeout chan <- bool, is_responsive *bool) {
	for i := 0; i < 10; i++ {
		select {
			case <- reset:
				i = 0
			default:
		}
		time.Sleep(500*time.Millisecond)
	}
	fmt.Println("Connection lost.")
	timeout <- true
	*is_responsive = false
}
