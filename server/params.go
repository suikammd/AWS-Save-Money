package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"net/rpc"
)

const ServiceName = "Service"

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

type ServiceInterface interface {
	StopInstance(line []byte, ack *bool) error
	StartInstance(line []byte, ack *bool) error
	DescribeInstance(line []byte, ack *bool) error
	ModifyUpInstance(line []byte, ack *bool) error
	ModifyDownInstance(line []byte, ack *bool) error
}

func RegisterService(svc ServiceInterface) error {
	return rpc.RegisterName(ServiceName, svc)
}