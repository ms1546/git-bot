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

func getGithubEvents(ctx context.Context, client *github.Client, username string) ([]*github.Event, error) {
	events, _, err := client.Activity.ListEventsPerformedByUser(ctx, username, false, nil)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func sendLineMessage(bot *linebot.Client, userID, message string) error {
	if _, err := bot.PushMessage(userID, linebot.NewTextMessage(message)).Do(); err != nil {
		return err
	}
	return nil
}

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

	events, err := getGithubEvents(ctx, client, username)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}
	if len(events) == 0 {
		log.Println("No events fetched for today.")
	}

	var grassExists bool
	var latestEvent *github.Event
	for _, event := range events {
		eventDate := event.GetCreatedAt().Format("2006-01-02")
		if eventDate == today {
			grassExists = true
			latestEvent = event
			break
		}
	}

	if !grassExists {
		message := "今日はまだGitHubに草が生えていませんw"
		if err := sendLineMessage(bot, os.Getenv("LINE_USER_ID"), message); err != nil {
			log.Fatalf("Error sending message to LINE: %v", err)
		}
	} else if latestEvent != nil {
		message := "更新がありました！\n"
		message += "イベントの種類: " + latestEvent.GetType() + "\n"
		message += "リポジトリ: " + latestEvent.GetRepo().GetName() + "\n"

		switch latestEvent.GetType() {
		case "PushEvent":
			pushEventPayload, payloadErr := latestEvent.ParsePayload()
			if payloadErr != nil {
				log.Fatalf("Error parsing payload: %v", payloadErr)
			}
			pushEvent, ok := pushEventPayload.(*github.PushEvent)
			if !ok {
				log.Fatalf("Error casting to push event")
			}
			message += "詳細: " + pushEvent.GetHead()
		default:
			message += "詳細: イベントの詳細は対応していません。"
		}

		if err := sendLineMessage(bot, os.Getenv("LINE_USER_ID"), message); err != nil {
			log.Fatalf("Error sending message to LINE: %v", err)
		}
	}

	if time.Now().Hour() == 23 && time.Now().Minute() >= 59 {
		finalMessage := "今日のGitHubのコントリビューションはありませんでした"
		if grassExists {
			finalMessage = "今日のGitHubのコントリビューションがありましたwww"
		}
		if err := sendLineMessage(bot, os.Getenv("LINE_USER_ID"), finalMessage); err != nil {
			log.Fatalf("Error sending message to LINE: %v", err)
		}
	}
}
