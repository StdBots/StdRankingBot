package database

import "time"

// MemberStats stores user statistics within a specific group
type MemberStats struct {
	ChatID          int64     `bson:"chat_id" json:"chat_id"`
	UserID          int64     `bson:"user_id" json:"user_id"`
	Username        string    `bson:"username" json:"username"`
	FirstName       string    `bson:"first_name" json:"first_name"`
	DailyCount      int64     `bson:"daily_count" json:"daily_count"`
	WeeklyCount     int64     `bson:"weekly_count" json:"weekly_count"`
	MonthlyCount    int64     `bson:"monthly_count" json:"monthly_count"`
	AllTimeCount    int64     `bson:"all_time_count" json:"all_time_count"`
	XP              int64     `bson:"xp" json:"xp"`
	Level           int       `bson:"level" json:"level"`
	LastActiveAt    time.Time `bson:"last_active_at" json:"last_active_at"`
	LastMessageHash string    `bson:"last_message_hash" json:"last_message_hash"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updated_at"`
}

// GroupSettings stores group-level configuration and reset timestamps
type GroupSettings struct {
	ChatID          int64     `bson:"chat_id" json:"chat_id"`
	Title           string    `bson:"title" json:"title"`
	Username        string    `bson:"username" json:"username"`
	ChatType        string    `bson:"chat_type" json:"chat_type"` // group, supergroup
	Active          bool      `bson:"active" json:"active"`
	LevelUpAlerts   bool      `bson:"level_up_alerts" json:"level_up_alerts"`
	MinMsgLength    int       `bson:"min_msg_length" json:"min_msg_length"`
	CooldownSeconds int       `bson:"cooldown_seconds" json:"cooldown_seconds"`
	LastDailyReset  time.Time `bson:"last_daily_reset" json:"last_daily_reset"`
	LastWeeklyReset time.Time `bson:"last_weekly_reset" json:"last_weekly_reset"`
	LastMonthlyReset time.Time `bson:"last_monthly_reset" json:"last_monthly_reset"`
	TotalMessages   int64     `bson:"total_messages" json:"total_messages"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updated_at"`
}

// GlobalUser represents cross-group statistics for a user
type GlobalUser struct {
	UserID        int64     `bson:"user_id" json:"user_id"`
	Username      string    `bson:"username" json:"username"`
	FirstName     string    `bson:"first_name" json:"first_name"`
	TotalMessages int64     `bson:"total_messages" json:"total_messages"`
	GlobalXP      int64     `bson:"global_xp" json:"global_xp"`
	GlobalLevel   int       `bson:"global_level" json:"global_level"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

// SystemStats for analytics & admin dashboard
type SystemStats struct {
	TotalGroups    int64 `json:"total_groups"`
	TotalUsers     int64 `json:"total_users"`
	TotalMessages  int64 `json:"total_messages"`
	DailyMessages  int64 `json:"daily_messages"`
	WeeklyMessages int64 `json:"weekly_messages"`
}
