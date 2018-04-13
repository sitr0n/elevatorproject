package main

import ("fmt"
	"os/exec"
	"time"
)
import ele "./elevator"
import driver "./driver"
import def "./def"
//import network "./network"



func main() {

	fmt.Println("\nCompiled successfully!")
	ele.Init()
	
	for {
	}
	ElevatorServer := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := ElevatorServer.Start()
	def.Check(err)
	time.Sleep(500*time.Millisecond)
	
	var elevator = def.Elevator{}
	var elevator2 = def.Elevator{}
	ele.Load_state(&elevator)
	
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	fmt.Println("    STARTING ELEVATOR     ")
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	
	ch_floors  := make(chan int)
	ch_buttons := make(chan def.ButtonEvent)
	ch_obstr   := make(chan bool)
	ch_stop    := make(chan bool)
	//ch_newstate:= make(chan bool)
	//ch_bcast   := make(chan state.Elevator)
	ch_exit	   := make(chan bool)



	driver.Init("localhost:15657", def.FLOORS)
	go ele.Button_manager(ch_buttons, &elevator)
	go ele.Event_manager(ch_floors, &elevator)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	

	//ch_listen <- true
	for {
		//ch_bcast <- elevator
		fmt.Println("Remote currently: ", elevator2)
		time.Sleep(5*time.Second)
	}
	
	<- ch_exit
}
