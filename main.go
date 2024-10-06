package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/oauth2"
)

func createGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func createLineBotClient(secret, token string) (*linebot.Client, error) {
	return linebot.New(secret, token)
}

func getGithubEvents(ctx context.Context, client *github.Client, username, date string) ([]*github.Event, error) {
	opts := &github.ListOptions{}
	var allEvents []*github.Event

	for {
		events, resp, err := client.Activity.ListEventsPerformedByUser(ctx, username, false, opts)
		if err != nil {
			return nil, err
		}
		for _, event := range events {
			eventDate := event.GetCreatedAt().Format("2006-01-02")
			if eventDate == date {
				allEvents = append(allEvents, event)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allEvents, nil
}

func buildMessage(events []*github.Event, isFinalCheck bool) string {
	today := time.Now()
	todayString := today.Format("2006-01-02")
	if len(events) == 0 {
		if isFinalCheck {
			return todayString + "の草が生えませんでした"
		}
		return todayString + "の草が生えていません"
	}
	message := todayString + "の草www:\n"

	uniqueEvents := make(map[string]bool)
	for _, event := range events {
		repoName := event.GetRepo().GetName()
		eventType := event.GetType()
		key := repoName + ":" + eventType
		if !uniqueEvents[key] {
			uniqueEvents[key] = true
			message += "\nリポジトリ: " + repoName + " (" + eventType + ")"
		}
	}

	return message
}

func sendLineMessage(bot *linebot.Client, userID, message string) error {
	textMessage := linebot.NewTextMessage(message)

	stickerMessage := linebot.NewStickerMessage("11537", "52002735") // 例: LINEの人気スタンプを指定

	_, err := bot.PushMessage(userID, textMessage, stickerMessage).Do()
	return err
}

func main() {
	ctx := context.Background()

	githubToken := os.Getenv("GH_TOKEN")
	githubUsername := os.Getenv("GH_USERNAME")
	lineChannelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	lineChannelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	lineUserID := os.Getenv("LINE_USER_ID")
	isFinalCheck := os.Getenv("IS_FINAL_CHECK") == "true"

	client := createGithubClient(ctx, githubToken)

	bot, err := createLineBotClient(lineChannelSecret, lineChannelToken)
	if err != nil {
		log.Fatalf("LINEボットクライアントの作成中にエラーが発生しました: %v", err)
	}

	today := time.Now()
	if isFinalCheck {
		today = today.AddDate(0, 0, -1)
	}
	dateString := today.Format("2006-01-02")

	events, err := getGithubEvents(ctx, client, githubUsername, dateString)
	if err != nil {
		log.Fatalf("イベントの取得中にエラーが発生しました: %v", err)
	}

	message := buildMessage(events, isFinalCheck)

	if err := sendLineMessage(bot, lineUserID, message); err != nil {
		log.Fatalf("LINEへのメッセージ送信中にエラーが発生しました: %v", err)
	}
}
