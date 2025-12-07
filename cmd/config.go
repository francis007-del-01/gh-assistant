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

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

