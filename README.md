---

# 目的

在编译Petalinux的时候，需要至少4核CPU。GCE的虚拟机用完了，AWS计算资源又非常贵，不过还好有免费的100$可以用。因此我决定采取非常丑的办法（还不是因为菜），在准备编译时关闭AWS Instance，从t2.micro（AWS免费）升级到t2.xlarge，然后启动机器将Public IP发送到Telegram Bot；在编译完成时关闭并降级最后再重新启动Instance。

# 环境

AWS Instance 1 (Server t2.micro)

AWS Instance 2 (Client  t2.xlarge / t2.micro)

Telegram Bot

# 通信

一开始没有考虑ssh连接到实例需要公网IP的问题，在实例升级完成重启之后就结束了。考虑了IP的问题之后，决定将IP发送到Telegram Bot上。最后才去的做法是，通过RPC在服务端注册对象，通过对象的类型名暴露这个服务，客户端通过远程调用使用这些方法。公网IP在完成升级重启后发送到Telegram Bot。当然可以不用RPC，把所有请求都丢给Telegram Bot，好看一些。

1. RPC

   使用Go的`net/rpc`库

2. Telegram Bot

   参考[Telegram Bot使用](<https://github.com/go-telegram-bot-api/telegram-bot-api/blob/master/README.md>)

   - 向Telegram BotFather发送`/newbot`
   - 输入以`bot`结尾的Bot名字
   - 获得BotFather返回的Bot Token

# Server / Client 代码实现
在使用该代码前请手动定义这些参数

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

  1. 设置服务端监听端口

  2. 在服务端注册一个`Listener对象，`type Listener ec2.EC2`, 创建这个对象之后注册其对应的服务，这些服务必须是可以暴露的

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

  3. 实现需要的功能

     根据需求实习需要暴露的功能

- Client

  需要的功能有启动、关闭、升级和降级实例，用`switch`语句选择即可

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
