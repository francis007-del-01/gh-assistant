package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/namin2/gh-assistant/internal/ai"
	"github.com/namin2/gh-assistant/internal/git"
	"github.com/namin2/gh-assistant/internal/jira"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	autoConfirm bool
	stageAll    bool
)

var pushxCmd = &cobra.Command{
	Use:   "pushx",
	Short: "Generate AI commit message and push",
	Long: `Analyzes your git diff, generates a meaningful commit message using AI,
and pushes to the remote repository.

Examples:
  gh-assistant pushx           # Commit staged changes with AI message and push
  gh-assistant pushx -a        # Stage all changes, commit with AI message and push
  gh-assistant pushx -y        # Skip confirmation prompt`,
	RunE: runPushx,
}

func init() {
	rootCmd.AddCommand(pushxCmd)
	pushxCmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "Auto-confirm the generated commit message")
	pushxCmd.Flags().BoolVarP(&stageAll, "all", "a", false, "Stage all changes before committing")
}

func runPushx(cmd *cobra.Command, args []string) error {
	// Check configuration
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	if apiKey == "" {
		return fmt.Errorf(`API key not configured. Set it up using one of:
  1. Run: gh-assistant config --api-key YOUR_KEY
  2. Set environment variable: export OPENAI_API_KEY=your_key
  3. Set environment variable: export ANTHROPIC_API_KEY=your_key`)
	}

	// Determine provider
	provider := ai.Provider(viper.GetString("provider"))
	if provider == "" {
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			provider = ai.ProviderAnthropic
		} else {
			provider = ai.ProviderOpenAI
		}
	}

	// Initialize git
	g := git.New("")

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	fmt.Println("🔍 Analyzing your changes...")

	// Stage all if requested
	if stageAll {
		fmt.Println("📦 Staging all changes...")
		if err := g.StageAll(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}
	}

	// Check for staged changes
	hasStaged, err := g.HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	var diff string
	var changedFiles []string
	var needsCommit bool

	if hasStaged {
		// We have staged changes that need to be committed
		diff, err = g.GetStagedDiff()
		if err != nil {
			return fmt.Errorf("failed to get staged diff: %w", err)
		}
		needsCommit = true
		fmt.Println("📝 Found staged changes to commit")
	} else {
		// Check for unpushed commits
		diff, err = g.GetUnpushedDiff()
		if err != nil {
			// Might be first push
			diff, err = g.GetAllDiff()
			if err != nil {
				return fmt.Errorf("failed to get diff: %w", err)
			}
		}

		if diff == "" {
			// Check if there are unstaged changes
			hasUnstaged, _ := g.HasUnstagedChanges()
			if hasUnstaged {
				return fmt.Errorf("you have unstaged changes. Use -a flag to stage all, or stage manually with 'git add'")
			}
			return fmt.Errorf("no changes to commit or push")
		}
		fmt.Println("📝 Found unpushed commits")
	}

	changedFiles, _ = g.GetChangedFiles()

	if diff == "" {
		return fmt.Errorf("no changes detected")
	}

	// Initialize AI client
	aiClient := ai.New(ai.Config{
		Provider: provider,
		APIKey:   apiKey,
		Model:    viper.GetString("model"),
	})

	fmt.Println("🤖 Generating commit message...")

	// Generate commit message
	message, err := aiClient.GenerateCommitMessage(diff, changedFiles)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Display the generated message
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📋 Generated commit message:")
	fmt.Println()
	fmt.Printf("   %s\n", message)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Confirm with user
	if !autoConfirm {
		fmt.Print("Proceed with this message? [Y/n/e(dit)]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "n", "no":
			fmt.Println("❌ Aborted")
			return nil
		case "e", "edit":
			fmt.Println("Enter your commit message (press Enter twice to finish):")
			var lines []string
			for {
				line, _ := reader.ReadString('\n')
				line = strings.TrimRight(line, "\n\r")
				if line == "" && len(lines) > 0 {
					break
				}
				if line != "" {
					lines = append(lines, line)
				}
			}
			if len(lines) > 0 {
				message = strings.Join(lines, "\n")
			}
		case "", "y", "yes":
			// Continue with the message
		default:
			fmt.Println("❌ Invalid input, aborted")
			return nil
		}
	}

	// Commit if we have staged changes
	if needsCommit {
		fmt.Println("💾 Creating commit...")
		if err := g.Commit(message); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
		fmt.Printf("✅ Committed: %s\n", message)
	}

	// Check if this is a first push to a new branch (for Jira creation)
	isFirstPush, _ := g.IsFirstPushToBranch()
	isMainBranch := g.IsMainBranch()

	// Push
	fmt.Println("🚀 Pushing to remote...")
	err = g.Push()
	if err != nil {
		// Try with set-upstream
		err = g.PushSetUpstream()
		if err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
	}

	fmt.Println("✅ Successfully pushed!")

	// Create Jira ticket on first push to a new branch (not main/master)
	if isFirstPush && !isMainBranch {
		jiraClient := jira.New(jira.Config{
			BaseURL:  viper.GetString("jira_url"),
			Email:    viper.GetString("jira_email"),
			APIToken: viper.GetString("jira_token"),
			Project:  viper.GetString("jira_project"),
		})

		if jiraClient.IsConfigured() {
			fmt.Println()
			fmt.Println("🎫 Creating Jira ticket...")

			title, err := jiraClient.CreateIssueWithTitle(message)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to create Jira ticket: %v\n", err)
			} else {
				// Extract issue key from title (format: "KEY-123 - message")
				parts := strings.SplitN(title, " - ", 2)
				issueKey := parts[0]
				fmt.Printf("✅ Jira ticket created: %s\n", title)
				fmt.Printf("🔗 %s\n", jiraClient.GetIssueURL(issueKey))
			}
		}
	}

	return nil
}

