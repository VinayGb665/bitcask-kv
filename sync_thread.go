package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// import (

// )

func makeSum(inputs []int, returnChannel chan int) {
	sum := 0
	for _, v := range inputs {
		sum += v
	}
	sum += rand.Int()
	// fmt.Println("Sum in here was", sum)
	returnChannel <- sum
	return
}

func main() {
	// resultChannel := make(chan int)
	// inputs := []int{1, 2, 3, 4, 5}
	// go makeSum(inputs, resultChannel)
	// go makeSum(inputs, resultChannel)
	// go makeSum(inputs, resultChannel)
	// res1, res2 := <-resultChannel, <-resultChannel
	// fmt.Println("Result1 was ", res1)
	// fmt.Println("Result1 was ", res2)
	// var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	// var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		// wg.Add(1)
		go func() {
			mu.Lock()
			defer mu.Unlock()
			counter = counter + 1
			if counter == 1001 {
				cond.Broadcast()
			}
			// wg.Done()
		}()
	}
	mu.Lock()
	cond.Wait()
	fmt.Println("COunt", counter)
	mu.Unlock()

	// go periodic()
	// time.Sleep(5 * time.Second)
}

func periodic() {
	for {

		fmt.Println("Ticking")
		time.Sleep(1 * time.Second)
	}
}
