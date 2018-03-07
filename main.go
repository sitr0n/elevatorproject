package main

import ("fmt"
	"./order"
	)



func main() {
	stp := [order.FLOORS]int{0,1,1,0}
	ordr := order.Order{Dir: false, Floor: 1}
	elv := order.Elevator{Dir: true, CurrentFloor: 0, Stops: stp}
	fmt.Println(elv)
	a := order.Evaluate (elv, ordr)
	fmt.Println(a)
	elv.CurrentFloor = 1
	order.Order_complete(elv)
	order.Order_accept(elv, ordr)
}