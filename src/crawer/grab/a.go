package main

import (
	"fmt"
	//"runtime/debug"
	"logging"
)

func main() {
	//defer func() {
	//	if err := recover(); err != nil {
	//		//debug.PrintStack()
	//		fmt.Sprintf("Fatal Error: %s\n", err)
	//		fmt.Println(err)
	//	}
	//}()

	logger := logging.NewSimpleLogger()
	fmt.Println(logger)
	//filename := "crawer.log"
	//file, err := os.Create(filename)
	//if err != nil {
	//	panic(err)
	//}
	//defer file.Close()
	//logging := &log.Logger{}
	//logging.SetOutput(file)
	//logger := &ConsoleLogger{
	//	log: logging,
	//}
	//logger.SetDefaultInvokingNumber()
	//expectedInvokingNumber := uint(1)
	//currentInvokingNumber := logger.getInvokingNumber()
	//if currentInvokingNumber != expectedInvokingNumber {
	//	t.Errorf("The current invoking number %d should be %d!", currentInvokingNumber, expectedInvokingNumber)
	//}

	logger.Info("hello world")
	//testLogger(t, logger)
}
