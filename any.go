package main

import (
	"fmt"
	"time"
)

func main() {
	t1 := time.Now()
	a := "12312312312312312312312312312312312312"

	if a == "" {
		fmt.Println("OK google")
	}
	fmt.Println(time.Since(t1))

	t2 := time.Now()
	if len(a) == 0{
		fmt.Println("ok google")
	}

	fmt.Println(time.Since(t2))
	
}