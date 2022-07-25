package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xanzy/go-gitlab"
)

const projectId = ""
const tokenBot = ""
const tokenGit = ""
const chatId = 

func main() {
	git, err := gitlab.NewClient(tokenGit)
	if err != nil {
		log.Fatalf("Failed to connect gitlab: %v", err)
	}
	bot, err := tgbotapi.NewBotAPI(tokenBot)
	if err != nil {
		log.Fatalf("Failed to connect telegram: %v", err)
	}
	bot.Debug = true
	id := 1
	project, statusCode, err := git.Projects.GetProject(projectId, nil)
	if err != nil {
		log.Fatalf("Git.Projects.GetProject: %v", err)
	}
	if statusCode.StatusCode != 200 || project == nil {
		log.Fatalf("Project not found: %v", err)
	}
	go getData(bot, git, &id, project)
	c := make(chan int)
	t := <-c
	fmt.Println(t)
}

func getData(bot *tgbotapi.BotAPI, git *gitlab.Client, lastId *int, project *gitlab.Project) {
	if bot == nil || git == nil || project == nil {
		return
	}
	for {
		events, _, _ := git.Events.ListProjectVisibleEvents(projectId, nil)
		if events != nil {
			event := events[0]
			data := "Notification"
			if *lastId != event.ID {
				*lastId = event.ID
				if strings.HasPrefix(event.ActionName, "pushed") {
					branch, statusCode, err := git.Branches.GetBranch(projectId, event.PushData.Ref)
					if err != nil {
						log.Printf("git.Branches.GetBranch: %v", err)
						return
					}
					if statusCode.StatusCode != 200 {
						log.Printf("Branch not found: %v", err)
						return
					}
					commit, statusCode, err := git.Commits.GetCommit(projectId, event.PushData.CommitTo)
					if err != nil {
						log.Printf("git.Commits.GetCommit: %v", err)
						return
					}
					if statusCode.StatusCode != 200 {
						log.Printf("Commit not found: %v", err)
						return
					}
					fmt.Println(commit)
					if branch != nil && commit != nil {
						data = fmt.Sprintf("*%s* [%s](%s) to [%s/%s/%s](%s) \n *%s* [%s](%s) [%s](%s) ``` @%s \n Additions:%d, Deletions:%d, Total:%d```",
							event.Author.Name, "pushed", branch.Commit.WebURL, event.Author.Username, project.Name, event.PushData.Ref, branch.WebURL,
							branch.Commit.AuthorName, branch.Commit.Message, branch.Commit.WebURL,
							"Commit", branch.Commit.WebURL, commit.ShortID,
							commit.Stats.Additions, commit.Stats.Deletions, commit.Stats.Total)
					}
				} else {
					data = fmt.Sprintf("*%s* %s", event.ActionName, event.PushData.Ref)
				}
				if bot != nil {
					msg := tgbotapi.NewMessage(chatId, data)
					msg.ParseMode = "markdown"
					if _, err := bot.Send(msg); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
