package main

import ("fmt"
	//"./order"
	"./driver"
	"./state"
	"os/exec"
	"time"
)


func main() {
	elevator := state.Elevator{}
	state.Load(&elevator)
	
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := cmd.Start()
	state.Check(err)
	
	fmt.Println("Starting ElevatorServer...")
	time.Sleep(1*time.Second)
	
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
	
	<- ch_exit
}
