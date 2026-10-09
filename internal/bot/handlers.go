package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/StdBots/StdRankingBot/internal/credit"
	"github.com/StdBots/StdRankingBot/internal/database"
	"github.com/StdBots/StdRankingBot/internal/graphics"
)

// HandleCommand routes incoming slash commands
func (b *Bot) HandleCommand(msg *tgbotapi.Message) {
	cmd := strings.ToLower(msg.Command())

	switch cmd {
	case "start":
		b.handleStart(msg)
	case "top", "ranking", "leaderboard":
		b.handleLeaderboard(msg, "daily")
	case "me", "profile", "rank":
		b.handleProfile(msg)
	case "global":
		b.handleGlobal(msg)
	case "settings":
		b.handleSettings(msg)
	case "help":
		b.handleHelp(msg)
	case "stats":
		b.handleStats(msg)
	case "broadcast":
		b.handleBroadcast(msg)
	default:
		// Unknown command
	}
}

// handleStart responds to /start
func (b *Bot) handleStart(msg *tgbotapi.Message) {
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		text := fmt.Sprintf(
			"⚔️ <b>StdRankingBot is ACTIVE in %s!</b>\n\n"+
				"• Send messages to earn XP and level up.\n"+
				"• View rankings with /top\n"+
				"• Check your gamer card with /me\n\n"+
				"%s",
			msg.Chat.Title, credit.GetFooter(),
		)
		reply := tgbotapi.NewMessage(msg.Chat.ID, credit.WatermarkMessage(text))
		reply.ParseMode = tgbotapi.ModeHTML
		_, _ = b.api.Send(reply)
		return
	}

	welcomeText := fmt.Sprintf(
		"🏆 <b>Welcome to @%s!</b>\n\n"+
			"The ultimate Telegram <b>Gamification & Chat Ranking System</b>.\n"+
			"Inspired by ChatFight with ultra-fast HD Leaderboard Banners & Gamer Cards!\n\n"+
			"✨ <b>Key Features:</b>\n"+
			"├ 🖼 <b>Dynamic HD Banners:</b> Daily, Weekly, Monthly & All-Time\n"+
			"├ 🎮 <b>Fighter Profiles:</b> Custom Gamer Cards (/me)\n"+
			"├ ⚡ <b>XP Progression:</b> Earn XP & unlock 9 Tier Titles\n"+
			"├ 🛡 <b>Anti-Farming:</b> Cooldowns & duplicate text detection\n"+
			"└ 🌐 <b>Global Hall of Fame:</b> Compete across Telegram groups\n\n"+
			"<i>Add me to your group and give admin permissions to start ranking!</i>\n\n"+
			"%s",
		b.api.Self.UserName, credit.GetFooter(),
	)

	botUsername := b.api.Self.UserName
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("➕ Add to Group", fmt.Sprintf("https://t.me/%s?startgroup=true", botUsername)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📖 Help & Guides", "help:main"),
			tgbotapi.NewInlineKeyboardButtonData("🌐 Global Top", "rank:global"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 Updates Channel", "https://t.me/STDBOTS"),
			tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Developer", "https://deepanshu.in"),
		),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.WatermarkMessage(welcomeText))
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = keyboard
	_, _ = b.api.Send(reply)
}

// handleLeaderboard generates and sends the HD Leaderboard Banner
func (b *Bot) handleLeaderboard(msg *tgbotapi.Message, period string) {
	if msg.Chat.IsPrivate() {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ /top is designed for groups. Use /global to see global rankings, or add me to a group!")
		_, _ = b.api.Send(reply)
		return
	}

	chatID := msg.Chat.ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	members, err := b.db.GetLeaderboard(ctx, chatID, period, 10)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "❌ Error retrieving leaderboard data.")
		_, _ = b.api.Send(reply)
		return
	}

	if len(members) == 0 {
		reply := tgbotapi.NewMessage(chatID, "🌱 No message stats recorded yet for this period. Start chatting to climb the ranks!")
		_, _ = b.api.Send(reply)
		return
	}

	// Prepare data for graphics engine
	var rankedMembers []graphics.RankedMember
	var totalMsgs int64
	for i, m := range members {
		var cnt int64
		switch period {
		case "weekly":
			cnt = m.WeeklyCount
		case "monthly":
			cnt = m.MonthlyCount
		case "alltime":
			cnt = m.AllTimeCount
		default:
			cnt = m.DailyCount
		}
		totalMsgs += cnt

		name := m.FirstName
		if name == "" {
			name = m.Username
		}
		rankedMembers = append(rankedMembers, graphics.RankedMember{
			Rank:     i + 1,
			Name:     name,
			Username: m.Username,
			Level:    m.Level,
			Messages: cnt,
			XP:       m.XP,
		})
	}

	cardData := graphics.LeaderboardCardData{
		ChatTitle:   msg.Chat.Title,
		Period:      period,
		Members:     rankedMembers,
		TotalActive: int64(len(members)),
		TotalMsgs:   totalMsgs,
		GeneratedAt: time.Now(),
	}

	imgBytes, err := graphics.RenderLeaderboard(cardData)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "❌ Failed to render leaderboard banner.")
		_, _ = b.api.Send(reply)
		return
	}

	// Build textual leaderboard caption
	caption := b.buildLeaderboardCaption(msg.Chat.Title, period, rankedMembers)

	// Send Photo with period navigation buttons
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("leaderboard_%s.png", period),
		Bytes: imgBytes,
	})
	photo.Caption = credit.WatermarkMessage(caption)
	photo.ParseMode = tgbotapi.ModeHTML
	photo.ReplyMarkup = b.getLeaderboardKeyboard(period)

	_, _ = b.api.Send(photo)
}

