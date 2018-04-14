package elevator

import ("fmt"
	"encoding/json"
	"io/ioutil"
	"time"
	//"math/rand"
)

import driver "../driver"
import network "../network"
import def "../def"



func Init() {
	var elevator = def.Elevator{}
	Load_state(&elevator)
	driver.Init("localhost:15657", def.FLOORS)
	
	var remote [def.ELEVATORS]network.Remote
	network.Init(10, 11, &remote)
	
	//network.Init(10, 11, &remote)
	//<- remote[0].Orderchan
	ch_obstr   	:= make(chan bool)
	ch_stop    	:= make(chan bool, 100)
	ch_buttons := make(chan def.ButtonEvent)
	ch_floors  := make(chan int)
	ch_add_order := make(chan def.Order)
	ch_remove_order := make(chan def.Order)

	go Button_manager(ch_buttons, &elevator, &remote, ch_stop, ch_add_order, ch_remove_order)
	go driver.PollButtons(ch_buttons)
	go driver.PollStopButton(ch_stop)
	go Event_manager(ch_floors, &elevator)
	go driver.PollFloorSensor(ch_floors)
	go order_queue(ch_add_order, ch_remove_order, ch_buttons)	
	go driver.PollObstructionSwitch(ch_obstr)
	go driver.PollStopButton(ch_stop)


	time.Sleep(1*time.Second)

	fmt.Println("waiting...")
	//<- remote.Orderchan
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



func Button_manager(b <- chan def.ButtonEvent, e *def.Elevator, remote *[def.ELEVATORS]network.Remote, s <-chan bool, add_order chan<- def.Order, remove_order chan def.Order) {
	for {
		select {
		case event := <- b:
			order := button_event_to_order(event)
			if (event.Button == def.BT_Cab) {
				Order_state(e, order)
				add_order <- order
				Order_accept(e, order, remove_order)
				fmt.Println("-------------------------------")
				fmt.Println("Button - Order floor: ", event.Floor)
				fmt.Println("Button - Elevator stops: ", e.Stops)
				fmt.Println("Button - Current motor direction: ", e.Dir)
				if (e.Dir == def.MD_Stop) {
					move_to_next_floor(e)
				}
			} else { 
				for i := 0; i < def.ELEVATORS; i++ {
					remote[i].Send_order(order)
				}
				
				decision := decide_to_take_order(order, *e, *remote)
				if(decision == true) {
					Order_state(e, order)
					add_order <- order 
					Order_accept(e, order, remove_order) //ordre er bestemt til å taes av DENNE pcen, så goroutinen for completion startes her
					remote[0].Send_ack()
					remote[1].Send_ack()
				}
				
				timeout := make(chan bool)
				timer_cancel := make(chan bool)
				go timeout_timer(timer_cancel, timeout)
				if (decision == false) {
					select {
					case <- remote[0].Ackchan:
						timer_cancel <- true
					
					case <- timeout:
						Order_state(e, order)
					}
				}
			}
			Save_state(e)
			remote[0].Send_state()
			remote[1].Send_state()
		
		case stop := <- s:
			//var prevDir def.MotorDirection
			time.Sleep(time.Second)
			if stop == true {
				//prevDir = e.Dir
				//e.Dir = def.MD_Stop
				fmt.Println("stopping is true")
				//time.Sleep(20 * time.Millisecond)
				
			} else {
				//e.Dir = prevDir
				fmt.Println("stopping is false")
				//time.Sleep(20 * time.Millisecond)
			}
			
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



