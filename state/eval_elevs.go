package state

import (
    "fmt"
    "reflect"
)


elevator_local := *elevator
elevator_remote1 := *elevator2
elevator_remote2 := *elevator3


func Evaluate_all() {

        local_value := driver.Evaluate(elevator_local)
        remote_value1 := driver.Evaluate(elevator_remote1)
        remote_value2 := driver.Evaluate(elevator_remote2)
        
        case remote-elev1-alive && remote-elev2-alive:
                if (local_value > remote_value1 && local_value > remote_value2) {
                        //TODO: set order to local elevator
                }
        case remote-elev1-lost && remote-elev2-alive:
                if (local_value > remote_value2) {
                        //TODO: set order to local elevator
                }
        case remote-elev2-lost && remote-elev1-alive:
                if (local_value > remote_value1) {
                        //TODO: set order to local elevator
                }
        case remote-elev1-lost && remote-elev2-lost:
                //TODO: set order to local elevator
                
     //TODO: Order accept by elevator
}

/*
import (
    "fmt"
    "reflect"
)

func superstruct() {
        
        x := Superstruct{state.Elevator; state.Elevator; state.Elevator}{elevator_local, elevator_remote1, elevator_remote2}

        v := reflect.ValueOf(x)

        values := make([]interface{}, v.NumField())

        for i := 0; i < v.NumField(); i++ {
                values[i] = v.Field(i).Interface()
    }


}


func main() {
    x := struct{Foo string; Bar int }{"foo", 2}

    v := reflect.ValueOf(x)

    values := make([]interface{}, v.NumField())

    for i := 0; i < v.NumField(); i++ {
        values[i] = v.Field(i).Interface()
    }

    fmt.Println(values)
}
*/
