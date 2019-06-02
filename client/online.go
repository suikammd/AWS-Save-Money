package main

import (
	"os/exec"
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

var ChatID           int     = 12345678
var BotToken         string  = ""

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

func StopCommand() {
	client, err := rpc.Dial("tcp", "3.112.232.91:42586")
	if err != nil {
		log.Fatal(err)
	}

	// Init Telegram Bot
	bot, _ := NewTeleBot()

	var line []byte
	var reply bool

	err = client.Call("Listener.StopInstance", line, &reply)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Call("Listener.DescribeInstance", line, &reply)
	if err != nil {
		log.Fatal(err)
	}
	msg := tgbotapi.NewMessage(int64(ChatID), "Successfully Stop Instance")
	bot.Send(msg)
	fmt.Println("Success to Stop Instance")
}

func main() {
	cmd := exec.Command("/bin/bash", "-c", `who | wc -l`)
	count := 1
	for count != 0 {
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out)

		count, err = strconv.Atoi(string(out[0]))
		if (err != nil) {
			log.Fatal("Failed to Change")
		}
		fmt.Println(count)
		time.Sleep(2 * time.Minute)
	}
	StopCommand()
}
