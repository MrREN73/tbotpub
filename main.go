package main

import (
	"fmt"
	"github.com/mrren73/tbotv2.0/files"
	"github.com/mrren73/tbotv2.0/telegram"
	"log"
	"strconv"
)

func welcomeMessage(user string) string {
	return fmt.Sprintf(
		"Добро пожаловать в мой мир, %s! Не удивляйся, откуда я знаю твое имя. "+
			"Я немного волшебник! Хочешь еще немного магии? Тогда пришли мне текстовый файл "+
			"со списком URL и в описание напиши количество желаемых одновременных потоков проверки "+
			"этих самых URL", user)
}

func getConfig() telegram.Config {
	return telegram.Config{
		Debug:         false,
		Token:         "YOUR_TOKEN",
		ProxyURL:      "YOUR_PROXY_ADDR",
		ProxyUser:     "YUOR_PROXY_USER",
		ProxyPassword: "YOUR_PROXY_PASS",
	}
}

//nolint:funlen
func main() {
	cfg := getConfig()

	bot, err := telegram.New(cfg)
	if err != nil {
		log.Panic(err)
	}

	updates, err := bot.UpdateChan()
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		chat, err := bot.NewChat(update.Message)
		if err != nil {
			continue
		}

		switch update.Message.Command() {
		case "start":
			text := welcomeMessage(chat.UserName)
			_ = chat.Send(text)
		default:
			if update.Message.Document == nil {
				text := fmt.Sprintf("Мне нечего тебе сказать на такое, знаешь ли, %s", chat.UserName)
				_ = chat.Send(text)

				continue
			}

			threadCountStr := update.Message.Caption

			threadCount, err := strconv.Atoi(threadCountStr)
			if err != nil {
				log.Printf("Error %v from User %s", err, chat.UserName)

				text := "Вы ввели не число в описании файла. Установленно значение по умолчанию - 1"
				_ = chat.Send(text)
				threadCount = 1
			}

			filePath, err := bot.GetFile(update.Message.Document.FileID)
			if err != nil {
				chat.SendErr(err)
				continue
			}

			fp := files.GetPath(filePath)
			url := bot.GetFileURL(filePath)

			if err := files.Download(url, fp); err != nil {
				chat.SendErr(err)
				continue
			}

			ch, err := files.ScannerChan(fp, threadCount)
			if err != nil {
				chat.SendErr(err)
				continue
			}

			chat.SendAsync(ch, threadCount)
		}
	}
}
