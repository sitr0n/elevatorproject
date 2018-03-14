package main

import (
	"fmt"
	"io/ioutil"
)

	
type person struct {
    name string
    age  int

func check(e error) {
	if e != nil {
		panic(e)
	}
}


func main() {
    fmt.Printf("hello, world\n")
	d1 := person{name: "bob", age: 50}
	//d1 := []byte("hello\ngo\n")
	err := ioutil.WriteFile("./ElevatorState", d1, 0644)
	check(err)
	
}