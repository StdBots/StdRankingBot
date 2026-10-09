package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config represents runtime configurations for StdRankingBot
type Config struct {
	BotToken          string
	OwnerID           int64
	MongoURI          string
	DBName            string
	ForceSubChannel   string
	LogChannelID      int64
	AntiSpamCooldown  int // seconds
	MinMsgLength      int
	BaseXpPerMsg      int
	Env               string
}

// Load loads configurations from environment variables or .env file
func Load() *Config {
	_ = godotenv.Load()

	botToken := strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if botToken == "" {
		log.Panic("FATAL: BOT_TOKEN is required in environment variables or .env file")
	}

	ownerIDStr := strings.TrimSpace(os.Getenv("OWNER_ID"))
	if ownerIDStr == "" {
		ownerIDStr = "7394590844"
	}
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		log.Panicf("FATAL: Invalid OWNER_ID '%s': %v", ownerIDStr, err)
	}

	mongoURI := strings.TrimSpace(os.Getenv("MONGO_URI"))
	if mongoURI == "" {
		mongoURI = strings.TrimSpace(os.Getenv("DB_URL"))
	}
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/stdrankingbot"
	}

	dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
	if dbName == "" {
		dbName = "stdrankingbot"
	}

	forceSubChannel := strings.TrimPrefix(strings.TrimSpace(os.Getenv("FORCE_SUB_CHANNEL")), "@")

	var logChannelID int64
	logChannelStr := strings.TrimSpace(os.Getenv("LOG_CHANNEL_ID"))
	if logChannelStr != "" {
		logChannelID, _ = strconv.ParseInt(logChannelStr, 10, 64)
	}

	cooldown := 3
	cooldownStr := strings.TrimSpace(os.Getenv("ANTI_SPAM_COOLDOWN_SECONDS"))
	if cooldownStr != "" {
		if c, err := strconv.Atoi(cooldownStr); err == nil && c >= 0 {
			cooldown = c
		}
	}

	minLen := 3
	minLenStr := strings.TrimSpace(os.Getenv("MIN_MSG_LENGTH"))
	if minLenStr != "" {
		if m, err := strconv.Atoi(minLenStr); err == nil && m > 0 {
			minLen = m
		}
	}

	baseXp := 15
	baseXpStr := strings.TrimSpace(os.Getenv("BASE_XP_PER_MSG"))
	if baseXpStr != "" {
		if x, err := strconv.Atoi(baseXpStr); err == nil && x > 0 {
			baseXp = x
		}
	}

	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		env = "production"
	}

	return &Config{
		BotToken:         botToken,
		OwnerID:          ownerID,
		MongoURI:         mongoURI,
		DBName:           dbName,
		ForceSubChannel:  forceSubChannel,
		LogChannelID:     logChannelID,
		AntiSpamCooldown: cooldown,
		MinMsgLength:     minLen,
		BaseXpPerMsg:     baseXp,
		Env:              env,
	}
}
