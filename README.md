# Purpose

I found at least 4 core CPUs are required when compiling petalinux. The trail of GCE is over but fortunately AWS free $100 is available while its computing resources are extremely expensive. So I used a very ugly method to solve this problem:). When I need to complie petalinux, I stop the instance, upgrade the instance type up to t2.xlarge, restart the instance and then the public ip is automatically sent to Telegram Bot. When I finish compiling, I stop the instance, downgrade the instance type to t2.micro(free to use) and restart the instance.



# Prerequsites

AWS Instance 1 (Server t2.micro)

AWS Instance 2 (Client  t2.xlarge / t2.micro)

Telegram Bot



# Communication

Register object in server through RPC and expose the corresponding service. Client uses these methods through remote calls. Public IP will be sent to Telegram Bot after upgrading. 

1. RPC

   User `net/rpc `package

2. Telegram Bot

   Reference to [The usage of Telegram Bot](<https://github.com/go-telegram-bot-api/telegram-bot-api/blob/master/README.md>)

   - Send `/newbot` to Telegram BotFather
   - Send bot name ends with `bot`to Telegram BotFather 
   - Get Bot token returned ty BotFather

# Implementation of Server / Client 
Define Params

```go
import "github.com/aws/aws-sdk-go/aws"

var (
	InstanceID       *string = aws.String("")
	BotToken         string  = ""
	UpInstanceType   string  = ""
	DownInstanceType string  = ""
	ChatID           int     = 123456
	Region           *string = aws.String("")
	AWSID            string  = ""
	AWSSecret        string  = ""
)

```
- Server

  1. Set listening port

  2. Register a `Listener` object, `type Listener ec2.EC2`, register exposed services after creating this object.

     ```go
     type Service EC2.ec2
     
     func main() {
     	// listen on port 1234
     	listener, err := net.Listen("tcp", ":1234")
     	if err != nil {
     		log.Fatal("Listen TCP error: ", err)
     	}
     
     	// init server
     	c, err := cre.NewServerTLSFromFile("../conf/server.pem", "../conf/server.key")
     	if err != nil {
     		log.Fatalf("GRPC credentials.NewServerTLSFromFile err : %v", err)
     	}
     
     	server := grpc.NewServer(grpc.Creds(c))
     	pb.RegisterServerClientServer(server, NewEC2Sess())
     	reflection.Register(server)
     
     	if err := server.Serve(listener); err != nil {
     		log.Fatalf("Failed to Serve %v", err)
     	}
     }
     ```

  3. Implement the required functionality

- Client

  Needing functions: start instance, stop instance, upgrade and downgrade instance.

  ```go
  func main() {
  	c, err := credentials.NewClientTLSFromFile("../conf/server.pem", "Suika")
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
  ```