func (b *Bot) buildLeaderboardCaption(chatTitle, period string, members []graphics.RankedMember) string {
	periodTitle := strings.ToUpper(period)
	if period == "alltime" {
		periodTitle = "ALL-TIME"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏆 <b>%s LEADERBOARD • %s</b>\n\n", periodTitle, chatTitle))

	for _, m := range members {
		medal := "▫️"
		switch m.Rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		}

		userTag := m.Name
		if m.Username != "" {
			userTag = fmt.Sprintf("<a href=\"https://t.me/%s\">%s</a>", m.Username, m.Name)
		}

		sb.WriteString(fmt.Sprintf("%s <b>#%d</b> %s • <code>%d msgs</code> (Lv.%d)\n",
			medal, m.Rank, userTag, m.Messages, m.Level,
		))
	}

	sb.WriteString(fmt.Sprintf("\n%s", credit.GetFooter()))
	return sb.String()
}

func (b *Bot) getLeaderboardKeyboard(activePeriod string) tgbotapi.InlineKeyboardMarkup {
	dLabel := "📅 Daily"
	wLabel := "📆 Weekly"
	mLabel := "🗓 Monthly"
	aLabel := "👑 All-Time"

	switch activePeriod {
	case "daily":
		dLabel = "• 📅 Daily •"
	case "weekly":
		wLabel = "• 📆 Weekly •"
	case "monthly":
		mLabel = "• 🗓 Monthly •"
	case "alltime":
		aLabel = "• 👑 All-Time •"
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dLabel, "rank:daily"),
			tgbotapi.NewInlineKeyboardButtonData(wLabel, "rank:weekly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(mLabel, "rank:monthly"),
			tgbotapi.NewInlineKeyboardButtonData(aLabel, "rank:alltime"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 My Card", "rank:me"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", fmt.Sprintf("rank:%s", activePeriod)),
		),
	)
}

// handleProfile generates and sends the 840x500 Gamer Profile Card
func (b *Bot) handleProfile(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rank, stats, totalMembers, err := b.db.GetMemberRank(ctx, chatID, userID, "alltime")
	if err != nil || stats == nil {
		reply := tgbotapi.NewMessage(chatID, "🌱 You have not earned any stats in this chat yet. Send some messages to unlock your gamer card!")
		_, _ = b.api.Send(reply)
		return
	}

	baseXP, nextXP := database.GetLevelThresholds(stats.Level)
	progressRatio := 0.0
	if nextXP > baseXP {
		progressRatio = float64(stats.XP-baseXP) / float64(nextXP-baseXP)
	}

	cardData := graphics.ProfileData{
		Name:          stats.FirstName,
		Username:      stats.Username,
		ChatTitle:     msg.Chat.Title,
		Rank:          rank,
		TotalMembers:  totalMembers,
		Level:         stats.Level,
		DailyMsgs:     stats.DailyCount,
		WeeklyMsgs:    stats.WeeklyCount,
		MonthlyMsgs:   stats.MonthlyCount,
		AllTimeMsgs:   stats.AllTimeCount,
		CurrentXP:     stats.XP,
		LevelBaseXP:   baseXP,
		NextLevelXP:   nextXP,
		ProgressRatio: progressRatio,
	}

	imgBytes, err := graphics.RenderProfileCard(cardData)
	if err != nil {
		reply := tgbotapi.NewMessage(chatID, "❌ Failed to render profile card.")
		_, _ = b.api.Send(reply)
		return
	}

	caption := fmt.Sprintf(
		"🎮 <b>FIGHTER PROFILE • %s</b>\n\n"+
			"👤 Name: <b>%s</b>\n"+
			"🏆 Rank: <b>#%d</b> of %d\n"+
			"⚔️ Title: <b>%s</b>\n"+
			"🌟 Level: <b>%d</b>\n"+
			"⚡ Total XP: <code>%d XP</code>\n"+
			"💬 All-Time Msgs: <code>%d</code>\n\n"+
			"%s",
		msg.Chat.Title,
		stats.FirstName,
		rank, totalMembers,
		graphics.GetUserTitle(stats.Level),
		stats.Level,
		stats.XP,
		stats.AllTimeCount,
		credit.GetFooter(),
	)

	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("profile_%d.png", userID),
		Bytes: imgBytes,
	})
	photo.Caption = credit.WatermarkMessage(caption)
	photo.ParseMode = tgbotapi.ModeHTML
	photo.ReplyToMessageID = msg.MessageID

	_, _ = b.api.Send(photo)
}

