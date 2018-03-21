package main

import (
        "./network"
)

func main() {
        go network.Init(14400,14500) 
}
