package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BotToken string `json:"bot_token"`
	AdminID  string `json:"admin_id"`
}

const configFile = "config.json"

func main() {
	fmt.Println("🔐 Obfuscator Bot Setup")

	config := Config{}

	// Check if config file exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Print("⚙️ Config found. Reuse config? [y/n]: ")
		choice := readLine()
		if strings.ToLower(choice) == "y" {
			data, err := os.ReadFile(configFile)
			if err != nil {
				fmt.Println("❌ Failed to read config:", err)
				return
			}
			err = json.Unmarshal(data, &config)
			if err != nil {
				fmt.Println("❌ Failed to parse config:", err)
				return
			}
			fmt.Println("✅ Using saved config.")
		} else {
			config = askCredentials()
			saveConfig(config)
		}
	} else {
		config = askCredentials()
		saveConfig(config)
	}

	fmt.Println("\n🚀 Starting Obfuscator Bot...")
	fmt.Println("📌 Bot Token:", maskToken(config.BotToken))
	fmt.Println("👤 Admin ID :", config.AdminID)
	// You can call your real Python bot script or compiled obfuscator here
	// For now:
	fmt.Println("✅ Ready to run the bot!")
}

// Ask for user input
func askCredentials() Config {
	fmt.Print("📲 Enter your Telegram Bot Token: ")
	botToken := readLine()

	fmt.Print("🧑 Enter your Admin Telegram ID: ")
	adminID := readLine()

	return Config{BotToken: botToken, AdminID: adminID}
}

// Save to config file
func saveConfig(cfg Config) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Println("❌ Error saving config:", err)
		return
	}
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		fmt.Println("❌ Cannot write config file:", err)
	} else {
		fmt.Println("💾 Config saved successfully.")
	}
}

// Read line input
func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// Mask token for display
func maskToken(token string) string {
	if len(token) < 10 {
		return token
	}
	return token[:5] + "..." + token[len(token)-5:]
}
