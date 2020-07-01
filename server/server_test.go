package server

import (
	//"bytes"
	//"fmt"

	"log"

	//"os"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/slack-go/slack"

	//"github.com/slack-go/slack"

	"testing"
	//"github.com/setsal/go-AMF/server"
	"github.com/spf13/viper"
)

// 這樣執行 Test
// go test server/server_test.go  -v
// Ｊ一個？
// go test server/server_test.go server/server.go -v

// 沒辦法 只好寫這樣QQ  切太開有夠難測 不執行這個單獨測試 server 會導致 config 沒辦法 init
// 因為 init 是透過 root sub command 間接執行的
func init() { //用這個名字可以不用額外呼叫

	// bind server listen
	viper.SetDefault("bind", "127.0.0.1")
	viper.SetDefault("port", "3000")

	// add config path in current project dir
	viper.SetConfigType("yaml")

	viper.SetConfigName(".go-AMF")
	viper.AddConfigPath("/home/sqlab/st/go-AMF")

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal(err)
		//panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	viper.AutomaticEnv() // read in environment variables that match
}

/*
func TestSendMsg(t *testing.T) {
	initConfig()

	// 這個是為了讓主程式的 stdout 能倒向 test 的 output, 方便 debug
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	//server.Start("/home/sqlab/st/go-AMF")

	// 主 function 執行
	// server.Start("/home/sqlab/st/go-AMF")

	// 還是我們應該在切更細一點？
	// 結束 output
	t.Log(buf.String())

}
*/
func TestTelegramSendMsg(t *testing.T) {

	//TO DO
	msgText := "TEST TELEGRAM MSG"

	telegramAPI, err := InitialTelegramBot(viper.GetString("telegram_token"), viper.GetString("telegram_webhook_url"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := TelegramSendMsg(telegramAPI, msgText)
	if err != nil {
		t.Fatal(err)
	}

	if res.Text != msgText {
		t.Fatalf("Error %v; want %v", res.Text, msgText)
	}
}

func TestLineSendMsg(t *testing.T) {
	//TO DO
	msgText := "TEST LINE MSG"
	msgID := "325708"

	lineAPI, err := InitialLineBot(viper.GetString("line_secret"), viper.GetString("line_token"))
	if err != nil {
		t.Fatal(err)
	}

	if LineSendMsg(lineAPI, &linebot.TextMessage{ID: msgID, Text: msgText}) != nil {
		t.Fatal(err)
	}

	//APIError 404 Not found 抓不到ＲＲＲ
	/*
		res, err := lineAPI.GetMessageContent(msgID).Do()
		if err != nil {
			t.Fatal(err)
		}

		t.Fatalf("Error %v; want %v",res.ContentType , msgText)
	*/
}

func TestSlackSendMsg(t *testing.T) {

	msgText := "TEST SLACK MSG"
	slackAPI := InitialSlackBot(viper.GetString("slack_oauth_token"))

	err := SlackSendMsg(slackAPI, msgText)
	if err != nil {
		t.Fatal(err)
	}

	params := slack.GetConversationHistoryParameters{
		ChannelID: viper.GetString("slack_channel_id"),
		Inclusive: false,
		Latest:    "",
		Oldest:    "0",
		Limit:     1,
	}
	res, err := slackAPI.GetConversationHistory(&params)
	if err != nil {
		t.Fatal(err)
	}

	if res.Messages[0].Msg.Text != msgText {
		t.Fatalf("Error %s; want %s", res.Messages[0].Msg.Text, msgText)
	}

}

//TestMessageAppendWithPlatform
func TestAMF(t *testing.T) {
	//TO DO
	msgText := "TEST AMF MSG"
	msgUser := "AMF"

	slackAPI := InitialSlackBot(viper.GetString("slack_oauth_token"))
	lineAPI, err := InitialLineBot(viper.GetString("line_secret"), viper.GetString("line_token"))
	if err != nil {
		t.Fatal(err)
	}
	telegramAPI, err := InitialTelegramBot(viper.GetString("telegram_token"), viper.GetString("telegram_webhook_url"))
	if err != nil {
		t.Fatal(err)
	}

	//為啥米要三個API都當input
	/*用空的interface嗎 還是不要浪費時間ㄎ
	type Api interface {
	}
	*/
	if AMF(SLACK, msgText, msgUser, slackAPI, lineAPI, telegramAPI) == msgText {
		t.Fatalf("Error %v", msgText)
	}
	if AMF(LINE, msgText, msgUser, slackAPI, lineAPI, telegramAPI) == msgText {
		t.Fatalf("Error %v", msgText)
	}
	if AMF(TELEGRAM, msgText, msgUser, slackAPI, lineAPI, telegramAPI) == msgText {
		t.Fatalf("Error %v", msgText)
	}
}

//TestReturnSupportedPlatforms 未實作

//TestVerifyBOT
func TestInitialSlackBot(t *testing.T) {
	_ = InitialSlackBot(viper.GetString("slack_oauth_token"))
	//??
}
func TestInitialTelegramBot(t *testing.T) {
	_, err := InitialTelegramBot(viper.GetString("telegram_token"), viper.GetString("telegram_webhook_url"))
	if err != nil {
		t.Fatal(err)
	}
}
func TestInitialLineBot(t *testing.T) {
	_, err := InitialLineBot(viper.GetString("line_secret"), viper.GetString("line_token"))
	if err != nil {
		t.Fatal(err)
	}
}

//TestServerStatus 未實作

// // Test Handler  大失敗QQ
// func TestAMFslackHandler(t *testing.T) {
// 	// Make channels to pass fatal errors in WaitGroup
// 	fatalErrors := make(chan error)
// 	wgDone := make(chan bool)

// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	go func() {
// 		time.Sleep(time.Duration(10) * time.Second)
// 		msgText := "TEST HANDLER"
// 		fmt.Println("TRY TO SEND")
// 		slackAPI := InitialSlackBot(viper.GetString("slack_oauth_token"))

// 		err := SlackSendMsg(slackAPI, msgText)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		wg.Done()
// 	}()

// 	go func() {
// 		log.Println("Waitgroup 2")
// 		// Example function which returns an error
// 		err := Start("/home/sqlab/st/go-AMF")
// 		if err != nil {
// 			return
// 		}
// 		wg.Done()
// 	}()

// 	// Important final goroutine to wait until WaitGroup is done
// 	go func() {
// 		wg.Wait()
// 		close(wgDone)
// 	}()

// 	select {
// 	case <-wgDone:
// 		// carry on
// 		break
// 	case err := <-fatalErrors:
// 		close(fatalErrors)
// 		log.Fatal("Error: ", err)
// 	}
// }
