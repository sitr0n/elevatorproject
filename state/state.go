package state

import ("fmt"
	"encoding/json"
	"io/ioutil"
)


type MotorDirection int
const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type Elevator struct {
	Dir MotorDirection
	CurrentFloor int
	Stops [4]int
}


func Save(s <- chan bool, state *Elevator) {
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

func Load(state *Elevator) {
	fmt.Println("\nLoading data...\n")
	
	jsonState, err := ioutil.ReadFile("state/state.json")
	Check(err)
	
	err = json.Unmarshal(jsonState, &state)
	Check(err)
}

func LoadState_test(state *Elevator) {
	var jsonBlob = []byte(`{"dir":1,"currentfloor":0,"stops":[1,1,1,1]}`)
	
	err := json.Unmarshal(jsonBlob, &state)
	Check(err)
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Check_conn(e error) {
	if e != nil {
		fmt.Println("Connection error")
	}
}
