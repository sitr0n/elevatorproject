package main

import ("fmt"
	//"./order"
	"./driver"
	"./state"
	"os/exec"
	"time"
	//"./bcast"
	"./network"
)


func main() {

	fmt.Println("\nCompiled successfully!")
	
	network.Init(10, 11)
	for {
	}
	ElevatorServer := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := ElevatorServer.Start()
	state.Check(err)
	time.Sleep(500*time.Millisecond)
	
	var elevator = state.Elevator{}
	var elevator2 = state.Elevator{}
	var elevator3 = state.Elevator{}
	state.Load(&elevator)
	
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	fmt.Println("    STARTING ELEVATOR     ")
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	
	ch_floors  := make(chan int)
	ch_buttons := make(chan driver.ButtonEvent)
	ch_obstr   := make(chan bool)
	ch_stop    := make(chan bool)
	ch_newstate:= make(chan bool)
	ch_bcast   := make(chan state.Elevator)
	//ch_listen  := make(chan bool)
	ch_exit	   := make(chan bool)
	ack_wd1_reset   := make(chan bool)
	ack_wd2_reset  := make(chan bool)


	driver.Init("localhost:15657", 4)
	go driver.Button_manager(ch_buttons, ch_newstate, &elevator)
	go driver.Event_manager(ch_floors, ch_newstate, &elevator)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	
	go state.Save(ch_newstate, &elevator)
	
	go network.Broadcast_state(ch_bcast)
	go network.Poll_remote_state(&elevator2)
	go network.Poll_remote_state2(&elevator3)
	go network.Ack_listener1(ack_wd1_reset)
	go network.Ack_listener2(ack_wd2_reset)
	go network.Ack_broadcaster()
	//ch_listen <- true
	for {
		ch_bcast <- elevator
		fmt.Println("Remote currently: ", elevator2)
		time.Sleep(5*time.Second)
	}
	
	<- ch_exit
}
