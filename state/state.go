package state

import ("fmt"
	"encoding/json"
	"io/ioutil"
)
import order "../order"

func SaveState(s <- chan bool, elevator *order.Elevator) {
	for {
		select {
		case <- s :
			fmt.Println("Saving state.")
			fmt.Println(elevator.Stops)
			
			jsonState, err := json.Marshal(elevator)
			if err != nil {
				fmt.Println("Failed to convert elevator state to Marshal.")
				panic(err)
			}

			err = ioutil.WriteFile("state/state.json", jsonState, 0644)
			if err != nil {
				fmt.Println("Failed to save state.")
				panic(err)
			}

		}
	}
}

func LoadState(elevator *order.Elevator) {
	//var jsonBlob = []byte(`{"dir":1,"currentfloor":2,"stops":[1,1,1,1]}`)
	fmt.Println("--------------------------")
	fmt.Println("Loading state:")
	fmt.Println("Previous state:\t", elevator.Stops)
	
	jsonState, errr := ioutil.ReadFile("state/state.json")
	if errr != nil {
		panic(errr)
	}
	
	err := json.Unmarshal(jsonState, &elevator)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("New state:\t", elevator.Stops)
	fmt.Println("--------------------------")
}
