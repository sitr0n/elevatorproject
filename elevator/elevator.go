package elevator

import ("fmt"
	"encoding/json"
	"io/ioutil"
	"time"
	//"math/rand"
	"os"
	"os/exec"
)

import driver "../driver"
import network "../network"
import def "../def"

var _DOOR_IS_OPEN = false

func Init(remote_address []string) {
	check_remote_address(remote_address)
	start_elevator_server()

	var elevator = def.Elevator{}
	Load_state(&elevator)
	driver.Init("localhost:15657", def.FLOORS)
	
	var remote [def.ELEVATORS]network.Remote
	network.Init(remote_address, &remote)
	
	ch_obstr   	:= make(chan bool, 100)

	ch_stop    	:= make(chan bool, 100)

	ch_buttons := make(chan def.ButtonEvent, 100)
	ch_floors  := make(chan int, 100)
	ch_add_order := make(chan def.Order, 100)
	ch_remove_order := make(chan def.Order, 100)

	go Button_manager(ch_buttons, &elevator, &remote, /*ch_stop,*/ ch_add_order, ch_remove_order)
	go driver.PollButtons(ch_buttons)
	go driver.PollStopButton(ch_stop)
	go Event_manager(ch_floors, &elevator, &remote)
	go driver.PollFloorSensor(ch_floors)
	go order_queue(ch_add_order, ch_remove_order, ch_buttons, &remote)	
	go driver.PollObstructionSwitch(ch_obstr)
	//go driver.PollStopButton(ch_stop)
	go order_handler(&remote, ch_add_order, ch_remove_order, &elevator)
}

func start_elevator_server() {
	ElevatorServer := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer;")
	err := ElevatorServer.Start()
	def.Check(err)
	time.Sleep(500*time.Millisecond)
}

func check_remote_address(arg []string) {
	array_length := len(arg)
	if (array_length != def.ELEVATORS) {
		fmt.Println("Expecting", def.ELEVATORS, "arguments.")
		fmt.Println("Enter remote elevator IP address(es) or workstation number(s).")
		os.Exit(0)
	}

}



func Button_manager(b <- chan def.ButtonEvent, e *def.Elevator, remote *[def.ELEVATORS]network.Remote, /*s <-chan bool,*/ add_order chan<- def.Order, remove_order chan def.Order) {

	for {
		select {
		case event := <- b:
			order := button_event_to_order(event)
			if (event.Button == def.BT_Cab) {
				Order_accept(e, order)
				//add_order <- order
				//Order_undergoing(e, order, remove_order, remote)
				fmt.Println("Cab-Call - Order floor: ", event.Floor)
			} else { 
				network.Broadcast_order(order, remote)
				add_order <- order 
				taker := delegate_order(order, *e, *remote)
				if(taker == -1) {
					Order_accept(e, order)
					
					Order_undergoing(e, order, remove_order, remote) //ordre er bestemt til å taes av DENNE pcen, så goroutinen for completion startes her
					network.Send_ack(*remote)
				} else {
					order_taken := remote[taker].Await_ack()
					if (order_taken == false) {
						Order_accept(e, order)
						Order_undergoing(e, order, remove_order, remote)
					}
				}
			}
			Save_state(e)

			network.Send_state_to_all(*e, remote)
		
		/*
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
		*/
			
		}
	}
}

func Event_manager(f <- chan int, e *def.Elevator, remote *[def.ELEVATORS]network.Remote) {
	prev_floor := -1
	for {
		select {
		case floor := <- f:
			if (floor != prev_floor) {
				driver.SetFloorIndicator(floor)
				e.CurrentFloor = floor
				fmt.Println("Current floor: ", e.CurrentFloor)
				if (e.Stops[floor] > 0) {
					e.Stops[floor] = 0
					fmt.Println(e.Stops)
					open_door()
					
				}
				move_to_next_floor(e)
			}
			Save_state(e)
			prev_floor = floor

		case <- remote[0].Reconnected:
			go remote[0].Send_state(*e)
		
		case <- remote[1].Reconnected:
			go remote[1].Send_state(*e)
		
		}
	}
}

func open_door() {
	//fmt.Println("Stopping at floor ", floor)
	_DOOR_IS_OPEN = true
	driver.SetMotorDirection(def.MD_Stop)
	driver.SetDoorOpenLamp(true)
	time.Sleep(5*time.Second)
	driver.SetDoorOpenLamp(false)
	_DOOR_IS_OPEN = false
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
	jsonState, err := json.Marshal(state)
	def.Check(err)

	err = ioutil.WriteFile("elevator/state.json", jsonState, 0644) // ERROR PRONE
	def.Check(err)
}

func Load_state(state *def.Elevator) {	
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



