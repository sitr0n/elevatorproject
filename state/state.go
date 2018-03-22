package state

import ("fmt"
	"encoding/json"
	"io/ioutil"
)
import order "../order"

func Save(s <- chan bool, state *order.Elevator) {
	for {
		select {
		case <- s :
			fmt.Println("Saving state.")
			
			jsonState, err := json.Marshal(state)
			Check(err)

			err = ioutil.WriteFile("state/state.json", jsonState, 0644)
			Check(err)
		}
	}
}

func Load(state *order.Elevator) {
	fmt.Println("--------------------------")
	fmt.Println("Loading state:")
	fmt.Println("Previous state:\t", state.Stops)
	
	jsonState, err := ioutil.ReadFile("state/state.json")
	Check(err)
	
	err = json.Unmarshal(jsonState, &state)
	Check(err)
	
	fmt.Println("New state:\t", state.Stops)
	fmt.Println("--------------------------")
}

func LoadState_test(state *order.Elevator) {
	var jsonBlob = []byte(`{"dir":1,"currentfloor":0,"stops":[1,1,1,1]}`)
	
	err := json.Unmarshal(jsonBlob, &state)
	Check(err)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
