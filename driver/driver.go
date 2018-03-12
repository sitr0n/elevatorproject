package driver

import ("fmt"
	"sync"
	"net"
	"time"
)

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn
//var d elevio.MotorDirection = elevio.MD_Up

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Button_manager(b <- chan ButtonEvent, e *order.Elevator) {
	for {
		select {
		case event := <- b:
			if (event.Button == BT_Cab) {
				e.Stops[event.Floor]
			} else {
				//TODO: Broadcast corresponding order
				//TODO: Evaluate all elevators and decide which one taking the order
				//TODO: Update corresponding elevator struct -> Stops[event.Floor]
			}
		}
	}
}

func Event_manager(f <- chan int, e *Elevator) {
	prev_floor := -1
	for {
		select {
		case floor := <- f:
			if (floor != prev_floor) {
				SetFloorIndicator(floor)
				if (e.Stops[floor] > 0) {
					SetMotorDirection(MD_Stop)
					SetDoorOpenLamp(true)
					time.Sleep(5*time.Second)
					SetDoorOpenLamp(false)
				}
				motor_direction := Find_next_stop(e)
				SetMotorDirection(motor_direction)
			}
			prev_floor = floor
		}
	}
}

func Find_next_stop(e Elevator) MotorDirection {
	direction := MD_Stop
	if (e.Dir) {
		for i := e.Floor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = MD_Up
				return direction
			}
		}
		for i := e.Floor; i > 0; i-- {
			if (e.Stops[i] > 0) {
				direction = MD_Down
				return direction
			}
		}
	} else {
		for i := e.Floor; i > 0; i-- {
			if (e.Stops[i] > 0) {
				direction = MD_Down
				return direction
			}
		}
		for i := e.Floor; i < 4; i++ {
			if (e.Stops[i] > 0) {
				direction = MD_Up
				return direction
			}
		}
	}
	return direction
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}


func SetMotorDirection(dir MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}



func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func getButton(button ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
