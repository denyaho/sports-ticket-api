package main

import "sync"

func getCounter() int {
	var counter int
	var wg sync.WaitGroup
	var lock sync.Mutex
	
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			lock.Lock()
			defer lock.Unlock()
			for i := 0; i < 1000; i++ {
				counter++
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return counter
}

func main() {
	counter := getCounter()
	println("Final counter value:", counter)	
}