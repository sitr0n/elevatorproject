package driver

import ("fmt"
	"time"
)
import state "../state"

func timeout_timer(cancel <- chan bool, timeout chan <- bool) {
	for i := 0; i < 10; i++ {
		time.Sleep(500*time.Millisecond)
		select {
		case <- cancel:
			return

		default:
		}
	}
	timeout <- true
}

func Button_manager(b <- chan ButtonEvent, save chan <- bool, e *state.Elevator) {
	for {
		select {
		case event := <- b:
			if (event.Button == BT_Cab) {
				order := Button_event_to_order(event)
				Order_accept(e, order)
				
				fmt.Println("-------------------------------")
				fmt.Println("Button - Order floor: ", event.Floor)
				fmt.Println("Button - Elevator stops: ", e.Stops)
				fmt.Println("Button - Current motor direction: ", e.Dir)
				if (e.Dir == state.MD_Stop) {
					move_to_next_floor(e)
				}
			} else { 
			
			        //network.broadcast_state()
			        //poll states
			        //time.Sleep(100*Millisecond)
				/*
				cost_local := driver.Evaluate(e)
				cost_e2 := driver.Evaluate(e2)
				cost_e3 := driver.Evaluate(e3)
			        Choose_elevator()
				
				if (network.is_Alive)
					wd_timeout := make(chan bool)
					wd_cancel := make(chan bool)
					ack_1 := make(chan bool)
					ack_2 := make(chan bool)
					
					Activate_timeout(ch_cancel, ch_timeout)
					network.Ack_listener1(ack1)
					netowrk.Ack_listener2(ack2)
					select {
					case <- ack_1:
						wd_cancel <- true
						break
					case <- ack_2:
						wd_cancel <- true
						break
					case <- wd_timeout
						Order_accept(e, order)
					}
				else {
					Order_accept(e, order)
				}
				*/
				//TODO: Broadcast corresponding order
				//TODO: Evaluate all elevators and decide which one taking the order
				//TODO: Update corresponding elevator struct -> Stops[event.Floor]
			}
			save <- true
		}
	}
}

func Event_manager(f <- chan int, save chan <- bool, ep *state.Elevator) {
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
					SetMotorDirection(state.MD_Stop)
					SetDoorOpenLamp(true)
					ep.Stops[floor] = 0
					// fmt.Println("Elevator stops: ", ep.Stops)
					time.Sleep(5*time.Second)
					SetDoorOpenLamp(false)
				}
				move_to_next_floor(ep)
			}
			save <- true
			prev_floor = floor
		}
	}
}

func Find_next_stop(e *state.Elevator) state.MotorDirection {
	var direction state.MotorDirection = state.MD_Stop
	if (e.Dir == state.MD_Up) {
		for i := e.CurrentFloor; i < 4; i++ {
			// fmt.Println("Percieved elevator stops: ", e.Stops)
			if (e.Stops[i] > 0) {
				direction = state.MD_Up
				fmt.Println("continuing up")
				return direction
			}
		}
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = state.MD_Down
				fmt.Println("turning down")
				e.Dir = direction
				return direction
			}
		}
	} else {
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = state.MD_Down
				fmt.Println("continuing down")
				return direction
			}
		}
		for i := e.CurrentFloor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = state.MD_Up
				fmt.Println("turning up")
				e.Dir = direction
				return direction
			}
		}
	}
	fmt.Println("No pending orders. Stopping")
	e.Dir = state.MD_Stop
	return direction
}

func move_to_next_floor(elevator *state.Elevator) {
	motor_direction := Find_next_stop(elevator)
	SetMotorDirection(motor_direction)
}
