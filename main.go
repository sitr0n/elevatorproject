package main

import ("fmt"
	//"./order"
	"./driver"
	"./state"
	"os/exec"
	"time"
	"./bcast"
)


func main() {
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := cmd.Start()
	state.Check(err)
	
	time.Sleep(200*time.Millisecond)
	var elevator = state.Elevator{}
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
	ch_sChange := make(chan bool)
	ch_exit	   := make(chan bool)

	driver.Init("localhost:15657", 4)
	go driver.Button_manager(ch_buttons, ch_sChange, &elevator)
	go driver.Event_manager(ch_floors, ch_sChange, &elevator)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	
	go state.Save(ch_sChange, &elevator)
	
	for {
		bcast.Broadcast(&elevator)
		time.Sleep(5*time.Second)
	}
	
	<- ch_exit
}
