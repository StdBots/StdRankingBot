package database

import (
	"context"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CalculateLevel computes level based on total XP: Level = floor(sqrt(XP / 50)) + 1
func CalculateLevel(xp int64) int {
	if xp <= 0 {
		return 1
	}
	lvl := int(math.Sqrt(float64(xp) / 50.0)) + 1
	if lvl < 1 {
		return 1
	}
	return lvl
}

// GetLevelThresholds returns the base XP of the current level and required XP for next level
func GetLevelThresholds(level int) (baseXP int64, nextXP int64) {
	if level <= 1 {
		return 0, 50
	}
	baseXP = int64(50 * math.Pow(float64(level-1), 2))
	nextXP = int64(50 * math.Pow(float64(level), 2))
	return baseXP, nextXP
}

// RecordMessage handles XP addition, count increments, level calculation, and global sync
func (db *DB) RecordMessage(ctx context.Context, chatID, userID int64, username, firstName, msgText string, xpGain int) (*MemberStats, bool, error) {
	now := time.Now()

	// 1. Fetch existing member stats or initialize
	var current MemberStats
	err := db.Members.FindOne(ctx, bson.M{"chat_id": chatID, "user_id": userID}).Decode(&current)
	isNew := false
	if err != nil {
		if err == mongo.ErrNoDocuments {
			isNew = true
			current = MemberStats{
				ChatID:       chatID,
				UserID:       userID,
				Username:     username,
				FirstName:    firstName,
				DailyCount:   0,
				WeeklyCount:  0,
				MonthlyCount: 0,
				AllTimeCount: 0,
				XP:           0,
				Level:        1,
				CreatedAt:    now,
			}
		} else {
			return nil, false, err
		}
	}

	oldLevel := current.Level
	if oldLevel < 1 {
		oldLevel = 1
	}

	newXP := current.XP + int64(xpGain)
	newLevel := CalculateLevel(newXP)
	leveledUp := newLevel > oldLevel && !isNew

	// Update document
	update := bson.M{
		"$inc": bson.M{
			"daily_count":    1,
			"weekly_count":   1,
			"monthly_count":  1,
			"all_time_count": 1,
			"xp":             int64(xpGain),
		},
		"$set": bson.M{
			"username":       username,
			"first_name":     firstName,
			"level":          newLevel,
			"last_active_at": now,
			"updated_at":     now,
		},
	}

	if isNew {
		update["$setOnInsert"] = bson.M{
			"created_at": now,
		}
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updated MemberStats
	err = db.Members.FindOneAndUpdate(ctx, bson.M{"chat_id": chatID, "user_id": userID}, update, opts).Decode(&updated)
	if err != nil {
		return nil, false, err
	}

	// Sync to Global Users
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var gUser GlobalUser
		_ = db.GlobalUsers.FindOne(bgCtx, bson.M{"user_id": userID}).Decode(&gUser)
		newGlobalXP := gUser.GlobalXP + int64(xpGain)
		newGlobalLevel := CalculateLevel(newGlobalXP)

		gUpdate := bson.M{
			"$inc": bson.M{
				"total_messages": 1,
				"global_xp":      int64(xpGain),
			},
			"$set": bson.M{
				"username":     username,
				"first_name":   firstName,
				"global_level": newGlobalLevel,
				"updated_at":   now,
			},
			"$setOnInsert": bson.M{
				"created_at": now,
			},
		}
		_, _ = db.GlobalUsers.UpdateOne(bgCtx, bson.M{"user_id": userID}, gUpdate, options.Update().SetUpsert(true))

		// Increment group total message counter
		_, _ = db.Groups.UpdateOne(bgCtx, bson.M{"chat_id": chatID}, bson.M{
			"$inc": bson.M{"total_messages": 1},
			"$set": bson.M{"updated_at": now},
		})
	}()

	return &updated, leveledUp, nil
}

// GetLeaderboard retrieves top members for a given chat and period
func (db *DB) GetLeaderboard(ctx context.Context, chatID int64, period string, limit int) ([]MemberStats, error) {
	if limit <= 0 {
		limit = 10
	}

	sortField := "daily_count"
	switch strings.ToLower(period) {
	case "weekly":
		sortField = "weekly_count"
	case "monthly":
		sortField = "monthly_count"
	case "alltime":
		sortField = "all_time_count"
	}

	filter := bson.M{
		"chat_id": chatID,
		sortField: bson.M{"$gt": 0},
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: -1}, {Key: "xp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := db.Members.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []MemberStats
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// GetMemberRank calculates a user's exact numerical rank in a group
func (db *DB) GetMemberRank(ctx context.Context, chatID, userID int64, period string) (int, *MemberStats, int, error) {
	var userStats MemberStats
	err := db.Members.FindOne(ctx, bson.M{"chat_id": chatID, "user_id": userID}).Decode(&userStats)
	if err != nil {
		return 0, nil, 0, err
	}

	sortField := "daily_count"
	var userCount int64 = userStats.DailyCount
	switch strings.ToLower(period) {
	case "weekly":
		sortField = "weekly_count"
		userCount = userStats.WeeklyCount
	case "monthly":
		sortField = "monthly_count"
		userCount = userStats.MonthlyCount
	case "alltime":
		sortField = "all_time_count"
		userCount = userStats.AllTimeCount
	}

	// Count users with higher count in this group
	higherFilter := bson.M{
		"chat_id": chatID,
		sortField: bson.M{"$gt": userCount},
	}
	higherCount, err := db.Members.CountDocuments(ctx, higherFilter)
	if err != nil {
		return 0, &userStats, 0, err
	}

	// Total group members tracked
	totalGroupMembers, _ := db.Members.CountDocuments(ctx, bson.M{"chat_id": chatID})

	return int(higherCount) + 1, &userStats, int(totalGroupMembers), nil
}

// ResetPeriodCounters resets daily, weekly, or monthly message counts across all users
func (db *DB) ResetPeriodCounters(ctx context.Context, period string) error {
	field := ""
	switch strings.ToLower(period) {
	case "daily":
		field = "daily_count"
	case "weekly":
		field = "weekly_count"
	case "monthly":
		field = "monthly_count"
	default:
		return nil
	}

	_, err := db.Members.UpdateMany(ctx, bson.M{field: bson.M{"$gt": 0}}, bson.M{"$set": bson.M{field: 0}})
	return err
}
