package main

import (
	"fmt"

	"github.com/seamia/libs/where"
)

func main() {
	fmt.Println("-------------")
	// fmt.Println(where.AmI())
	two()
}

func two() {
	three()
}

func three() {
	four()
}

func four() {
	five()
}

func five() {
	fmt.Println(where.AmI())
}
