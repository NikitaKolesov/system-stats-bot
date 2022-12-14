package main

import (
	"bytes"
	"github.com/ricochet2200/go-disk-usage/du"
	"log"
	"os"
	"strconv"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var funcMap = template.FuncMap{
	"humanSize":      humanSize,
	"percentConvert": percentConvert,
}

type MessageData struct {
	Hostname              string
	Size, Used, Available uint64
	Usage                 float32
}

const usageMessage = `Hostname: {{ .Hostname }}
Size: {{ humanSize .Size }}
Used: {{ humanSize .Used }}
Available: {{ humanSize .Available }}
Usage: {{ percentConvert .Usage }}`

func humanSize(s uint64) string {
	return strconv.FormatUint(s/uint64(1073741824), 10) + " GB"
}

func percentConvert(f float32) string {
	return strconv.FormatFloat(float64(f)*100, 'f', 0, 64) + "%"
}

func getHostname() (string, error) {
	hostname := os.Getenv("HOSTNAME_OVERRIDE")
	if hostname == "" {
		return os.Hostname()
	}
	return hostname, nil
}

func getChatId() int64 {
	chatIdStr := os.Getenv("STATS_CHAT_ID")
	if chatIdStr == "" {
		log.Fatal("STATS_CHAT_ID environment variable is not set!")
	}
	chatIdInt, err := strconv.Atoi(chatIdStr)
	if err != nil {
		log.Fatal("Failed to convert chat_id to integer")
	}
	return int64(chatIdInt)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	hostname, err := getHostname()
	if err != nil {
		log.Panic(err)
	}

	usage := du.NewDiskUsage(".")

	messageData := MessageData{
		Hostname:  hostname,
		Size:      usage.Size(),
		Used:      usage.Used(),
		Available: usage.Available(),
		Usage:     usage.Usage(),
	}

	t := template.Must(template.New("").Funcs(funcMap).Parse(usageMessage))

	var tpl bytes.Buffer

	err = t.Execute(&tpl, messageData)
	if err != nil {
		log.Println("executing template:", err)
	}

	//bot.Debug = true

	if err != nil {
		log.Fatal("STATS_CHAT_ID environment variable is not set!")
	}

	if os.Getenv("STATS_ALERT_ONLY") != "" && messageData.Usage < 0.9 {
		return
	}

	msg := tgbotapi.NewMessage(getChatId(), tpl.String())
	_, err = bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	//updateConfig := tgbotapi.NewUpdate(0)
	//updateConfig.Timeout = 30
	//updates := bot.GetUpdatesChan(updateConfig)
	//for update := range updates {
	//	// Telegram can send many types of updates depending on what your Bot
	//	// is up to. We only want to look at messages for now, so we can
	//	// discard any other updates.
	//	if update.Message == nil {
	//		continue
	//	}
	//
	//	// Now that we know we've gotten a new message, we can construct a
	//	// reply! We'll take the Chat ID and Text from the incoming message
	//	// and use it to create a new message.
	//	msg = tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//	bot.Send(msg)
	//}
}
