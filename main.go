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
	
	
	ch_exit	   := make(chan bool)


	<- ch_exit
}
