package main

import ("fmt"
	"os/exec"
	"time"
)
import elevator "./elevator"
//import driver "./driver"
import def "./def"
//import network "./network"



func main() {

	ElevatorServer := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := ElevatorServer.Start()
	def.Check(err)
	time.Sleep(500*time.Millisecond)
	
	elevator.Init()
	
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	fmt.Println("    STARTING ELEVATOR     ")
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	
<<<<<<< HEAD
	
	ch_exit	   := make(chan bool)


=======
	ch_floors  := make(chan int)
	ch_buttons := make(chan def.ButtonEvent)
	ch_obstr   := make(chan bool)
	ch_stop    := make(chan bool)
	//ch_newstate:= make(chan bool)
	//ch_bcast   := make(chan state.Elevator)
	ch_add_order := make(chan def.Order)
	ch_remove_order := make(chan def.Order)
	ch_exit	   := make(chan bool)



	driver.Init("localhost:15657", def.FLOORS)
	go ele.Button_manager(ch_buttons, &elevator)
	go ele.Event_manager(ch_floors, &elevator)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	go ele.Order_queue(ch_add_order, ch_remove_order)

	//ch_listen <- true
	for {
		//ch_bcast <- elevator
		fmt.Println("Remote currently: ", elevator2)
		time.Sleep(5*time.Second)
	}
	
>>>>>>> eac38eb2964cb053276c9ef231e42d7154e1e697
	<- ch_exit
}
