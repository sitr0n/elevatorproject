package elevator
import driver "../driver"
import network "../network"
import def "../def"
import ("fmt"
	"encoding/json"
	"io/ioutil"
	"time"
	"os"
	"os/exec"
)

func Init(remote_address []string) {
	check_remote_address(remote_address)
	start_elevator_server()
	driver.Init("localhost:15657", def.FLOORS)
	
	var remote [def.ELEVATORS]network.Remote
	network.Init(remote_address, &remote)

	
	var elevator = def.Elevator{}
	load_state(&elevator)
	elevator.Dir = def.MD_Stop

	
	//var remote [def.ELEVATORS]network.Remote
	//network.Init(remote_address, &remote)
	
	ch_obstr   	:= make(chan bool, 100)

	ch_stop    	:= make(chan bool, 100)

	ch_buttons := make(chan def.ButtonEvent, 100)
	ch_floors  := make(chan int, 100)
	ch_add_order := make(chan def.Order, 100)
	ch_remove_order := make(chan def.Order, 100)
	ch_turn_on_light := make(chan def.Order, 100)
	ch_turn_off_light := make(chan def.Order, 100)
	
	

	go Button_manager(ch_buttons, &elevator, &remote, ch_stop, ch_add_order, ch_remove_order, ch_turn_on_light)
	go driver.PollButtons(ch_buttons)
	go driver.PollStopButton(ch_stop)
	go Event_manager(ch_floors, &elevator, &remote)
	go driver.PollFloorSensor(ch_floors)
	go order_queue(ch_add_order, ch_remove_order, ch_buttons, &remote, ch_turn_off_light)	
	go driver.PollObstructionSwitch(ch_obstr)
	go order_handler(&remote, ch_add_order, ch_remove_order, &elevator, ch_turn_on_light)
	
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
		os.Exit(2)
	}
}


func Button_manager(button <- chan def.ButtonEvent, e *def.Elevator, remote *[def.ELEVATORS]network.Remote, stop <-chan bool, add_order chan<- def.Order, remove_order chan def.Order, turn_on_light chan<- def.Order) {

	for {
		select {
		case event := <- button:
			order := button_event_to_order(event)
			if (event.Button == def.BT_Cab) {
				Order_accept(e, order)
				fmt.Println("Cab-Call - Order floor: ", event.Floor)
			} else { 
				network.Broadcast_order(order, remote)
				add_order <- order 
				turn_on_light <- order
				taker := delegate_order(order, *e, *remote)
				if(taker == -1) {
					Order_accept(e, order)
					Order_undergoing(e, order, remove_order, remote) //ordre er bestemt til å taes av DENNE pcen, så goroutinen for completion startes her
					network.Send_ack(def.Ack_order_accept, remote)
				} else {
					order_taken := remote[taker].Await_ack(def.Ack_order_accept)
					if (order_taken == false) {
						fmt.Println("BM: ack failed")
						Order_accept(e, order)
						Order_undergoing(e, order, remove_order, remote)
					}
				}
			}
			save_state(e)

			network.Send_state_to_all(*e, remote)
		

		case <- stop:
			emergency_stop(e)
		}
	}
}

func emergency_stop(e *def.Elevator) {
	if (e.EMERG_STOP == false) {
		e.EMERG_STOP = true
		driver.SetMotorDirection(def.MD_Stop)
		e.Dir = def.MD_Stop
		driver.SetStopLamp(true)
	} else {
		move_to_next_floor(e)
		e.EMERG_STOP = false
		driver.SetStopLamp(false)
	}
	
}

func lights_manager(turn_on_light <-chan def.Order, turn_off_light <-chan def.Order) {
	//on order all PCs floor light should turn on
	//on completion of order, floor light should turn off.
	for {
		select {
		case order := <-turn_on_light:
			if (order.Dir == def.MD_Up) {
				driver.SetButtonLamp(def.UP, order.Floor, true)
			}
			if (order.Dir == def.MD_Down){
				driver.SetButtonLamp(def.DOWN, order.Floor, true)
			}
		case order := <-turn_off_light:
			if (order.Dir == def.MD_Up) {
				driver.SetButtonLamp(def.UP, order.Floor, false)
			}
			if (order.Dir == def.MD_Down){
				driver.SetButtonLamp(def.DOWN, order.Floor, false)
			}
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
				//fmt.Println("Current floor: ", e.CurrentFloor)
				if (e.Stops[floor] > 0) {
					e.Stops[floor] = 0
					fmt.Println(e.Stops)
					open_door(e)
					
				}
				if (e.EMERG_STOP == false) {
					move_to_next_floor(e)
				}
			}
			save_state(e)
			prev_floor = floor

		case <- remote[0].Reconnected:
			go remote[0].Send_state(*e)
		
		case <- remote[1].Reconnected:
			go remote[1].Send_state(*e)
		
		}
	}
}

func open_door(e *def.Elevator) {
	e.DOOR_OPEN = true
	driver.SetMotorDirection(def.MD_Stop)
	driver.SetDoorOpenLamp(true)
	time.Sleep(5*time.Second)
	driver.SetDoorOpenLamp(false)
	e.DOOR_OPEN = false
}

func Find_next_stop(e *def.Elevator) def.MotorDirection {
	var direction def.MotorDirection = def.MD_Stop
	if (e.Dir == def.MD_Up) {
		for i := e.CurrentFloor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = def.MD_Up
				//fmt.Println("Continuing up")
				return direction
			}
		}
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = def.MD_Down
				//fmt.Println("Turning down")
				e.Dir = direction
				return direction
			}
		}
	} else {
		for i := e.CurrentFloor; i >= 0; i-- {
			if (e.Stops[i] > 0) {
				direction = def.MD_Down
				//fmt.Println("Continuing down")
				return direction
			}
		}
		for i := e.CurrentFloor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = def.MD_Up
				//fmt.Println("Turning up")
				e.Dir = direction
				return direction
			}
		}
	}
	fmt.Println("No pending orders. Stopping.")
	e.Dir = def.MD_Stop
	return direction
}

func move_to_next_floor(elevator *def.Elevator) {
	motor_direction := Find_next_stop(elevator)
	driver.SetMotorDirection(motor_direction)
}

func save_state(state *def.Elevator) {
	jsonState, err := json.Marshal(state)
	def.Check(err)

	err = ioutil.WriteFile("elevator/state.json", jsonState, 0644)
	def.Check(err)
}

func load_state(state *def.Elevator) {	
	jsonState, err := ioutil.ReadFile("elevator/state.json")
	def.Check(err)
	
	err = json.Unmarshal(jsonState, &state)
	def.Check(err)

	driver.SetStopLamp(state.EMERG_STOP)
	state.DOOR_OPEN = false
	driver.SetDoorOpenLamp(false)
	state.Dir = def.MD_Stop
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



