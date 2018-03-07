package main

import "fmt"

const FLOORS = 4
const STOP_WEIGHT = 2



type order struct {
	Dir bool
	Floor int
}

type elevator struct {
	Dir bool
	CurrentFloor int
	stops [FLOORS]int
}

func order_complete(e elevator) {
	e.stops[e.CurrentFloor] = 0
	fmt.Println(e.stops)
}

func order_accept(e elevator, o order) {
	e.stops[o.Floor] = 1
	fmt.Println(e.stops)
}

func evaluate(e elevator, o order) int {
	value := 0
	distance := o.Floor - e.CurrentFloor
	if (distance < 0) {
		distance *= -1
	}
	if (e.Dir == true) {
		if (o.Floor > e.CurrentFloor) {
			for i := e.CurrentFloor; i < o.Floor; i++ {
				value += 2*e.stops[i]
			}
		} else {
			for i := o.Floor; i < FLOORS; i++ {
				value += 2*e.stops[i]
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
				value += 2*e.stops[i]
			}
		} else {
			for i := 0; i < o.Floor; i++ {
				value += 2*e.stops[i]
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

func main() {
	stp := [FLOORS]int{0,1,1,0}
	ordr := order{Dir: false, Floor: 1}
	elv := elevator{Dir: true, CurrentFloor: 0, stops: stp}
	fmt.Println(elv)
	a := evaluate (elv, ordr)
	fmt.Println(a)
	elv.CurrentFloor = 1
	order_complete(elv)
	order_accept(elv, ordr)
}