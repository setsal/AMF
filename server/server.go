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

// var slackAPI *slack.Client
// var lineAPI *linebot.Client
// var telegramAPI *tgbotapi.BotAPI

// InitalBot then return  我嘗試在切開一點點看看
func InitialBots() (sApi *slack.Client, lApi *linebot.Client, tApi *tgbotapi.BotAPI) {
	// Initial the bot api
	sApi = InitialSlackBot(viper.GetString("slack_oauth_token"))
	tApi, err := InitialTelegramBot(viper.GetString("telegram_token"), viper.GetString("telegram_webhook_url"))
	if err != nil {
		log.Fatal(err)
	}
	lApi, err = InitialLineBot(viper.GetString("line_secret"), viper.GetString("line_token"))
	if err != nil {
		log.Fatal(err)
	}
	return
}
func InitialSlackBot(Token string) *slack.Client {
	// Initial the bot api
	return slack.New(Token)
}

func InitialTelegramBot(Token string, WebhookURL string) (*tgbotapi.BotAPI, error) {
	// Initail telegram
	tApi, _ := tgbotapi.NewBotAPI(Token)
	_, err := tApi.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		log.Fatal(err)
	}
	info, err := tApi.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("[Telegram callback failed] %s", info.LastErrorMessage)
	}

	return tApi, err
}

func InitialLineBot(Secret string, Token string) (*linebot.Client, error) {
	// Initial line bot
	return linebot.New(Secret, Token)
}

// Start Requests handler HTTP Server.
func Start(logDir string) error {
	// InitalBot then return  我嘗試在切開一點點看看
	slackAPI, lineAPI, telegramAPI := InitialBots()

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
	e.POST("/api/slack", slackHandler(slackAPI, lineAPI, telegramAPI))
	e.POST("/api/line", lineHandler(slackAPI, lineAPI, telegramAPI))
	e.POST("/api/telegram", telegramHandler(slackAPI, lineAPI, telegramAPI))

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

func slackHandler(slackAPI *slack.Client, lineAPI *linebot.Client, telegramAPI *tgbotapi.BotAPI) echo.HandlerFunc {
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
				if ev.Username != "AMF" {
					log.Infof("[INFO] %+v\n", ev.User)
					user, err := slackAPI.GetUserInfo(ev.User)
					if err != nil {
						log.Errorf("[ERROR] %s\n", err)
					}
					log.Infof("[INFO] Try to send message %s %s\n", ev.Text, user.RealName)
					AMF(SLACK, ev.Text, user.RealName, slackAPI, lineAPI, telegramAPI)
					return c.String(http.StatusOK, "")
				}
			}
		}

		return c.String(http.StatusOK, "")
	}
}

func lineHandler(slackAPI *slack.Client, lineAPI *linebot.Client, telegramAPI *tgbotapi.BotAPI) echo.HandlerFunc {
	return func(c echo.Context) error {

		events, err := lineAPI.ParseRequest(c.Request())
		// return c.String(http.StatusOK, "")
		if err != nil {
			// return c.String(http.StatusOK, "")
			if err == linebot.ErrInvalidSignature {
				return echo.NewHTTPError(400, "")
				//w.WriteHeader(400)
			} else {
				return echo.NewHTTPError(500, "")
				//w.WriteHeader(500)
			}
			//return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
		}
		// return c.String(http.StatusOK, "")
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					res, err := lineAPI.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do()
					if err != nil {
						res, err = lineAPI.GetProfile(event.Source.UserID).Do()
						if err != nil {
							//??
						}
						// return c.String(http.StatusOK, "")
						//return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
					}
					AMF(LINE, message.Text, res.DisplayName, slackAPI, lineAPI, telegramAPI)
				}
			}
		}
		return c.String(http.StatusOK, "")
	}
}

func telegramHandler(slackAPI *slack.Client, lineAPI *linebot.Client, telegramAPI *tgbotapi.BotAPI) echo.HandlerFunc {
	return func(c echo.Context) error {
		bytes, err := ioutil.ReadAll(c.Request().Body)

		var update tgbotapi.Update
		err = json.Unmarshal(bytes, &update)
		if err != nil {
			log.Error(err)
		}

		userName := update.Message.From.LastName + " " + update.Message.From.FirstName
		log.Infof("From: %s Text: %s\n", userName, update.Message.Text)
		AMF(TELEGRAM, update.Message.Text, userName, slackAPI, lineAPI, telegramAPI)

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
func AMF(which ChatApp, MsgText string, User string, slackAPI *slack.Client, lineAPI *linebot.Client, telegramAPI *tgbotapi.BotAPI) string {
	log.Infof("[%s] Receive \"%s\" from %s.", ChatApp(which), MsgText, User)
	msg := MsgText + "\n> [" + ChatApp(which).String() + "] " + User + "📍"
	switch which {
	case LINE:
		TelegramSendMsg(telegramAPI, msg)
		SlackSendMsg(slackAPI, msg)
	case TELEGRAM:
		LineSendMsg(lineAPI, linebot.NewTextMessage(msg))
		SlackSendMsg(slackAPI, msg)
	case SLACK:
		TelegramSendMsg(telegramAPI, msg)
		LineSendMsg(lineAPI, linebot.NewTextMessage(msg))
	}
	return msg
}

// TelegramSendMsg Send Messgae to Telegram
func TelegramSendMsg(telegramAPI *tgbotapi.BotAPI, MsgText string) (tgbotapi.Message, error) {
	res, err := telegramAPI.Send(tgbotapi.NewMessage(viper.GetInt64("telegram_chat_id"), MsgText))
	if err != nil {
		log.Errorf("[ERROR] %s\n", err)
		return res, err
	}
	log.Infof("[Telegram] Send %s", MsgText)
	return res, nil
}

// LineSendMsg Send Messgae to Line
func LineSendMsg(lineAPI *linebot.Client, TextMessage *linebot.TextMessage) error {
	_, err := lineAPI.PushMessage(viper.GetString("line_group_id"), TextMessage).Do()
	if err != nil {
		log.Errorf("[ERROR] %s\n", err)
		return err
	}
	log.Infof("[LINE] Send %s", TextMessage.Text)
	return nil
}

// SlackSendMsg Send Messgae to Slack
func SlackSendMsg(slackAPI *slack.Client, MsgText string) error {
	if _, _, err := slackAPI.PostMessage(viper.GetString("slack_channel_id"), slack.MsgOptionText(MsgText, false)); err != nil {
		log.Errorf("[ERROR] %s\n", err)
		return err
	}
	log.Infof("[Slack] Send %s", MsgText)
	return nil
}
