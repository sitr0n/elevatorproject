package driver

import "fmt"

func FSM(f <- chan int) {
	for {
		select {
		case floor := <- f
			
			
		}
	}
}