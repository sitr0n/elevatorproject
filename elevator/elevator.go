package elevator

import ("fmt"
	"encoding/json"
	"io/ioutil"
	"time"
)

import driver "../driver"
import network "../network"
import def "../def"


func Init() {
	//var elevator = def.Elevator{}
	var remote [def.FLOORS]network.Remote
	ch_ack := make(chan bool)
	ch_order := make(chan def.Order)
	network.Init(10, 11, &remote, ch_order, ch_ack)
	
	fmt.Println(remote[0])
}

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

func Button_manager(b <- chan def.ButtonEvent, e *def.Elevator) {
	for {
		select {
		case event := <- b:
			order := button_event_to_order(event)
			if (event.Button == def.BT_Cab) {
				Order_accept(e, order)
				
				fmt.Println("-------------------------------")
				fmt.Println("Button - Order floor: ", event.Floor)
				fmt.Println("Button - Elevator stops: ", e.Stops)
				fmt.Println("Button - Current motor direction: ", e.Dir)
				if (e.Dir == def.MD_Stop) {
					move_to_next_floor(e)
				}
			} else { 
				//bcast <- order
				//local_cost := Evaluate(*e, order)
				//remote1 := network.Get_remote(0)
				//remote2 := network.Get_remote(2)
				
				//cost1 := Evaluate(remote1.state, order)
				//cost2 := Evaluate(remote2.state, order)
				
			
				
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
			Save_state(e)
		}
	}
}

func Event_manager(f <- chan int, ep *def.Elevator) {
	prev_floor := -1
	for {
		select {
		case floor := <- f:
			if (floor != prev_floor) {
				driver.SetFloorIndicator(floor)
				ep.CurrentFloor = floor
				fmt.Println("Current floor: ", ep.CurrentFloor)
				if (ep.Stops[floor] > 0) {
					fmt.Println("Stopping at floor ", floor)
					driver.SetMotorDirection(def.MD_Stop)
					driver.SetDoorOpenLamp(true)
					ep.Stops[floor] = 0
					// fmt.Println("Elevator stops: ", ep.Stops)
					time.Sleep(5*time.Second)
					driver.SetDoorOpenLamp(false)
				}
				move_to_next_floor(ep)
			}
			Save_state(ep)
			prev_floor = floor
		}
	}
}

func Find_next_stop(e *def.Elevator) def.MotorDirection {
	var direction def.MotorDirection = def.MD_Stop
	if (e.Dir == def.MD_Up) {
		for i := e.CurrentFloor; i < 4; i++ {
			// fmt.Println("Percieved elevator stops: ", e.Stops)
			if (e.Stops[i] > 0) {
				direction = def.MD_Up
				fmt.Println("continuing up")
				return direction
			}
		}
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = def.MD_Down
				fmt.Println("turning down")
				e.Dir = direction
				return direction
			}
		}
	} else {
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = def.MD_Down
				fmt.Println("continuing down")
				return direction
			}
		}
		for i := e.CurrentFloor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = def.MD_Up
				fmt.Println("turning up")
				e.Dir = direction
				return direction
			}
		}
	}
	fmt.Println("No pending orders. Stopping")
	e.Dir = def.MD_Stop
	return direction
}

func move_to_next_floor(elevator *def.Elevator) {
	motor_direction := Find_next_stop(elevator)
	driver.SetMotorDirection(motor_direction)
}


func Save_state(state *def.Elevator) {
	fmt.Println("Saving state.")

	jsonState, err := json.Marshal(state)
	def.Check(err)

	err = ioutil.WriteFile("elevator/state.json", jsonState, 0644) // ERROR PRONE
	def.Check(err)
}

func Load_state(state *def.Elevator) {
	fmt.Println("\nLoading data...\n")
	
	jsonState, err := ioutil.ReadFile("state/state.json")
	def.Check(err)
	
	err = json.Unmarshal(jsonState, &state)
	def.Check(err)
}

func LoadState_test(state *def.Elevator) {
	var jsonBlob = []byte(`{"dir":1,"currentfloor":0,"stops":[1,1,1,1]}`)
	
	err := json.Unmarshal(jsonBlob, &state)
	def.Check(err)
}

func Order_complete(e *def.Elevator) {
	e.Stops[e.CurrentFloor] = 0
	//fmt.Println(e.Stops)
}

func Order_accept(e *def.Elevator, o def.Order) {
	e.Stops[o.Floor] = 1
	//fmt.Println(e.Stops)
}

func button_event_to_order(be def.ButtonEvent) def.Order {
	var order = def.Order{}
	order.Floor = be.Floor
	switch be.Button {
		case def.BT_Cab:
			order.Dir = 0
			return order
		case def.BT_HallUp:
			order.Dir = def.MD_Up
			return order
		case def.BT_HallDown:
			order.Dir = def.MD_Down
			return order
		default:
			return order
	}
}

func Evaluate(e def.Elevator, o def.Order) int {
	value := 0
	distance := o.Floor - e.CurrentFloor
	if (distance < 0) {
		distance *= -1
	}
	if (e.Dir == def.MD_Up) {
		if (o.Floor > e.CurrentFloor) {
			for i := e.CurrentFloor; i < o.Floor; i++ {
				value += e.Stops[i] * def.STOP_WEIGHT
			}
		} else {
			for i := o.Floor; i < def.FLOORS; i++ {
				value += e.Stops[i] * def.STOP_WEIGHT
			}
		}
		if (e.Dir != o.Dir) {
			if (o.Floor > e.CurrentFloor) {
				value += 2*(def.FLOORS - o.Floor)
			} else {
				value += 2*(def.FLOORS - e.CurrentFloor)
			}
			value += distance
		} else {
			value += distance
		}
	} else {
		if (o.Floor < e.CurrentFloor) {
			for i := o.Floor; i < e.CurrentFloor; i++ {
				value += e.Stops[i] * def.STOP_WEIGHT
			}
		} else {
			for i := 0; i < o.Floor; i++ {
				value += e.Stops[i] * def.STOP_WEIGHT
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



