package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StdBots/StdRankingBot/internal/bot"
	"github.com/StdBots/StdRankingBot/internal/config"
	"github.com/StdBots/StdRankingBot/internal/credit"
	"github.com/StdBots/StdRankingBot/internal/database"
)

const Banner = `
  ____ _____ ____    ____      _    _   _ _  _____ _   _  ____ 
 / ___|_   _|  _ \  |  _ \    / \  | \ | | |/ /_ _| \ | |/ ___|
 \___ \ | | | | | | | |_) |  / _ \ |  \| | ' / | ||  \| | |  _ 
  ___) || | | |_| | |  _ <  / ___ \| |\  | . \ | || |\  | |_| |
 |____/ |_| |____/  |_| \_\/_/   \_\_| \_|_|\_\___|_| \_|\____|
                                                               
   Enterprise Telegram Gamification, XP & Ranking Engine
   Developer: STD DEEPANSHU (https://deepanshu.in) | @STDBOTS
`

func main() {
	fmt.Println(Banner)

	// Step 1: Security & Credit Integrity Verification
	ok, issues := credit.VerifyIntegrity()
	if !ok {
		log.Fatalf("[FATAL] Tamper detected in core credit modules: %v", issues)
	}

	// Step 2: Load runtime configuration
	cfg := config.Load()

	// Step 3: Connect to MongoDB
	db, err := database.Connect(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("[FATAL] Could not connect to MongoDB: %v", err)
	}
	defer db.Disconnect()

	// Step 4: Initialize bot instance
	botInstance, err := bot.New(cfg, db)
	if err != nil {
		log.Fatalf("[FATAL] Bot initialization failed: %v", err)
	}

	// Step 5: Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n[SYSTEM] Received shutdown signal. Gracefully stopping StdRankingBot...")
		db.Disconnect()
		os.Exit(0)
	}()

	// Step 6: Start listening for messages and interactions
	botInstance.Start()
}
