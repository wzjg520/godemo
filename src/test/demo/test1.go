package main

import (
	//"log"
	//"os"
	"fmt"
)

type struct1 struct {
	name string
}

func (s1 *struct1) sayHello() {
	fmt.Print("hello,world")
}

func (s1 *struct1) setName() {
	s1.name = "struct"
}

func (s1 *struct1) getName() {
	fmt.Println(s1.name)
}

type struct2 struct {
	name string
	like string
	d    struct1
}

func main() {
	//file, err := os.Create("data.log")
	//if err != nil {
	//	panic(err)
	//}
	//defer file.Close()
	//
	//log.SetOutput(file)
	//log.Println("hello, world")

	c := &struct2{}
	fmt.Println(c)
	c.d.sayHello()
	c.d.setName()
	c.d.getName()
	fmt.Println(c)

}
