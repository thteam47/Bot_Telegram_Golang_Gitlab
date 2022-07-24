package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xanzy/go-gitlab"
)

func main() {
	git, err := gitlab.NewClient("token-git")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	bot, err := tgbotapi.NewBotAPI("token-bot")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	id := 1
	go getData(bot, git, &id)
	c := make(chan int)
	t := <-c
	fmt.Println(t)
}

func getData(bot *tgbotapi.BotAPI, git *gitlab.Client, lastId *int) {
	for {
		events, _, _ := git.Events.ListProjectVisibleEvents("project-id", nil)
		if events != nil {
			if *lastId != events[0].ID {
				*lastId = events[0].ID
				if bot != nil {
					data := strconv.Itoa(events[0].ID) + events[0].ActionName + events[0].TargetType
					msg := tgbotapi.NewMessage("id chat", data)
					if _, err := bot.Send(msg); err != nil {
					}
					time.Sleep(20 * time.Second)
				}
				fmt.Println(events[0])
			}
		}
	}
}

// func getProjectIssues(gitlab *gogitlab.Gitlab, projectId int) {

// 	events := gitlab.ProjectEvents(projectId)
// 	for _, event := range events {

// 		var iconName string
// 		switch event.TargetType {
// 		case "Issue":
// 			iconName = ":beer:"
// 		default:
// 			iconName = ":punch:"
// 		}

// 		fmt.Printf("ProjectID[%d] action[%s] targetId[%d] targetType[%s] targetTitle[%s]\n", event.ProductId, event.ActionName,event.TargetId, event.TargetType, event.TargetTitle)
// 		if event.TargetId != 0 {
// 			actionText := color.Sprintf("@y[%s]", event.ActionName)
// 			repositoriesText := color.Sprintf("@c%s(%d)", event.TargetType, event.TargetId)
// 			userText := color.Sprintf("@c%s", event.Data.UserName)
// 			titleText := color.Sprintf("@g%s", event.TargetTitle)
// 			emoji.Println("@{"+iconName+"}", actionText, repositoriesText, userText, titleText)

// 		} else if event.TargetId == 0 {

// 			actionText := color.Sprintf("@y[%s]", event.ActionName)
// 			repositoriesText := color.Sprintf("@c%s", event.Data.Repository.Name)
// 			userText := color.Sprintf("@c%s", event.Data.UserName)
// 			var titleText string
// 			if event.Data.TotalCommitsCount > 0 {
// 				commitMessage := event.Data.Commits[0].Message
// 				commitMessage = strings.Replace(commitMessage, "\n\n", "\t", -1)
// 				titleText = color.Sprintf("@g%s", commitMessage)
// 			} else if event.Data.Before == "0000000000000000000000000000000000000000" {
// 				titleText = color.Sprintf("@g%s %s", emoji.Sprint("@{:fire:}"), "create New branch")
// 			}
// 			emoji.Println("@{"+iconName+"}", actionText, repositoriesText, userText, titleText)

// 						fmt.Println(" \t user   -> ", event.Data.UserName, event.Data.UserId)
// 						fmt.Println(" \t author -> ", event.Data.AuthorId)

// 						fmt.Println(" \t\t name        -> ", event.Data.Repository.Name)
// 						fmt.Println(" \t\t description -> ", event.Data.Repository.Description)
// 						fmt.Println(" \t\t gitUrl      -> ", event.Data.Repository.GitUrl)
// 						fmt.Println(" \t\t pageUrl     -> ", event.Data.Repository.PageUrl)

// 						fmt.Println(" \t\t totalCount  -> ", event.Data.TotalCommitsCount)

// 						if event.Data.TotalCommitsCount > 0 {
// 							fmt.Println(" \t\t message     -> ", event.Data.Commits[0].Message)
// 							fmt.Println(" \t\t time        -> ", event.Data.Commits[0].Timestamp)
// 						}
// 		}
// 	}
// }
