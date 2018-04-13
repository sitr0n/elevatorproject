package elevator

import ("fmt"
	"encoding/json"
	"io/ioutil"
	"time"
	"math/rand"
)

import driver "../driver"
import network "../network"
import def "../def"



func Init() {
	var elevator = def.Elevator{}
	Load_state(&elevator)
	driver.Init("localhost:15657", def.FLOORS)
	
	var remote [def.ELEVATORS]network.Remote
	ch_ack 		:= make(chan bool)
	ch_order 	:= make(chan def.Order)
	ch_obstr   	:= make(chan bool)
	ch_stop    	:= make(chan bool)
	
	network.Init(10, 11, &remote, ch_order, ch_ack)
	
	
	
	ch_buttons := make(chan def.ButtonEvent)
	go Button_manager(ch_buttons, &elevator, &remote)
	//go driver.PollButtons(ch_buttons)
	
	ch_floors  := make(chan int)
	go Event_manager(ch_floors, &elevator)
	go driver.PollFloorSensor(ch_floors)
	
	ch_add_order := make(chan def.Order)
	ch_remove_order := make(chan def.Order)
	go order_queue(ch_add_order, ch_remove_order, ch_buttons)
	
		
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)


	time.Sleep(1*time.Second)

	remote[0].Send <- "wopwop"
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

func Button_manager(b <- chan def.ButtonEvent, e *def.Elevator, remote *[def.ELEVATORS]network.Remote) {
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
				for i := 0; i < def.ELEVATORS; i++ {
					remote[i].Send <- order
				}
				
				//local_cost := Evaluate(*e, order)
				
				//cost1 := Evaluate(remote[0].State, order)
				//cost2 := Evaluate(remote[1].State, order)
				
			
				
			        //network.broadcast_state()
			        //poll states
			        //time.Sleep(100*Millisecond)
				/*
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
	
	jsonState, err := ioutil.ReadFile("elevator/state.json")
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
	order.ID = random_generator(10000)
	order.Stamp = time.Now()
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

func order_queue(ch_add_order <-chan def.Order, ch_remove_order chan def.Order, ch_buttons chan<- def.ButtonEvent) {

	var q []def.Order
	
	for {
		timecheck_order_queue(q, ch_buttons, ch_remove_order)
		select {
		case newQ := <- ch_add_order:
			q = append(q, newQ)
			fmt.Println("added",newQ.ID)
			
		case removeQ := <- ch_remove_order:
			i := 0
			for _,c := range q {
				if c.ID == removeQ.ID {
					fmt.Println("removing ID:", c.ID)
					q = q[:i+copy(q[i:], q[i+1:])]
					fmt.Println(q)
				}
				i++
			}			
		}
	}
}

func timecheck_order_queue(q []def.Order, ch_buttons chan<- def.ButtonEvent, ch_remove_order chan<- def.Order) {
	for _, c := range q {
		if time.Now().Sub(c.Stamp) > 30*time.Second {
			fmt.Println(c.ID," failed")
			//TODO: remove order from slice
			//TODO: create local order - create new buttoneven which is Cab-order
			//TODO: send to channnel which goes to button_manager
			newEvent :=  def.ButtonEvent{Floor: c.Floor, Button: def.BT_Cab}
			ch_buttons <- newEvent
			ch_remove_order <- c
		}
	}
}
func random_generator(size int) int {
	nanotime := rand.NewSource(time.Now().UnixNano())
	convert := rand.New(nanotime)
	random := convert.Intn(size) 
	return random
}
