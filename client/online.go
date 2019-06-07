package main

import (
	"os/exec"
	"fmt"
	"log"
	"strconv"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"time"
)

var ChatID           int     = 12345678
var BotToken         string  = ""


func main() {
	bot, _ := tgbotapi.NewBotAPI(BotToken)
	cmd := exec.Command("/bin/bash", "-c", `who | wc -l`)
	count := 1
	for count == 0 {
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
	msg := "shut down"
	bot.Send(tgbotapi.NewMessage(int64(ChatID), msg))
	cmd = exec.Command("/bin/bash", "-c", "sudo poweroff")
	cmd.Run()
}