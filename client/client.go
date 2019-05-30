package main

import (
	"bufio"
	"os"
	"github.com/prometheus/common/log"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "3.112.232.91:42586")
	if err != nil {
		log.Fatal(err)
	}

	in := bufio.NewReader(os.Stdin)
	line, _, err := in.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	method := string(line)
	var reply bool

	switch method {
	case "b":
		err = client.Call("Listener.ModifyUpInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "f":
		err = client.Call("Listener.ModifyDownInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "start":
		err = client.Call("Listener.StartInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "stop":
		err = client.Call("Listener.StopInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	case "des":
		err = client.Call("Listener.DescribeInstance", line, &reply)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("No Such Method!!!")
	}
}