package bot

import (
	"github.com/StdBots/StdRankingBot/internal/analytics"
	"github.com/StdBots/StdRankingBot/internal/broadcast"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStats delegates to analytics engine
func (b *Bot) handleStats(msg *tgbotapi.Message) {
	engine := analytics.NewEngine(b.api, b.db, b.cfg.OwnerID)
	engine.HandleStats(msg)
}

// handleBroadcast delegates to broadcast engine
func (b *Bot) handleBroadcast(msg *tgbotapi.Message) {
	engine := broadcast.NewEngine(b.api, b.db, b.cfg.OwnerID)
	engine.HandleBroadcast(msg)
}
