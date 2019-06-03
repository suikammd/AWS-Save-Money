package main

import (
	"github.com/prometheus/common/log"
	"net/rpc"
	"flag"
	"fmt"
)

var InstanceIP = ""

func main() {
	client, err := rpc.Dial("tcp", InstanceIP + ":42586")
	if err != nil {
		log.Fatal(err)
	}

	var in string
	flag.StringVar(&in, "s", in, "name")
	flag.Parse()
	var line []byte = []byte(in)

	method := string(line)
	fmt.Printf("Current Method is %s\n", method)
	var reply bool

	switch method {
	case "b":
		fmt.Println("Restart and Modify to Build")
		err = client.Call("Listener.ModifyUpInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "f":
		fmt.Println("Finish Building")
		err = client.Call("Listener.ModifyDownInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "start":
		fmt.Println("Start Instance")
		err = client.Call("Listener.StartInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "stop":
		fmt.Println("Stop Instace")
		err = client.Call("Listener.StopInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "des":
		fmt.Println("Describe Instace")
		err = client.Call("Listener.DescribeInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
		//case "bill":
		//	err = client.Call("Listener.GetBilling", line, &reply)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
	default:
		log.Fatal("No Such Method!!!")
	}
}