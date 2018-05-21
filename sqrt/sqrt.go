package main

import (
	"fmt"
	"math"
)

func Sqrt(x float64) (float64, float64) {
	var i int
	guess := float64(1)
	actual := math.Sqrt(x)
	for ; actual != guess; i++ {
		guess += (x - guess*guess) / (2 * guess)
	}
	return float64(i), guess
}

func main() {
	var num float64
	fmt.Scanf("%f", &num)
	fmt.Println(Sqrt(num))
}
