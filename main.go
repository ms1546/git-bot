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

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	lineChannelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	lineChannelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	bot, err := linebot.New(lineChannelSecret, lineChannelToken)
	if err != nil {
		log.Fatalf("Error creating LINE bot client: %v", err)
	}

	username := os.Getenv("GH_USERNAME")

	today := time.Now().Format("2006-01-02")

	events, _, err := client.Activity.ListEventsPerformedByUser(ctx, username, false, nil)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	var grassExists bool
	for _, event := range events {
		if event.CreatedAt.Format("2006-01-02") == today {
			grassExists = true
			break
		}
	}

	if !grassExists {
		if !grassExists {
			message := "今日はまだGitHubに草が生えていませんw"
			if _, err := bot.PushMessage(os.Getenv("LINE_USER_ID"), linebot.NewTextMessage(message)).Do(); err != nil {
				log.Fatalf("Error sending message to LINE: %v", err)
			}
		}
	}

	if time.Now().Hour() == 23 && time.Now().Minute() >= 59 {
		finalMessage := "今日のGitHubのコントリビューションはありませんでした"
		if grassExists {
			finalMessage = "今日のGitHubのコントリビューションがありましたwww"
		}
		if _, err := bot.PushMessage(os.Getenv("LINE_USER_ID"), linebot.NewTextMessage(finalMessage)).Do(); err != nil {
			log.Fatalf("Error sending message to LINE: %v", err)
		}
	}
}
