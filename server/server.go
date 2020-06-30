package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/spf13/viper"
)

var slackAPI *slack.Client
var lineAPI *linebot.Client
var telegramAPI *tgbotapi.BotAPI

// Start Requests handler HTTP Server.
func Start(logDir string) error {

	// Initial the bot api
	slackAPI = slack.New(viper.GetString("slack_oauth_token"))

	// Initail telegram
	telegramAPI, _ = tgbotapi.NewBotAPI(viper.GetString("telegram_token"))
	_, err := telegramAPI.SetWebhook(tgbotapi.NewWebhook(viper.GetString("telegram_webhook_url")))
	if err != nil {
		log.Fatal(err)
	}
	info, err := telegramAPI.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("[Telegram callback failed] %s", info.LastErrorMessage)
	}

	// Initial line bot
	lineAPI, _ = linebot.New(viper.GetString("line_secret"), viper.GetString("line_token"))

	// Run echo web server
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// setup access logger
	logPath := filepath.Join(logDir, "httpd.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if _, err := os.Create(logPath); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
		Output: f,
	}))

	// Routes
	e.GET("/health", health())
	e.POST("/api/slack", slackHandler())
	e.POST("/api/line", lineHandler())
	e.POST("/api/telegram", telegramHandler())

	bindURL := fmt.Sprintf("%s:%s", viper.GetString("bind"), viper.GetString("port"))
	log.Infof("Listening on %s", bindURL)

	return e.Start(bindURL)
}

// Handler
func health() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}
}

func slackHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		buf := new(bytes.Buffer)
		buf.ReadFrom(c.Request().Body)
		body := buf.String()

		eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: viper.GetString("slack_token")}))
		if e != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
			}
			return c.String(http.StatusOK, r.Challenge)
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				if ev.Username != "writeupBot" {
					user, err := slackAPI.GetUserInfo(ev.User)
					if err != nil {
						log.Errorf("[ERROR] %s\n", err)
					}
					log.Infof("[INFO] Try to send message %s %s\n", ev.Text, user.RealName)
					AMF(SLACK, ev.Text, user.RealName)
					return c.String(http.StatusOK, "")
				}
			}
		}

		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
	}
}

func lineHandler() echo.HandlerFunc {
	return func(c echo.Context) error {

		events, err := lineAPI.ParseRequest(c.Request())
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					res, err := lineAPI.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
					if err != nil {
						return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
					}
					AMF(LINE, message.Text, res.DisplayName)
				}
			}
		}
		return c.String(http.StatusOK, "")
	}
}

func telegramHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		bytes, err := ioutil.ReadAll(c.Request().Body)

		var update tgbotapi.Update
		err = json.Unmarshal(bytes, &update)
		if err != nil {
			log.Error(err)
		}

		log.Infof("From: %s Text: %s\n", update.Message.From.UserName, update.Message.Text)
		AMF(TELEGRAM, update.Message.Text, update.Message.From.UserName)

		return c.String(http.StatusOK, "")
	}
}

type ChatApp int

const (
	LINE     ChatApp = 1
	TELEGRAM ChatApp = 2
	SLACK    ChatApp = 3
)

func (c ChatApp) String() string {
	switch c {
	case LINE:
		return "LINE"
	case TELEGRAM:
		return "TELEGRAM"
	case SLACK:
		return "SLACK"
	default:
		return fmt.Sprintf("%d", int(c))
	}
}

// AMF handle main function
func AMF(which ChatApp, MsgText string, User string) {
	log.Infof("[%s] Receive \"%s\" from %s.", ChatApp(which), MsgText, User)
	msg := MsgText + "\n<FROM [" + ChatApp(which).String() + "] BY " + User + ">"
	switch which {
	case LINE:
		TelegramSendMsg(msg)
		SlackSendMsg(msg)
	case TELEGRAM:
		LineSendMsg(msg)
		SlackSendMsg(msg)
	case SLACK:
		TelegramSendMsg(msg)
		LineSendMsg(msg)
	}
}

// TelegramSendMsg Send Messgae to Telegram
func TelegramSendMsg(MsgText string) {
	if _, err := telegramAPI.Send(tgbotapi.NewMessage(viper.GetInt64("telegram_chat_id"), MsgText)); err != nil {
		log.Errorf("[ERROR] %s\n", err)
	}
	log.Infof("[Telegram] Send %s", MsgText)
}

// LineSendMsg Send Messgae to Line
func LineSendMsg(MsgText string) {
	if _, err := lineAPI.PushMessage(viper.GetString("line_group_id"), linebot.NewTextMessage(MsgText)).Do(); err != nil {
		log.Errorf("[ERROR] %s\n", err)
	}
	log.Infof("[LINE] Send %s", MsgText)
}

// SlackSendMsg Send Messgae to Slack
func SlackSendMsg(MsgText string) {
	_, _, err := slackAPI.PostMessage(viper.GetString("slack_channel_id"), slack.MsgOptionText(MsgText, false))
	if err != nil {
		log.Errorf("[ERROR] %s\n", err)
	}
	log.Infof("[Slack] Send %s", MsgText)
}
