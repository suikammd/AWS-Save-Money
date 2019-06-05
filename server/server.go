package main

import (
	"fmt"
	"log"
	"net"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
	"context"
	"google.golang.org/grpc"
	cre "google.golang.org/grpc/credentials"
	pb "../proto"
	"google.golang.org/grpc/reflection"
)

type Service ec2.EC2

func NewTeleBot() (*tgbotapi.BotAPI, error){
	// Init Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
		return &tgbotapi.BotAPI{}, err
	}
	bot.Debug = true
	return bot, nil
}

func NewEC2Sess() (l Service){
	// create EC2 service client
	sess, _ := session.NewSession(&aws.Config{
		Region:      Region,
		Credentials: credentials.NewStaticCredentials(AWSID, AWSSecret, ""),
	})

	l = (Service)(*ec2.New(sess))
	return
}

func (l Service) describeInstance(describeInput *ec2.DescribeInstancesInput, state string) (*ec2.DescribeInstancesOutput, error) {
	lptr := (ec2.EC2)(l)
	for {
		fmt.Println("Enter Stop Instance Describe Output")
		describeOutput, err := lptr.DescribeInstances(describeInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return &ec2.DescribeInstancesOutput{}, err
		}
		fmt.Println(describeOutput)
		if *describeOutput.Reservations[0].Instances[0].State.Name == state {
			fmt.Printf("Instance state is %s!\n", state)
			return describeOutput, nil
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (l Service) stopInstance() (*ec2.StopInstancesOutput, error) {
	fmt.Println("Enter StopInstance")
	lptr := (ec2.EC2)(l)
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
		DryRun: aws.Bool(true),
	}
	result, err := lptr.StopInstances(input)
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == "DryRunOperation" {
		input.DryRun = aws.Bool(false)
		result, err = lptr.StopInstances(input)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Success", result.StoppingInstances)
		}
	} else {
		log.Fatal(err)
	}
	return result, nil
}

func (l Service) StopInstance(ctx context.Context, request *pb.Request) (*pb.Response, error){
	_, err := l.stopInstance()
	if (err != nil) {
		fmt.Println(err)
	}
	fmt.Println("Successfully Stop Instance")
	return &pb.Response{}, nil
}

func (l Service) startInstance() (string, error) {
	lptr := (ec2.EC2)(l)
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}

	//Init TeleBot
	//bot, _ := NewTeleBot()

	_, err := lptr.StartInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatal(err)
		}
		return "", err
	}

	// Make a Describe Request to Get Result
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	result, err := l.describeInstance(describeInput, "running")
	instancePublicIP := result.Reservations[0].Instances[0].PublicIpAddress
	msg := "Current Instance Type is " + *result.Reservations[0].Instances[0].InstanceType + "\n" + *instancePublicIP

	//bot.Send(msg)
	return msg, nil
}

func (l Service) StartInstance(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	// Init Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	msg, err := l.startInstance()
	if (err != nil) {
		fmt.Println(err)
	}
	fmt.Println("Successfully Start Instance")

	bot.Send(tgbotapi.NewMessage(int64(ChatID), msg))
	return &pb.Response{}, nil
}

func (l Service) DescribeInstance(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	var instancePublicIP *string
	lptr := (ec2.EC2)(l)
	bot, _ := NewTeleBot()
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	describeOutput, err := lptr.DescribeInstances(describeInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err)
		}
		return &pb.Response{}, err
	}
	fmt.Println(describeOutput)
	if *describeOutput.Reservations[0].Instances[0].State.Name == "stopped" {
		msg := "Instance Stopped!"
		bot.Send(tgbotapi.NewMessage(int64(ChatID), msg))
		return &pb.Response{}, nil
	}
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	instanceType := describeOutput.Reservations[0].Instances[0].InstanceType
	msg := tgbotapi.NewMessage(int64(ChatID), *instanceType + " " + *instancePublicIP)
	bot.Send(msg)
	return &pb.Response{}, nil
}

