package main

import (
	"github.com/prometheus/common/log"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	pb "../proto"
	"context"
	"google.golang.org/grpc/credentials"
)

var InstanceIP = ""

func main() {
	c, err := credentials.NewClientTLSFromFile("../conf/server.pem", "SUika")
	if err != nil {
		log.Fatalf("credentials.NewClientTLSFromFile err: err")
	}

	conn, err := grpc.Dial(InstanceIP + ":1234", grpc.WithTransportCredentials(c))
	if err != nil {
		log.Fatalf("grpc.Dail error %v", err)
	}
	defer conn.Close()

	client := pb.NewServerClientClient(conn)
	// flags
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
		_, err = client.ModifyUpInstance(context.Background(), &pb.Request{Line:line, Ack:reply})
		if err != nil {
			log.Fatal(err)
		}
	case "f":
		fmt.Println("Finish Building")
		_, err = client.ModifyDownInstance(context.Background(), &pb.Request{Line:line, Ack:reply})
		if err != nil {
			log.Fatal(err)
		}
	case "start":
		fmt.Println("Start Instance")
		_, err = client.StartInstance(context.Background(), &pb.Request{Line:line, Ack:reply})
		if err != nil {
			log.Fatal(err)
		}
	case "stop":
		fmt.Println("Stop Instace")
		_, err = client.StopInstance(context.Background(), &pb.Request{Line:line, Ack:reply})
		if err != nil {
			log.Fatal(err)
		}
	case "des":
		fmt.Println("Describe Instace")
		_, err = client.DescribeInstance(context.Background(), &pb.Request{Line:line, Ack:reply})
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("No Such Method!!!")
	}
}
