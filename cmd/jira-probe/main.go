package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/predicta/predicta/config"
	"github.com/predicta/predicta/internal/infrastructure/jira"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.ValidateJira(); err != nil {
		log.Fatalf("config: %v", err)
	}

	client := jira.NewClient(
		cfg.JiraBaseURL,
		cfg.JiraEmail,
		cfg.JiraAPIToken,
		cfg.JiraSprintID,
		cfg.JiraProjectKey,
		cfg.SprintDaysRemaining,
	)
	ctx := context.Background()

	sprint, err := client.GetSprintInfo(ctx)
	if err != nil {
		log.Fatalf("sprint: %v", err)
	}
	fmt.Printf("sprint: %s (%d) state=%s\n", sprint.Name, sprint.ID, sprint.State)

	tasks, err := client.GetSprintTasks(ctx, "")
	if err != nil {
		log.Fatalf("tasks: %v", err)
	}
	fmt.Printf("tasks in sprint: %d\n", len(tasks))
	for i, t := range tasks {
		if i >= 5 {
			fmt.Printf("... and %d more\n", len(tasks)-5)
			break
		}
		fmt.Printf("  %s | %s | assignee=%s | %s\n", t.ID, t.Title, t.AssigneeID, t.Status)
	}

	users, err := client.ListAssignableUsers(ctx, cfg.JiraProjectKey)
	if err != nil {
		log.Fatalf("users: %v", err)
	}
	fmt.Printf("assignable users: %d\n", len(users))
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(users)
}
