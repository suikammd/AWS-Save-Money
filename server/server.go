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

type Listener ec2.EC2

func New() (l Listener){
	// create EC2 service client
	sess, _ := session.NewSession(&aws.Config{
		Region:      Region,
		Credentials: credentials.NewStaticCredentials(AWSID, AWSSecret, ""),
	})

	l = (Listener)(*ec2.New(sess))
	return
}

func (l Listener) stopInstance() (*ec2.StopInstancesOutput, error) {
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

func (l Listener) StopInstance(line []byte, ack *bool) error {
	_, err := l.stopInstance()
	if (err != nil) {
		fmt.Println(err)
	}
	fmt.Println("Successfully Stop Instance")
	return nil
}

func (l Listener) startInstance() error {
	lptr := (ec2.EC2)(l)
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			InstanceID,
		},
	}

	var result *ec2.StartInstancesOutput
	result, err := lptr.StartInstances(input)
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

	fmt.Println(result)
	return nil
}

func (l Listener) StartInstance(line []byte, ack *bool) error {
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

func (l Listener) describeInstance(describeInput *ec2.DescribeInstancesInput, state string) (*ec2.DescribeInstancesOutput, error) {
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

func (l Listener) DescribeInstance(line []byte, ack *bool) error {
	lptr := (ec2.EC2)(l)
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
	return nil
}

func (l Listener) modifyInstance(instanceAttribute string) error {
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
func (l Listener) ModifyUpInstance(line []byte, ack *bool) error {
	//lptr := (ec2.EC2)(l)
	var instancePublicIP *string

	// Init Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// To stop instance
	_, err = l.stopInstance()
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
	fmt.Printf("IP is %s\n", *instancePublicIP)
	msg := tgbotapi.NewMessage(int64(ChatID), *instancePublicIP)
	bot.Send(msg)
	return nil
}

func (l Listener) ModifyDownInstance(line []byte, ack *bool) error {
	var instancePublicIP *string

	// Init Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// To stop instance
	_, err = l.stopInstance()
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
	fmt.Printf("IP is %s\n", *instancePublicIP)
	msg := tgbotapi.NewMessage(int64(ChatID), *instancePublicIP)
	bot.Send(msg)
	return nil
	
}

func main() {
	// listen on port 42586
	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	var listener Listener
	listener = New()
	rpc.Register(listener)
	rpc.Accept(inbound)
}