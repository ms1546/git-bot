package git_bot

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/oauth2"
)

type GitHubClient interface {
	ListEventsPerformedByUser(ctx context.Context, username string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error)
}

type GitHubClientWrapper struct {
	client *github.Client
}

func (w *GitHubClientWrapper) ListEventsPerformedByUser(ctx context.Context, username string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error) {
	return w.client.Activity.ListEventsPerformedByUser(ctx, username, publicOnly, opt)
}

func createGithubClient(ctx context.Context, token string) GitHubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return &GitHubClientWrapper{client: github.NewClient(tc)}
}

type LineBotClient interface {
	PushMessage(to string, messages ...linebot.SendingMessage) (*linebot.BasicResponse, error)
}

type LineBotClientWrapper struct {
	client *linebot.Client
}

func (w *LineBotClientWrapper) PushMessage(to string, messages ...linebot.SendingMessage) (*linebot.BasicResponse, error) {
	return w.client.PushMessage(to, messages...).Do()
}

func createLineBotClient(secret, token string) (LineBotClient, error) {
	client, err := linebot.New(secret, token)
	if err != nil {
		return nil, err
	}
	return &LineBotClientWrapper{client: client}, nil
}

func getGithubEvents(ctx context.Context, client GitHubClient, username, date string) ([]*github.Event, error) {
	opts := &github.ListOptions{}
	var allEvents []*github.Event

	for retry := 0; retry < 3; retry++ {
		events, resp, err := client.ListEventsPerformedByUser(ctx, username, false, opts)
		if err != nil {
			log.Printf("Retrying due to error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, event := range events {
			eventDate := event.GetCreatedAt().In(time.Local).Format("2006-01-02")
			if eventDate == date {
				allEvents = append(allEvents, event)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(allEvents) == 0 {
		return nil, nil
	}
	return allEvents, nil
}

func sendLineMessage(bot LineBotClient, userID, message string, hasEvents bool) error {
	textMessage := linebot.NewTextMessage(message)
	if hasEvents {
		stickerMessage := linebot.NewStickerMessage("11537", "52002735")
		_, err := bot.PushMessage(userID, textMessage, stickerMessage)
		return err
	}
	_, err := bot.PushMessage(userID, textMessage)
	return err
}

func sendErrorMessage(bot LineBotClient, userID, errMsg string) {
	textMessage := linebot.NewTextMessage("Error: " + errMsg)
	if _, err := bot.PushMessage(userID, textMessage); err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

func buildMessage(events []*github.Event, isFinalCheck bool) string {
	todayString := time.Now().Format("2006-01-02")
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
		log.Printf("Error creating LINE bot client: %v", err)
		sendErrorMessage(bot, lineUserID, err.Error())
		return
	}

	today := time.Now()
	if isFinalCheck {
		today = today.AddDate(0, 0, -1)
	}
	dateString := today.Format("2006-01-02")

	events, err := getGithubEvents(ctx, client, githubUsername, dateString)
	if err != nil {
		log.Printf("Error fetching events: %v", err)
		sendErrorMessage(bot, lineUserID, err.Error())
		return
	}

	message := buildMessage(events, isFinalCheck)
	hasEvents := len(events) > 0

	if err := sendLineMessage(bot, lineUserID, message, hasEvents); err != nil {
		log.Printf("Error sending LINE message: %v", err)
		sendErrorMessage(bot, lineUserID, err.Error())
	}
}
