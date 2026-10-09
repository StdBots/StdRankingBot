package bot

import (
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/StdBots/StdRankingBot/internal/config"
	"github.com/StdBots/StdRankingBot/internal/credit"
	"github.com/StdBots/StdRankingBot/internal/database"
)

// Bot represents the StdRankingBot instance
type Bot struct {
	api       *tgbotapi.BotAPI
	db        *database.DB
	cfg       *config.Config
	cache     *database.AntiFarmCache
	startTime time.Time
}

// New creates a new Bot instance
func New(cfg *config.Config, db *database.DB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	if cfg.Env == "development" {
		api.Debug = true
	}

	log.Printf("[BOT] Authorized on account @%s (ID: %d)", api.Self.UserName, api.Self.ID)

	b := &Bot{
		api:       api,
		db:        db,
		cfg:       cfg,
		cache:     database.NewAntiFarmCache(),
		startTime: time.Now(),
	}

	// Set Telegram Bot Command List
	b.registerBotCommands()

	// Verify credit integrity
	ok, tampered := credit.VerifyIntegrity()
	if !ok {
		log.Printf("[WARN] Credit verification issues detected: %v", tampered)
	}

	// Telemetry ping
	credit.ReportForkStatus(api.Self.UserName, ok)

	return b, nil
}

func (b *Bot) registerBotCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "top", Description: "View chat leaderboard banner"},
		{Command: "ranking", Description: "View chat ranking with buttons"},
		{Command: "me", Description: "View your personal fighter profile card"},
		{Command: "global", Description: "Global cross-chat leaderboard"},
		{Command: "settings", Description: "Configure group ranking preferences"},
		{Command: "help", Description: "How XP, leveling & ranking works"},
	}

	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(cmdConfig)
	if err != nil {
		log.Printf("[WARN] Failed to set bot commands: %v", err)
	}
}

// Start initiates update polling and the periodic reset scheduler
func (b *Bot) Start() {
	b.startScheduler()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)
	log.Printf("[BOT] @%s listening for group interactions & rankings...", b.api.Self.UserName)

	for update := range updates {
		up := update
		go b.SafeHandle(func() {
			b.processUpdate(up)
		})
	}
}

func (b *Bot) processUpdate(update tgbotapi.Update) {
	// Handle Callback Queries (Inline buttons)
	if update.CallbackQuery != nil {
		b.HandleCallback(update.CallbackQuery)
		return
	}

	// Handle Messages
	if update.Message != nil {
		msg := update.Message

		// If message is in group or supergroup, ensure group exists and track activity
		if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, _ = b.db.EnsureGroup(ctx, msg.Chat.ID, msg.Chat.Title, msg.Chat.UserName, msg.Chat.Type)
			}()

			// If it's a command, execute handler
			if msg.IsCommand() {
				b.HandleCommand(msg)
				return
			}

			// Regular chat message: track and award XP
			b.TrackGroupMessage(msg)
			return
		}

		// Private Chat Messages
		if msg.Chat.IsPrivate() {
			if msg.IsCommand() {
				b.HandleCommand(msg)
			} else {
				b.HandlePrivateMessage(msg)
			}
		}
	}
}

// startScheduler handles daily, weekly, and monthly counter rollovers
func (b *Bot) startScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		var lastDailyDay = time.Now().UTC().Day()
		var lastWeeklyWeek, _ = time.Now().UTC().ISOWeek()
		var lastMonthlyMonth = time.Now().UTC().Month()

		for range ticker.C {
			now := time.Now().UTC()

			// Daily Reset check (at midnight UTC)
			if now.Day() != lastDailyDay {
				log.Println("[SCHEDULER] Triggering Daily Leaderboard Reset...")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_ = b.db.ResetPeriodCounters(ctx, "daily")
				cancel()
				lastDailyDay = now.Day()
			}

			// Weekly Reset check (Mondays)
			currentWeek, _ := now.ISOWeek()
			if currentWeek != lastWeeklyWeek {
				log.Println("[SCHEDULER] Triggering Weekly Leaderboard Reset...")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_ = b.db.ResetPeriodCounters(ctx, "weekly")
				cancel()
				lastWeeklyWeek = currentWeek
			}

			// Monthly Reset check (1st of month)
			if now.Month() != lastMonthlyMonth {
				log.Println("[SCHEDULER] Triggering Monthly Leaderboard Reset...")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				_ = b.db.ResetPeriodCounters(ctx, "monthly")
				cancel()
				lastMonthlyMonth = now.Month()
			}
		}
	}()
}
