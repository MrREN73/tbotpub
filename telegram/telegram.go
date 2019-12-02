package telegram

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	tg "github.com/Syfaro/telegram-bot-api"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"golang.org/x/net/proxy"
)

type Config struct {
	Token         string
	ProxyUser     string
	ProxyPassword string
	ProxyURL      string
	Debug         bool
}

type Telegram struct {
	bot   *tgbotapi.BotAPI
	token string
}

func getProxyClient(cfg Config) *http.Client {
	auth := &proxy.Auth{
		User:     cfg.ProxyUser,
		Password: cfg.ProxyPassword,
	}

	dialSocksProxy, err := proxy.SOCKS5("tcp", cfg.ProxyURL, auth, proxy.Direct)
	if err != nil {
		log.Printf("Error connecting to proxy: %+v", err)
	}

	client := http.DefaultClient
	client.Transport = &http.Transport{Dial: dialSocksProxy.Dial}

	return client
}

func New(cfg Config) (*Telegram, error) {
	bot, err := tg.NewBotAPIWithClient(cfg.Token, getProxyClient(cfg))
	if err != nil {
		return nil, err
	}

	bot.Debug = cfg.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Telegram{
		bot:   bot,
		token: cfg.Token,
	}, nil
}

func (t Telegram) UpdateChan() (tg.UpdatesChannel, error) {
	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	return t.bot.GetUpdatesChan(ucfg)
}

func (t Telegram) GetFile(fileID string) (string, error) {
	fileConfig := tgbotapi.FileConfig{
		FileID: fileID,
	}

	fileCon, err := t.bot.GetFile(fileConfig)
	if err != nil {
		return "", err
	}

	return fileCon.FilePath, nil
}

func (t Telegram) GetFileURL(filePath string) string {
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", t.token, filePath)
}

type Chat struct {
	t        *Telegram
	ChatID   int64
	UserName string
}

func (t *Telegram) NewChat(msg *tg.Message) (*Chat, error) {
	if msg == nil {
		return nil, errors.New("no message")
	}

	chatID := int64(0)
	if msg.Chat != nil {
		chatID = msg.Chat.ID
	}

	if chatID == 0 {
		return nil, errors.New("no chat ID")
	}

	userName := "<anonymous>"
	if msg.From != nil {
		userName = msg.From.String()
	}

	return &Chat{
		t:        t,
		ChatID:   chatID,
		UserName: userName,
	}, nil
}

func (c Chat) Send(message string) error {
	msg := tg.NewMessage(c.ChatID, message)
	msg.DisableWebPagePreview = true

	_, err := c.t.bot.Send(msg)

	return err
}

func (c Chat) SendErr(err error) {
	log.Println(err)

	_ = c.Send("упс, что-то пошло не так")
}

func (c Chat) SendAsync(in chan string, threads int) {
	wg := &sync.WaitGroup{}
	wg.Add(threads)

	for i := 0; i < threads; i++ {
		go func() {
			for j := range in {
				text := fmt.Sprintf("%s - %d \n", j, len(j))
				if err := c.Send(text); err != nil {
					return
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	log.Println("Complete")
}
