package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

type Config struct {
    Token string `json:"token"`
    Admin string `json:"admin"`
}

func main() {
    var config Config

    if _, err := os.Stat("config.json"); err == nil {
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("⚙️ Config found. Reuse config? [y/n]: ")
        reuse, _ := reader.ReadString('\n')
        reuse = strings.TrimSpace(reuse)
        if strings.ToLower(reuse) == "y" {
            data, err := os.ReadFile("config.json")
            if err == nil {
                json.Unmarshal(data, &config)
            }
        }
    }

    if config.Token == "" || config.Admin == "" {
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("📲 Enter your Telegram Bot Token: ")
        token, _ := reader.ReadString('\n')
        config.Token = strings.TrimSpace(token)

        fmt.Print("🧑 Enter your Admin Telegram ID: ")
        admin, _ := reader.ReadString('\n')
        config.Admin = strings.TrimSpace(admin)

        data, err := json.MarshalIndent(config, "", "  ")
        if err == nil {
            os.WriteFile("config.json", data, 0644)
            fmt.Println("💾 Config saved successfully.")
        }
    }

    fmt.Println("\n🚀 Starting Obfuscator Bot...")
    fmt.Println("📌 Bot Token:", maskToken(config.Token))
    fmt.Println("👤 Admin ID :", config.Admin)
    fmt.Println("✅ Ready to run the bot!")

    // Run the python bot script
    cmd := exec.Command("python3", "bot.py")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        fmt.Println("❌ Failed to start Python bot:", err)
    }
}

func maskToken(token string) string {
    if len(token) < 10 {
        return token
    }
    return token[:5] + "..." + token[len(token)-5:]
}
