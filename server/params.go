package main

import (
	"github.com/aws/aws-sdk-go/aws"
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