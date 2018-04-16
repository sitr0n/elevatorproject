package main
import ("fmt"
	"os"
	"./elevator"
)

func main() {
	ch_exit	   := make(chan bool)
	args := os.Args[1:]
	elevator.Init(args)
	
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	fmt.Println("    STARTING ELEVATOR     ")
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	
	<- ch_exit
}
