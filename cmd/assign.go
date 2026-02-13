package cmd

import (
	"fmt"

	"github.com/marcus/td/internal/db"
	"github.com/marcus/td/internal/output"
	"github.com/marcus/td/internal/session"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:     "assign <issue-id> <agent-name>",
	Aliases: []string{"delegate"},
	Short:   "Assign an issue to an agent",
	Long: `Assign an issue to an agent (e.g., oracle, smith, executor-1).
This records the assignment and tracks it in the agent activity log.

Use an empty string to unassign: td assign <id> ""`,
	GroupID: "core",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()

		database, err := db.Open(baseDir)
		if err != nil {
			output.Error("%v", err)
			return err
		}
		defer database.Close()

		sess, err := session.GetOrCreate(database)
		if err != nil {
			output.Error("%v", err)
			return err
		}

		issueID := args[0]
		agentName := args[1]

		// Verify issue exists
		issue, err := database.GetIssue(issueID)
		if err != nil {
			output.Error("%v", err)
			return err
		}

		if err := database.AssignIssue(issue.ID, agentName, sess.ID); err != nil {
			output.Error("failed to assign: %v", err)
			return err
		}

		if agentName == "" {
			fmt.Printf("UNASSIGNED %s\n", issue.ID)
		} else {
			fmt.Printf("ASSIGNED %s -> %s\n", issue.ID, agentName)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(assignCmd)
}
