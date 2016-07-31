package main

import (
	//"log"
	//"os"
	"fmt"
)

func main() {
	//file, err := os.Create("data.log")
	//if err != nil {
	//	panic(err)
	//}
	//defer file.Close()
	//
	//log.SetOutput(file)
	//log.Println("hello, world")
	type hello struct {
		name string
		like string
	}

	a := &hello{
		name : "桂花",
		like : "apple",
	}

	fmt.Println(*a)
}
