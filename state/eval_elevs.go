package state

import (
    "fmt"
    
)
/*

elevator_local := *elevator
elevator_remote1 := *elevator2
elevator_remote2 := *elevator3



	local_value := driver.Evaluate(elevator_local)
    remote_value1 := driver.Evaluate(elevator_remote1)
    remote_value2 := driver.Evaluate(elevator_remote2)
*/


func Choose_elevator(local int, remote1 int, remote2 int) {
        
        if (state.remote_elev1_alive && state.remote_elev2_alive) {
                if (local_value < remote_value1 && local_value < remote_value2) {
                        //TODO: set order to local elevator
                }
		}
		if (!state.remote_elev1_alive && state.remote_elev2_alive) {
                if (local_value < remote_value2) {
                        //TODO: set order to local elevator
                }
		}
		if (state.remote_elev1_alive && !state.remote_elev2_alive) {
                if (local_value < remote_value1) {
                        //TODO: set order to local elevator
                }
		}
		if (!state.remote_elev1_alive && !state.remote_elev2_alive) {
              //TODO: set order to local elevator
		}
                
     //TODO: Order accept by elevator
}
