package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/namin2/gh-assistant/internal/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	apiKey      string
	providerArg string
	modelArg    string
	// Jira config flags
	jiraURL     string
	jiraEmail   string
	jiraToken   string
	jiraProject string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure gh-assistant settings",
	Long: `Configure API keys and other settings for gh-assistant.

Examples:
  gh-assistant config --api-key sk-xxx --provider openai
  gh-assistant config --api-key sk-ant-xxx --provider anthropic
  gh-assistant config --model gpt-4o
  gh-assistant config --show`,
	RunE: runConfig,
}

var showConfig bool

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringVar(&apiKey, "api-key", "", "Set the API key")
	configCmd.Flags().StringVar(&providerArg, "provider", "", "Set the AI provider (openai, anthropic)")
	configCmd.Flags().StringVar(&modelArg, "model", "", "Set the model to use")
	configCmd.Flags().BoolVar(&showConfig, "show", false, "Show current configuration")
	// Jira configuration flags
	configCmd.Flags().StringVar(&jiraURL, "jira-url", "", "Set Jira base URL (e.g., https://yourcompany.atlassian.net)")
	configCmd.Flags().StringVar(&jiraEmail, "jira-email", "", "Set Jira account email")
	configCmd.Flags().StringVar(&jiraToken, "jira-token", "", "Set Jira API token")
	configCmd.Flags().StringVar(&jiraProject, "jira-project", "", "Set Jira project key (e.g., PROJ)")
}

func runConfig(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".gh-assistant.yaml")

	// Show current config
	if showConfig {
		return showCurrentConfig()
	}

	// Load existing config
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		yaml.Unmarshal(data, &config)
	}

	// Update config
	updated := false

	if apiKey != "" {
		config["api_key"] = apiKey
		updated = true
		fmt.Println("✅ API key configured")
	}

	if providerArg != "" {
		p := ai.Provider(providerArg)
		if p != ai.ProviderOpenAI && p != ai.ProviderAnthropic {
			return fmt.Errorf("invalid provider: %s (use 'openai' or 'anthropic')", providerArg)
		}
		config["provider"] = providerArg
		updated = true
		fmt.Printf("✅ Provider set to: %s\n", providerArg)
	}

	if modelArg != "" {
		config["model"] = modelArg
		updated = true
		fmt.Printf("✅ Model set to: %s\n", modelArg)
	}

	// Jira configuration
	if jiraURL != "" {
		config["jira_url"] = jiraURL
		updated = true
		fmt.Printf("✅ Jira URL set to: %s\n", jiraURL)
	}

	if jiraEmail != "" {
		config["jira_email"] = jiraEmail
		updated = true
		fmt.Printf("✅ Jira email set to: %s\n", jiraEmail)
	}

	if jiraToken != "" {
		config["jira_token"] = jiraToken
		updated = true
		fmt.Println("✅ Jira API token configured")
	}

	if jiraProject != "" {
		config["jira_project"] = jiraProject
		updated = true
		fmt.Printf("✅ Jira project set to: %s\n", jiraProject)
	}

	if !updated {
		cmd.Help()
		return nil
	}

	// Save config
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n📁 Configuration saved to: %s\n", configPath)
	return nil
}

func showCurrentConfig() error {
	fmt.Println("Current configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Check file config
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gh-assistant.yaml")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("📁 Config file: %s\n", configPath)
	} else {
		fmt.Println("📁 Config file: not created")
	}

	fmt.Println()

	// Provider
	provider := viper.GetString("provider")
	if provider == "" {
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			provider = "anthropic (from env)"
		} else if os.Getenv("OPENAI_API_KEY") != "" {
			provider = "openai (from env)"
		} else {
			provider = "not set"
		}
	}
	fmt.Printf("🤖 Provider: %s\n", provider)

	// API Key
	key := viper.GetString("api_key")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
		if key == "" {
			key = os.Getenv("ANTHROPIC_API_KEY")
		}
	}
	if key != "" {
		// Mask the key
		if len(key) > 8 {
			key = key[:4] + "..." + key[len(key)-4:]
		} else {
			key = "****"
		}
		fmt.Printf("🔑 API Key: %s\n", key)
	} else {
		fmt.Println("🔑 API Key: not set")
	}

	// Model
	model := viper.GetString("model")
	if model == "" {
		model = "default"
	}
	fmt.Printf("📦 Model: %s\n", model)

	fmt.Println()
	fmt.Println("Jira Integration:")

	// Jira URL
	jURL := viper.GetString("jira_url")
	if jURL != "" {
		fmt.Printf("🔗 Jira URL: %s\n", jURL)
	} else {
		fmt.Println("🔗 Jira URL: not set")
	}

	// Jira Email
	jEmail := viper.GetString("jira_email")
	if jEmail != "" {
		fmt.Printf("📧 Jira Email: %s\n", jEmail)
	} else {
		fmt.Println("📧 Jira Email: not set")
	}

	// Jira Token
	jToken := viper.GetString("jira_token")
	if jToken != "" {
		if len(jToken) > 8 {
			jToken = jToken[:4] + "..." + jToken[len(jToken)-4:]
		} else {
			jToken = "****"
		}
		fmt.Printf("🔑 Jira Token: %s\n", jToken)
	} else {
		fmt.Println("🔑 Jira Token: not set")
	}

	// Jira Project
	jProject := viper.GetString("jira_project")
	if jProject != "" {
		fmt.Printf("📋 Jira Project: %s\n", jProject)
	} else {
		fmt.Println("📋 Jira Project: not set")
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

