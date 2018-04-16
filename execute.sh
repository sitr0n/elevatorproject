#!/bin/bash
package main

import ("fmt"
	//"os/exec"
	//"time"
	"os"
)
import elevator "./elevator"
//import driver "./driver"
//import def "./def"



func main() {

	//args := os.Args[1:]
	args := "6 7"
	elevator.Init(args)
	
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	fmt.Println("    STARTING ELEVATOR     ")
	fmt.Println("--------------------------")
	fmt.Println("--------------------------")
	
	
	ch_exit	   := make(chan bool)


	<- ch_exit
}
