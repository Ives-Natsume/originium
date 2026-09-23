package main

import (
	"fmt"
	"sync"
	// "time"
)

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		inGoroutine := 799
		ch1 <- inGoroutine
		fromMain := <-ch2
		fmt.Println("goroutine:", inGoroutine, fromMain)
	}()
	inMain := 325
	fromGoroutine := <-ch1
	ch2 <- inMain

	fmt.Println("main:", inMain, fromGoroutine)

	wg.Wait()
}