package order

//import "fmt"

const FLOORS = 4
//const STOP_WEIGHT = 2

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type Order struct {
	Dir MotorDirection
	Floor int
}

type Elevator struct {
	Dir MotorDirection
	CurrentFloor int
	Stops [FLOORS]int
}

func Order_complete(e Elevator) {
	e.Stops[e.CurrentFloor] = 0
	//fmt.Println(e.Stops)
}

func Order_accept(e Elevator, o Order) {
	e.Stops[o.Floor] = 1
	//fmt.Println(e.Stops)
}

func Evaluate(e Elevator, o Order) int {
	value := 0
	distance := o.Floor - e.CurrentFloor
	if (distance < 0) {
		distance *= -1
	}
	if (e.Dir == MD_Up) {
		if (o.Floor > e.CurrentFloor) {
			for i := e.CurrentFloor; i < o.Floor; i++ {
				value += 2*e.Stops[i]
			}
		} else {
			for i := o.Floor; i < FLOORS; i++ {
				value += 2*e.Stops[i]
			}
		}
		if (e.Dir != o.Dir) {
			if (o.Floor > e.CurrentFloor) {
				value += 2*(FLOORS - o.Floor)
			} else {
				value += 2*(FLOORS - e.CurrentFloor)
			}
			value += distance
		} else {
			value += distance
		}
	} else {
		if (o.Floor < e.CurrentFloor) {
			for i := o.Floor; i < e.CurrentFloor; i++ {
				value += 2*e.Stops[i]
			}
		} else {
			for i := 0; i < o.Floor; i++ {
				value += 2*e.Stops[i]
			}
		}
		if (e.Dir != o.Dir) {
			if (o.Floor > e.CurrentFloor) {
				value += 2*e.CurrentFloor
			} else {
				value += 2*o.Floor
			}
			value += distance
		} else {
			value += distance
		}
	}
		
	return value
}

