package main

import ("fmt"
	"./order"
	"./driver"
	"./state"
)


func main() {
	stp := [order.FLOORS]int{0,0,0,0}
	ordr := order.Order{Dir: order.MD_Up, Floor: 1}
	elv := order.Elevator{Dir: order.MD_Up, CurrentFloor: 0, Stops: stp}
	fmt.Println("TEST elevator: ", elv)
	a := order.Evaluate (elv, ordr)
	fmt.Println("TEST evaluate: ", a)
	elv.CurrentFloor = 1
	order.Order_complete(elv)
	order.Order_accept(elv, ordr)

	state.LoadState(&elv)
	
	ch_floors := make(chan int)
	ch_buttons := make(chan driver.ButtonEvent)
	ch_obstr   := make(chan bool)
	ch_stop    := make(chan bool)
	ch_sChange := make(chan bool)

	driver.Init("localhost:15657", 4)
	go driver.Button_manager(ch_buttons, ch_sChange, &elv)
	go driver.Event_manager(ch_floors, ch_sChange, &elv)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	go state.SaveState(ch_sChange, &elv)
	
	
	
	for {
		
	}
}
