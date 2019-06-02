---

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
     type Listener ec2.EC2
     
     // 初始化Listener对象
     func New() (l Listener){
     	// create EC2 service client
     	sess, _ := session.NewSession(&aws.Config{
     		Region:      Region,
     		Credentials: credentials.NewStaticCredentials(AWSID, AWSSecret, ""),
     	})
     
     	l = (Listener)(*ec2.New(sess))
     	return
     }
     
     // 对外暴露的服务等
     func (l Listener) StartInstance(line []byte, ack *bool) error {}
     
     ...
     
     listener = New()
     // 注册服务
     rpc.Register(listener)
     ```

  3. Implement the required functionality

- Client

  Needing functions: start instance, stop instance, upgrade and downgrade instance.

  ```go
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
  ```
