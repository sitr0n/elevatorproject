package main

import ("fmt"
	"./order"
	"./driver"
)

func PollButtons() {
	// IF BUTTONPRESS
	//	Broadcast corresponding order
	//	v1 = order.Evaluate(elev1, order)
	//	v2 = order.Evaluate(elev2, order)
	//	v3 = order.Evaluate(elev3, order)
	//	update elevN.Stops[order.Floor]
}

func PollFloorSensor() {
	// IF SENSOR
	//	IF elev.Stops[Floor] == 1 || elev.Floor == 4
	//		stop()
	//	IF elev.Direction == UP
	//		for i := elev.Floor; i < FLOORS; i++
	//			if elev.Stops[i] == 1
	//				continue() / break
	//			
}

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
	go driver.Event_manager(ch_floors, &elv)
	go driver.PollButtons(ch_buttons)
	go driver.PollFloorSensor(ch_floors)
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)
	
	
}