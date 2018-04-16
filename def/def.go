package def
import(
		"log"
		"time"
)

var PORT []string =  []string 	{":51299",
				 ":51300"}

var WORKSPACE []string = []string {"0",
				  "129.241.187.140",
				  "0",
				  "129.241.187.150",
				  "129.241.187.141",
				  "129.241.187.143",
				  "129.241.187.146",
				  "129.241.187.154",
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
	WORKSPACES = 23

	STOP_WEIGHT = 5
	
	Ack_order_accept = 2
)

type Elevator struct {
	Dir MotorDirection
	CurrentFloor int
	Stops [4]int
	EMERG_STOP bool
	DOOR_OPEN bool
}

type Order struct {
	Dir MotorDirection
	Floor int
	ID int
	Stamp time.Time
	AddOrRemove AoRdecision
}

type AoRdecision bool //Add or remove decision
const(
	ADD = true
	REMOVE = false
)

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


