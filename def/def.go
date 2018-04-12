package def

import "log"

const PORT string = ":10001"

type IP string
var WORKSPACE []IP = []IP	 {"0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "0",
				  "129.241.187.158",
				  "129.241.187.159",
				  "129.241.187.144",
				  "129.241.187.152",
				  "129.241.187.142",
				  "129.241.187.148",
				  "129.241.187.147",
				  "129.241.187.145",
				  "0",
				  "0",
				  "129.241.187.155",
				  "0",
				  "129.241.187.56",
				  "129.241.187.57"}

const (
	ELEVATORS = 2
	FLOORS = 4
	
	STOP_WEIGHT = 5
)

type Elevator struct {
	Dir MotorDirection
	CurrentFloor int
	Stops [4]int
}

type Order struct {
	Dir MotorDirection
	Floor int
}

type ButtonType int
const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type MotorDirection int
const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

func Check(e error) {
	if e != nil {
		log.Print(e)
		//continue
	}
}
