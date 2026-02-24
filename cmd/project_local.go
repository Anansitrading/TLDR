package cmd

import (
	"fmt"

	"github.com/Anansitrading/TLDR/internal/db"
	"github.com/Anansitrading/TLDR/internal/output"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Manage projects",
	Long:    "Create, list, delete, and link projects for organizing issues.",
	GroupID: "core",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()
		database, err := db.Open(baseDir)
		if err != nil {
			return err
		}
		defer database.Close()

		name := args[0]
		desc, _ := cmd.Flags().GetString("description")
		claudeTeam, _ := cmd.Flags().GetString("claude-team")

		id, err := database.CreateProject(name, desc, claudeTeam)
		if err != nil {
			output.Error("failed to create project: %v", err)
			return err
		}
		fmt.Printf("CREATED project %s (%s)\n", name, id)
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()
		database, err := db.Open(baseDir)
		if err != nil {
			return err
		}
		defer database.Close()

		projects, err := database.ListProjects()
		if err != nil {
			return err
		}
		if len(projects) == 0 {
			fmt.Println("No projects found. Create one with: td project create \"name\"")
			return nil
		}
		for _, p := range projects {
			team := ""
			if p.ClaudeTeamName != "" {
				team = fmt.Sprintf(" [claude-team: %s]", p.ClaudeTeamName)
			}
			fmt.Printf("%-12s %s%s\n", p.ID, p.Name, team)
		}
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()
		database, err := db.Open(baseDir)
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.DeleteProject(args[0]); err != nil {
			output.Error("%v", err)
			return err
		}
		fmt.Printf("DELETED project %s\n", args[0])
		return nil
	},
}

var projectLinkCmd = &cobra.Command{
	Use:   "link [name]",
	Short: "Link a project to a Claude Code team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir := getBaseDir()
		database, err := db.Open(baseDir)
		if err != nil {
			return err
		}
		defer database.Close()

		claudeTeam, _ := cmd.Flags().GetString("claude-team")
		if claudeTeam == "" {
			return fmt.Errorf("--claude-team is required")
		}

		if err := database.LinkProjectToClaudeTeam(args[0], claudeTeam); err != nil {
			output.Error("%v", err)
			return err
		}
		fmt.Printf("LINKED project %s → claude team %s\n", args[0], claudeTeam)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectLinkCmd)

	projectCreateCmd.Flags().String("description", "", "Project description")
	projectCreateCmd.Flags().String("claude-team", "", "Link to Claude Code team name")
	projectLinkCmd.Flags().String("claude-team", "", "Claude Code team name to link")
}
