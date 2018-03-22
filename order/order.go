package order

import state "../state"

const FLOORS = 4
const STOP_WEIGHT = 5


type Order struct {
	Dir state.MotorDirection
	Floor int
}


func Order_complete(e state.Elevator) {
	e.Stops[e.CurrentFloor] = 0
	//fmt.Println(e.Stops)
}

func Order_accept(e state.Elevator, o Order) {
	e.Stops[o.Floor] = 1
	//fmt.Println(e.Stops)
}

func Evaluate(e state.Elevator, o Order) int {
	value := 0
	distance := o.Floor - e.CurrentFloor
	if (distance < 0) {
		distance *= -1
	}
	if (e.Dir == state.MD_Up) {
		if (o.Floor > e.CurrentFloor) {
			for i := e.CurrentFloor; i < o.Floor; i++ {
				value += e.Stops[i] * STOP_WEIGHT
			}
		} else {
			for i := o.Floor; i < FLOORS; i++ {
				value += e.Stops[i] * STOP_WEIGHT
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
				value += e.Stops[i] * STOP_WEIGHT
			}
		} else {
			for i := 0; i < o.Floor; i++ {
				value += e.Stops[i] * STOP_WEIGHT
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