func (l Service) modifyInstance(instanceAttribute string) error {
	lptr := (ec2.EC2)(l)
	// Modify Instance Type
	modifyInput := &ec2.ModifyInstanceAttributeInput{
		InstanceId: InstanceID,
		InstanceType: &ec2.AttributeValue{
			Value: &instanceAttribute,
		},
	}

	_, err := lptr.ModifyInstanceAttribute(modifyInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err)
		}
		return err
	}
	fmt.Println("Modify Instance Type Successfully")
	return nil
}

// ModifyUpInstance event: StopInstance ==> Change Instance Attribute to "t2.xlarge"
// ==> Start Instance ==> Send Instance Public IP to Telegram Bot
func (l Service) ModifyUpInstance(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	//lptr := (ec2.EC2)(l)
	var instancePublicIP *string

	// Init Telegram Bot
	bot, _ := NewTeleBot()

	// To stop instance
	_, err := l.stopInstance()
	if err != nil {
		log.Fatal(err)
	}

	// Get Instance Status ==> if Status == Stopped ==> ModifyInstanceType
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	_, err = l.describeInstance(describeInput, "stopped")
	if (err != nil) {
		fmt.Println(err)
	}

	// Modify Instance Type
	err = l.modifyInstance(UpInstanceType)
	if err != nil {
		fmt.Println(err)
	}

	msg, err := l.startInstance()
	if err != nil {
		log.Fatal(err)
	}
	bot.Send(tgbotapi.NewMessage(int64(ChatID), msg))

	// Check Start Instance State
	describeOutput, err := l.describeInstance(describeInput, "running")
	if err != nil {
		fmt.Println(err)
	}
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	bot.Send(tgbotapi.NewMessage(int64(ChatID), *instancePublicIP))
	return &pb.Response{}, nil
}

func (l Service) ModifyDownInstance(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	var instancePublicIP *string

	// Init Telegram Bot
	bot, _ := NewTeleBot()

	// To stop instance
	_, err := l.stopInstance()
	if err != nil {
		log.Fatal(err)
	}

	// Get Instance Status ==> if Status == Stopped ==> ModifyInstanceType
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	_, err = l.describeInstance(describeInput, "stopped")
	if (err != nil) {
		fmt.Println(err)
	}

	// Modify Instance Type
	err = l.modifyInstance(DownInstanceType)
	if err != nil {
		fmt.Println(err)
	}

	msg, err := l.startInstance()
	if err != nil {
		log.Fatal(err)
	}
	bot.Send(tgbotapi.NewMessage(int64(ChatID), msg))

	// Check Start Instance State
	describeOutput, err := l.describeInstance(describeInput, "running")
	if err != nil {
		fmt.Println(err)
	}
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	instanceType := describeOutput.Reservations[0].Instances[0].InstanceType
	fmt.Printf("IP is %s\n", *instancePublicIP)
	bot.Send(tgbotapi.NewMessage(int64(ChatID), *instancePublicIP + *instanceType))
	return &pb.Response{}, nil
}

func (l Service) TGStartInstance(text string) (message string) {
	fmt.Println("Enter TGStartInstance")
	if text == "start" {
		fmt.Println("start instance")
		msg, err := l.startInstance()
		if err != nil {
			log.Fatalf("TGStartInstance error %v", err)
		}
		return msg
	}
	return "No such Method"
}

func initBot() (*tb.Bot){
	l := NewEC2Sess()
	fmt.Println("start init telegram bot")
	bot, err := tb.NewBot(tb.Settings{
		Token:  BotToken,
		Poller: &tb.LongPoller{Timeout: 1 * time.Second},
	})

	if err != nil {
		fmt.Printf("NewBot err %v\n", err)
	}

	bot.Handle(tb.OnText, func(m *tb.Message) {
		bot.Send(m.Sender, l.TGStartInstance(m.Text))
	})
	return bot
}

func main() {
	// listen on port 12345
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal("Listen TCP error: ", err)
	}

	// start TGBot
	go func() {
		//defer wg.Done()
		Bot = initBot()
		Bot.Start()
	}()

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