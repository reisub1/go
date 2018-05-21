package main

import (
	"fmt"
)

func SumFunc(s []int, c chan int) {
	if len(s) == 1 {
		c <- s[0]
	} else {
		SumFunc(s[:len(s)/2], c)
		SumFunc(s[len(s)/2:], c)
	}
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)

	SumFunc(s, c)

	close(c)
	var sum int
	for element := range c {
		fmt.Println(element)
		sum += element
	}
	fmt.Println(sum)
	fmt.Scanln()
}
