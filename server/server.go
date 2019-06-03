package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

type Service ec2.EC2

func NewEC2Sess() (l Service){
	// create EC2 service client
	sess, _ := session.NewSession(&aws.Config{
		Region:      Region,
		Credentials: credentials.NewStaticCredentials(AWSID, AWSSecret, ""),
	})

	l = (Service)(*ec2.New(sess))
	return
}

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

func (l Service) StopInstance(line []byte, ack *bool) error {
	_, err := l.stopInstance()
	if (err != nil) {
		fmt.Println(err)
	}
	fmt.Println("Successfully Stop Instance")
	return nil
}

func (l Service) startInstance() error {
	lptr := (ec2.EC2)(l)
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}

	//Init TeleBot
	bot, _ := NewTeleBot()

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
		return err
	}

	// Make a Describe Request to Get Result
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	result, err := l.describeInstance(describeInput, "running")
	msg := tgbotapi.NewMessage(int64(ChatID), "Current Instance Type is " + *result.Reservations[0].Instances[0].InstanceType)
	bot.Send(msg)
	return nil
}

func (l Service) StartInstance(line []byte, ack *bool) error {
	var instancePublicIP *string
	// Init Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	err = l.startInstance()
	if (err != nil) {
		fmt.Println(err)
	}
	fmt.Println("Successfully Start Instance")

	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}
	result, err := l.describeInstance(describeInput, "running")
	instancePublicIP = result.Reservations[0].Instances[0].PublicIpAddress
	msg := tgbotapi.NewMessage(int64(ChatID), *instancePublicIP)
	bot.Send(msg)
	return nil
}

func (l Service) DescribeInstance(line []byte, ack *bool) error {
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
		return err
	}
	fmt.Println(describeOutput)
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	instanceType := describeOutput.Reservations[0].Instances[0].InstanceType
	msg := tgbotapi.NewMessage(int64(ChatID), *instanceType + " " + *instancePublicIP)
	bot.Send(msg)
	return nil
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
func (l Service) ModifyUpInstance(line []byte, ack *bool) error {
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

	err = l.startInstance()
	if err != nil {
		log.Fatal(err)
	}

	// Check Start Instance State
	describeOutput, err := l.describeInstance(describeInput, "running")
	if err != nil {
		fmt.Println(err)
	}
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	instanceType := describeOutput.Reservations[0].Instances[0].InstanceType
	fmt.Printf("IP is %s\n", *instancePublicIP)
	msg := tgbotapi.NewMessage(int64(ChatID), *instanceType + *instancePublicIP)
	bot.Send(msg)
	return nil
}

func (l Service) ModifyDownInstance(line []byte, ack *bool) error {
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

	err = l.startInstance()
	if err != nil {
		log.Fatal(err)
	}

	// Check Start Instance State
	describeOutput, err := l.describeInstance(describeInput, "running")
	if err != nil {
		fmt.Println(err)
	}
	instancePublicIP = describeOutput.Reservations[0].Instances[0].PublicIpAddress
	instanceType := describeOutput.Reservations[0].Instances[0].InstanceType
	fmt.Printf("IP is %s\n", *instancePublicIP)
	msg := tgbotapi.NewMessage(int64(ChatID), *instanceType + *instancePublicIP)
	bot.Send(msg)
	return nil
}

func main() {
	// Rester Service
	RegisterService(NewEC2Sess())

	// listen on port 1234
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen TCP error: ", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Accept error")
	}

	rpc.ServeConn(conn)
}