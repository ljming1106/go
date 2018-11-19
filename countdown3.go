package main

import "fmt"

func main() {
	//fmt.Println("vim-go")
	fmt.Println("Commencing cuntdown. Press return to abort.")
	tick := time.Tick(1 * time.Second)
	for countdown := 10; countdown > 0; countdown-- {
		fmt.Println(countdown)
		select {
		case <-tick:
		//noting to do
		case <-abort:
			fmt.Println("Launch abotd")
			return
		}
		launch()
	}
}
