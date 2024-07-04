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

func getGithubEvents(ctx context.Context, client *github.Client, username string, date string) ([]*github.Event, error) {
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

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	events, err := getGithubEvents(ctx, client, username, yesterday)
	if err != nil {
		log.Fatalf("Error fetching events: %v", err)
	}

	var message string
	if len(events) == 0 {
		message = "昨日はGitHubに草が生えていませんでした。"
	} else {
		message = "昨日のGitHubイベント:\n"
		for _, event := range events {
			message += "イベントの種類: " + event.GetType() + "\n"
			message += "リポジトリ: " + event.GetRepo().GetName() + "\n"
			switch event.GetType() {
			case "PushEvent":
				pushEventPayload, payloadErr := event.ParsePayload()
				if payloadErr != nil {
					log.Fatalf("Error parsing payload: %v", payloadErr)
				}
				pushEvent, ok := pushEventPayload.(*github.PushEvent)
				if !ok {
					log.Fatalf("Error casting to push event")
				}
				message += "詳細: " + pushEvent.GetHead() + "\n"
			default:
				message += "詳細: イベントの詳細は対応していません。\n"
			}
		}
	}

	if err := sendLineMessage(bot, os.Getenv("LINE_USER_ID"), message); err != nil {
		log.Fatalf("Error sending message to LINE: %v", err)
	}
}
