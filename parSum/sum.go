package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

func SumFunc(s []int, c chan int) {
	if len(s) == 1 {
		c <- s[0]
		wg.Done()
	} else {
		go SumFunc(s[:len(s)/2], c)
		go SumFunc(s[len(s)/2:], c)
	}
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int, len(s))

	wg.Add(len(s))
	go SumFunc(s, c)

	wg.Wait()
	close(c)
	var sum int
	for element := range c {
		sum += element
	}
	fmt.Println(sum)
	fmt.Scanln()
}
