package driver

import "fmt"


func FSM(f <- chan int, e Elevator) {
	prev_floor := -1
	for {
		select {
		case floor := <- f
			if (floor != prev_floor) {
				if (
			}
			
			
		}
	}
}