// handleGlobal sends cross-chat global rankings
func (b *Bot) handleGlobal(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := b.db.GetGlobalLeaderboard(ctx, 10)
	if err != nil || len(users) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "🌐 Global leaderboards are initializing...")
		_, _ = b.api.Send(reply)
		return
	}

	var sb strings.Builder
	sb.WriteString("🌐 <b>GLOBAL HALL OF FAME (Cross-Group)</b>\n\n")

	for i, u := range users {
		medal := "▫️"
		switch i + 1 {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		}

		userTag := u.FirstName
		if u.Username != "" {
			userTag = fmt.Sprintf("<a href=\"https://t.me/%s\">%s</a>", u.Username, u.FirstName)
		}

		sb.WriteString(fmt.Sprintf("%s <b>#%d</b> %s • <code>%d msgs</code> (Lv.%d • %d XP)\n",
			medal, i+1, userTag, u.TotalMessages, u.GlobalLevel, u.GlobalXP,
		))
	}

	sb.WriteString(fmt.Sprintf("\n%s", credit.GetFooter()))

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.WatermarkMessage(sb.String()))
	reply.ParseMode = tgbotapi.ModeHTML
	_, _ = b.api.Send(reply)
}

// handleSettings allows group admins to configure ranking options
func (b *Bot) handleSettings(msg *tgbotapi.Message) {
	if msg.Chat.IsPrivate() {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ /settings can only be used in groups by administrators.")
		_, _ = b.api.Send(reply)
		return
	}

	if !b.IsChatAdmin(msg.Chat.ID, msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ You must be an administrator in this group to configure ranking settings.")
		_, _ = b.api.Send(reply)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	group, err := b.db.EnsureGroup(ctx, msg.Chat.ID, msg.Chat.Title, msg.Chat.UserName, msg.Chat.Type)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to load group settings.")
		_, _ = b.api.Send(reply)
		return
	}

	alertStatus := "ENABLED 🔔"
	if !group.LevelUpAlerts {
		alertStatus = "DISABLED 🔕"
	}

	text := fmt.Sprintf(
		"⚙️ <b>RANKING CONFIGURATION • %s</b>\n\n"+
			"• <b>Level-Up Alerts:</b> %s\n"+
			"• <b>Anti-Spam Cooldown:</b> <code>%d seconds</code>\n"+
			"• <b>Minimum Message Length:</b> <code>%d chars</code>\n"+
			"• <b>Total Messages Counted:</b> <code>%d</code>\n\n"+
			"Toggle settings using the buttons below:",
		msg.Chat.Title, alertStatus, group.CooldownSeconds, group.MinMsgLength, group.TotalMessages,
	)

	btnAlert := "🔔 Alerts: ON"
	if !group.LevelUpAlerts {
		btnAlert = "🔕 Alerts: OFF"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnAlert, "sett:toggle_alerts"),
		),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.WatermarkMessage(text))
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyMarkup = keyboard
	_, _ = b.api.Send(reply)
}

// handleHelp explains ranking mechanics
func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	helpText := fmt.Sprintf(
		"📖 <b>HOW @%s WORKS:</b>\n\n"+
			"<b>1. Chat & Earn XP:</b>\n"+
			"Every valid chat message awards <b>15 XP</b> (+ bonus for longer detailed messages).\n\n"+
			"<b>2. Anti-Farming System:</b>\n"+
			"To prevent cheating, messages have a 3-second cooldown and duplicate messages are ignored.\n\n"+
			"<b>3. Levels & Fighter Titles:</b>\n"+
			"As you gain XP, you level up and gain legendary titles:\n"+
			"• Lv. 1-4: 🌱 Rookie Fighter\n"+
			"• Lv. 5-9: 🗡️ Apprentice\n"+
			"• Lv. 10-14: 🎯 Warrior\n"+
			"• Lv. 15-24: 🔥 Battle Knight\n"+
			"• Lv. 25-34: 🛡️ Elite Veteran\n"+
			"• Lv. 35-49: ⚔️ Mythic Gladiator\n"+
			"• Lv. 50-74: ⚡ Grandmaster Fighter\n"+
			"• Lv. 75-99: 🔱 Celestial Warlord\n"+
			"• Lv. 100+: 👑 Omnipotent God\n\n"+
			"<b>Commands:</b>\n"+
			"• /top or /ranking - Group leaderboard banners\n"+
			"• /me or /profile - Your personal gamer card\n"+
			"• /global - Global Telegram rankings\n"+
			"• /settings - Group administrator controls\n\n"+
			"%s",
		b.api.Self.UserName, credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.WatermarkMessage(helpText))
	reply.ParseMode = tgbotapi.ModeHTML
	_, _ = b.api.Send(reply)
}

// handlePrivateMessage handles non-command private interactions
func (b *Bot) handlePrivateMessage(msg *tgbotapi.Message) {
	b.handleStart(msg)
}
