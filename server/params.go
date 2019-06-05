package main

import (
	"github.com/aws/aws-sdk-go/aws"
	tb "gopkg.in/tucnak/telebot.v2"
)

const ServiceName = "Service"

var (
	Bot              *tb.Bot
	InstanceID       *string = aws.String("")
	BotToken         string  = ""
	UpInstanceType   string  = ""
	DownInstanceType string  = ""
	ChatID           int     = 123456
	Region           *string = aws.String("")
	AWSID            string  = ""
	AWSSecret        string  = ""
)
