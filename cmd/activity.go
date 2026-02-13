package cmd

import (
	"fmt"
	"strings"

	"github.com/marcus/td/internal/db"
	"github.com/marcus/td/internal/output"
	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use:     "activity [issue-id]",
	Aliases: []string{"act"},
	Short:   "Show agent activity for an issue or agent",
	Long: `Show the agent activity log for an issue or a specific agent.

Examples:
  td activity td-abc123          # Activity for issue abc123
  td activity --agent oracle     # All activity by oracle agent
  td activity --agent smith -n 5 # Last 5 activities by smith`,
	GroupID: "core",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()

		database, err := db.Open(baseDir)
		if err != nil {
			output.Error("%v", err)
			return err
		}
		defer database.Close()

		limit, _ := cmd.Flags().GetInt("limit")
		agentName, _ := cmd.Flags().GetString("agent")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if len(args) == 0 && agentName == "" {
			output.Error("provide an issue ID or --agent <name>")
			return fmt.Errorf("provide an issue ID or --agent <name>")
		}

		if agentName != "" {
			// Show activity for a specific agent
			activities, err := database.ListAgentActivityByAgent(agentName, limit)
			if err != nil {
				output.Error("failed to get activity: %v", err)
				return err
			}

			if jsonOutput {
				return output.JSON(activities)
			}

			if len(activities) == 0 {
				fmt.Printf("No activity recorded for agent %q\n", agentName)
				return nil
			}

			fmt.Printf("Agent activity: %s (%d entries)\n", agentName, len(activities))
			fmt.Println(strings.Repeat("─", 60))
			for _, a := range activities {
				ts := a.CreatedAt.Format("2006-01-02 15:04")
				detail := ""
				if a.Details != "" {
					detail = " — " + a.Details
				}
				branch := ""
				if a.Branch != "" {
					branch = fmt.Sprintf(" [%s]", a.Branch)
				}
				fmt.Printf("  %s  %-14s  %s%s%s\n", ts, a.ActivityType, a.IssueID, branch, detail)
			}
			return nil
		}

		// Show activity for a specific issue
		issueID := args[0]
		issue, err := database.GetIssue(issueID)
		if err != nil {
			output.Error("%v", err)
			return err
		}

		activities, err := database.ListAgentActivity(issue.ID, limit)
		if err != nil {
			output.Error("failed to get activity: %v", err)
			return err
		}

		if jsonOutput {
			return output.JSON(activities)
		}

		assigneeInfo := ""
		if issue.Assignee != "" {
			assigneeInfo = fmt.Sprintf("  assignee: %s", issue.Assignee)
		}
		linearInfo := ""
		if issue.LinearIdentifier != "" {
			linearInfo = fmt.Sprintf("  linear: %s", issue.LinearIdentifier)
		}

		fmt.Printf("%s  %s%s%s\n", issue.ID, issue.Title, assigneeInfo, linearInfo)
		fmt.Println(strings.Repeat("─", 60))

		if len(activities) == 0 {
			fmt.Println("  No agent activity recorded")
			return nil
		}

		for _, a := range activities {
			ts := a.CreatedAt.Format("2006-01-02 15:04")
			detail := ""
			if a.Details != "" {
				detail = " — " + a.Details
			}
			branch := ""
			if a.Branch != "" {
				branch = fmt.Sprintf(" [%s]", a.Branch)
			}
			agent := a.AgentName
			fmt.Printf("  %s  %-12s  %-14s%s%s\n", ts, agent, a.ActivityType, branch, detail)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(activityCmd)
	activityCmd.Flags().IntP("limit", "n", 20, "Limit results")
	activityCmd.Flags().String("agent", "", "Show activity for a specific agent")
	activityCmd.Flags().Bool("json", false, "JSON output")
}
