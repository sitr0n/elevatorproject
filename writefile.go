package writefile

import (
	"fmt"
	"io/ioutil"
)
import order "../order"


/*	
type person struct {
    name string
    age  int
}
*/
func check(e error) {
	if e != nil {
		panic(e)
	}
}
func SaveState(s <- chan bool, elevator order.Elevator) {
	for {
		select {
		case <- s :
			var state_dir string	= string(elevator.Dir)
			var state_floor string	= string(elevator.CurrentFloor)
			var state_stops string	= string(elevator.Stops)
			
			fmt.Println("Direction: ", state_dir)
			fmt.Println("Direction: ", state_floor)
			fmt.Println("Direction: ", state_stops)
			//d1 := person{name: "bob", age: 50}
			//d1 := []byte("hello\ngo\n")
			//err := ioutil.WriteFile("./ElevatorState", d1, 0644)
			//check(err)
	}	}
}
