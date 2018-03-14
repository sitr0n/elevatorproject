package main

import (
		"fmt"
		"io/ioutil"
		"os"
)

type person struct {
    name1 string
    name2  string
}


func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main () {

	dat, err := ioutil.ReadFile("./ElevatorState")
    check(err)
    fmt.Print(string(dat))
	
	f, err := os.Open("./ElevatorState")
    check(err)
	
	o2, err := f.Seek(9, 0)
    check(err)
    b2 := make([]byte, 4)
    n2, err := f.Read(b2)
    check(err)
    fmt.Printf("%d bytes @ %d: %s\n", n2, o2, string(b2))
	s := string(b2)
	b := "bønna"

	navn := person{s,b}
	fmt.Println(navn.name1)
	
}