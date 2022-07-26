package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xanzy/go-gitlab"
)

const projectId = "38057156"
const tokenBot = "5204140121:AAFky6KMUqdAUhvWVUPBWoOghqH4cH8lW4c"
const tokenGit = ""
const chatId = 5172255611

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
				fmt.Println(event)
				if strings.HasPrefix(event.ActionName, "pushed to") {
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
					if branch != nil && commit != nil {
						data = fmt.Sprintf("*%s* [%s](%s) to [%s/%s/%s](%s) \n *%s* [%s](%s) [%s](%s) ``` @%s \n Additions:%d, Deletions:%d, Total:%d```",
							event.Author.Name, "pushed", branch.Commit.WebURL, event.Author.Username, project.Name, event.PushData.Ref, branch.WebURL,
							branch.Commit.AuthorName, branch.Commit.Message, branch.Commit.WebURL,
							"Commit", branch.Commit.WebURL, commit.ShortID,
							commit.Stats.Additions, commit.Stats.Deletions, commit.Stats.Total)
					}
				} else if strings.HasPrefix(event.ActionName, "pushed new") {
					branch, statusCode, err := git.Branches.GetBranch(projectId, event.PushData.Ref)
					if err != nil {
						log.Printf("git.Branches.GetBranch: %v", err)
						return
					}
					if statusCode.StatusCode != 200 {
						log.Printf("Branch not found: %v", err)
						return
					}
					if branch != nil {
						data = fmt.Sprintf("*%s* %s %s [%s/%s](%s)",
							event.Author.Name, event.PushData.Action, event.PushData.RefType, project.Name, event.PushData.Ref, branch.WebURL)
					}
				} else if strings.HasPrefix(event.ActionName, "opened") {
					if event.TargetType == "Issue" {
						issue, statusCode, err := git.Issues.GetIssue(projectId, event.TargetIID)
						if err != nil {
							log.Printf("git.Issues.GetIssue: %v", err)
							return
						}
						if statusCode.StatusCode != 200 {
							log.Printf("Issue not found: %v", err)
							return
						}
						if issue != nil {
							data = fmt.Sprintf("*%s* %s [%s](%s) at [%s/%s](%s/%s): *%s*",
								event.Author.Name, issue.State, *issue.IssueType, issue.WebURL, issue.Author.Username, project.Name, issue.Author.WebURL, project.Name, event.TargetTitle)
						}
					} else if event.TargetType == "MergeRequest" {
						mergeRequest, statusCode, err := git.MergeRequests.GetMergeRequest(projectId, event.TargetIID, nil)
						if err != nil {
							log.Printf("git.MergeRequests.GetMergeRequest: %v", err)
							return
						}
						if statusCode.StatusCode != 200 {
							log.Printf("Merge Request not found: %v", err)
							return
						}
						if mergeRequest != nil {
							data = fmt.Sprintf("*%s* %s [%s](%s) at [/%s](%s/%s): *%s*",
								event.Author.Name, mergeRequest.State, "merge request", mergeRequest.WebURL, project.Name, mergeRequest.Author.WebURL, project.Name, event.TargetTitle)
						}
					}
				} else if strings.HasPrefix(event.ActionName, "commented") {
					mergeRequest, statusCode, err := git.MergeRequests.GetMergeRequest(projectId, event.Note.NoteableIID, nil)
					if err != nil {
						log.Printf("git.MergeRequests.GetMergeRequest: %v", err)
						return
					}
					if statusCode.StatusCode != 200 {
						log.Printf("Merge Request not found: %v", err)
						return
					}
					if mergeRequest != nil {
						data = fmt.Sprintf("*%s* %s [%s](%s#note_%d): ``` %s ```",
							event.Author.Name, event.ActionName, "merge request", mergeRequest.WebURL, event.TargetID, event.Note.Body)
					}
				} else {
					data = fmt.Sprintf("*%s* %s %d %s", event.Author.Name, event.ActionName, event.TargetID, event.TargetTitle)
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
