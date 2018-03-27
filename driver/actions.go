package driver

import ("fmt"
	"time"
)
import m "../state"


func Button_manager(b <- chan ButtonEvent, save chan <- bool, e *m.Elevator) {
	for {
		select {
		case event := <- b:
			if (event.Button == BT_Cab) {
				var order = Button_event_to_order(event)
				Order_accept(e, order)
				
				fmt.Println("-------------------------------")
				fmt.Println("Button - Order floor: ", event.Floor)
				fmt.Println("Button - Elevator stops: ", e.Stops)
				fmt.Println("Button - Current motor direction: ", e.Dir)
				if (e.Dir == m.MD_Stop) {
					motor_direction := Find_next_stop(e)
					SetMotorDirection(motor_direction)
				}
			} else {
				//TODO: Broadcast corresponding order
				//TODO: Evaluate all elevators and decide which one taking the order
				//TODO: Update corresponding elevator struct -> Stops[event.Floor]
			}
			save <- true
		}
	}
}

func Event_manager(f <- chan int, save chan <- bool, ep *m.Elevator) {
	prev_floor := -1
	for {
		select {
		case floor := <- f:
			if (floor != prev_floor) {
				SetFloorIndicator(floor)
				ep.CurrentFloor = floor
				fmt.Println("Current floor: ", ep.CurrentFloor)
				if (ep.Stops[floor] > 0) {
					fmt.Println("Stopping at floor ", floor)
					SetMotorDirection(m.MD_Stop)
					SetDoorOpenLamp(true)
					ep.Stops[floor] = 0
					// fmt.Println("Elevator stops: ", ep.Stops)
					time.Sleep(5*time.Second)
					SetDoorOpenLamp(false)
				}
				motor_direction := Find_next_stop(ep)
				SetMotorDirection(motor_direction)
			}
			save <- true
			prev_floor = floor
		}
	}
}

func Find_next_stop(e *m.Elevator) m.MotorDirection {
	var direction m.MotorDirection = m.MD_Stop
	if (e.Dir == m.MD_Up) {
		for i := e.CurrentFloor; i < 4; i++ {
			// fmt.Println("Percieved elevator stops: ", e.Stops)
			if (e.Stops[i] > 0) {
				direction = m.MD_Up
				fmt.Println("continuing up")
				return direction
			}
		}
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = m.MD_Down
				fmt.Println("turning down")
				e.Dir = direction
				return direction
			}
		}
	} else {
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = m.MD_Down
				fmt.Println("continuing down")
				return direction
			}
		}
		for i := e.CurrentFloor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = m.MD_Up
				fmt.Println("turning up")
				e.Dir = direction
				return direction
			}
		}
	}
	fmt.Println("No pending orders. Stopping")
	e.Dir = m.MD_Stop
	return direction
}