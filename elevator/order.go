package elevator

import ("fmt"
	//"encoding/json"
	//"io/ioutil"
	"time"
	"math/rand"
)

import def "../def"
import network "../network"

func button_event_to_order(be def.ButtonEvent) def.Order {
	var order = def.Order{}
	order.Floor = be.Floor
	order.ID = random_generator(10000)
	order.Stamp = time.Now()
	order.AddOrRemove = def.ADD
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

func delegate_order(order def.Order, elevator def.Elevator, remote [def.ELEVATORS]network.Remote) int {
	var taker int = 0

	local_cost := Evaluate(elevator, order)
	var cost [def.ELEVATORS]int = [def.ELEVATORS]int{}

	var FIRST_SCAN bool = true
	for i := 0; i < def.ELEVATORS; i++ {
		cost[i] = Evaluate(remote[i].Get_state(), order)
		if (remote[i].Alive == false) {
			cost[i] = 999
		}
		if (FIRST_SCAN) {
			taker = i
			FIRST_SCAN = false
		}
		if (cost[i] < cost[taker]) {
			taker = i
		}
	}

	if (local_cost < cost[taker]) {
		return -1
	}
	return taker
}

func Wait_for_completion(e *def.Elevator, order def.Order, remove_order chan<- def.Order, r *[def.ELEVATORS]network.Remote) {
	for {
		if (e.CurrentFloor == order.Floor) {
			order.AddOrRemove = def.REMOVE
			network.Broadcast_order(order, r)
			remove_order <- order
			break
		}
	}
}

func Order_undergoing(e *def.Elevator, order def.Order, remove_order chan<- def.Order, r *[def.ELEVATORS]network.Remote) {
	go Wait_for_completion(e ,order, remove_order, r)
	
}

func Order_accept(e *def.Elevator, o def.Order) {
	e.Stops[o.Floor] = 1
	//fmt.Println(e.Stops)
}

func order_queue(ch_add_order <-chan def.Order, ch_remove_order chan def.Order, ch_buttons chan<- def.ButtonEvent, r *[def.ELEVATORS]network.Remote) {

	var q []def.Order
	
	for {
		
		select {
		case newO := <- ch_add_order:
			q = append(q, newO)
			fmt.Println("added order ID:",newO.ID)
		/*
		case remoteO := <- r[0].Orderchan:
			if (remoteO.AddOrRemove == def.ADD) {
				q = append(q, remoteO)
			} else {
				ch_remove_order <- remoteO
			}	
		*/			
		case removeO := <- ch_remove_order:
			i := 0
			for _,c := range q {
				if c.ID == removeO.ID {
					fmt.Println("removing order ID:", c.ID)
					q = q[:i+copy(q[i:], q[i+1:])]
					//fmt.Println(q)
				}
				i++
			}
		default:
			timecheck_order_queue(q, ch_buttons, ch_remove_order)			
		}
	}
}

func order_handler(r *[def.ELEVATORS]network.Remote, ch_add_order chan<- def.Order, ch_remove_order chan<- def.Order, e *def.Elevator) { //listener
	for {
		select {
		case order := <- r[0].Orderchan:
			if order.AddOrRemove == def.REMOVE {
				ch_remove_order <- order
			} else {
				taker := delegate_order(order, *e, *r)
				ch_add_order <- order
				if(taker == -1) {
					Order_accept(e, order) 
					Order_undergoing(e, order, ch_remove_order, r) //ordre er bestemt til å taes av DENNE pcen, så goroutinen for completion startes her
					network.Send_ack(*r)
				} else {
					order_taken := r[taker].Await_ack()
					if (order_taken == false) {
						Order_accept(e, order)
					}
				}
			}
		case order := <- r[1].Orderchan:
			if order.AddOrRemove == def.REMOVE {
				ch_remove_order <- order
			} else {
				taker := delegate_order(order, *e, *r)
				ch_add_order <- order
				if(taker == -1) {
					Order_accept(e, order) 
					Order_undergoing(e, order, ch_remove_order, r) //ordre er bestemt til å taes av DENNE pcen, så goroutinen for completion startes her
					network.Send_ack(*r)
				} else {
					order_taken := r[taker].Await_ack()
					if (order_taken == false) {
						Order_accept(e, order)
					}
				}
			}
		}
	}	
}

func timecheck_order_queue(q []def.Order, ch_buttons chan<- def.ButtonEvent, ch_remove_order chan<- def.Order) {
	for _, c := range q {
		//time.Sleep(time.Second)
		if time.Now().Sub(c.Stamp) > 5*time.Second {
			fmt.Println(c.ID," failed")
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
