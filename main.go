package main

import ("fmt"
	"./order"
	"./driver"
)


func main() {
	stp := [order.FLOORS]int{0,1,1,0}
	ordr := order.Order{Dir: false, Floor: 1}
	elv := order.Elevator{Dir: true, CurrentFloor: 0, Stops: stp}
	fmt.Println(elv)
	a := order.Evaluate (elv, ordr)
	fmt.Println(a)
	elv.CurrentFloor = 1
	order.Order_complete(elv)
	order.Order_accept(elv, ordr)
	
	ch_floors := make(chan int)
	ch_buttons := make(chan driver.ButtonEvent)
	ch_obstr   := make(chan bool)
	ch_stop    := make(chan bool)

	driver.Init("localhost:15657", 4)
	go driver.Button_manager(ch_buttons, &elv)
	go driver.Event_manager(ch_floors, elv, &elv)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	
	for {
		
	}
	
}