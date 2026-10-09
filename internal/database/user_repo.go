package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureGroup verifies group exists in DB and returns its configuration
func (db *DB) EnsureGroup(ctx context.Context, chatID int64, title, username, chatType string) (*GroupSettings, error) {
	now := time.Now()
	var group GroupSettings

	err := db.Groups.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&group)
	if err == mongo.ErrNoDocuments {
		group = GroupSettings{
			ChatID:          chatID,
			Title:           title,
			Username:        username,
			ChatType:        chatType,
			Active:          true,
			LevelUpAlerts:   true,
			MinMsgLength:    3,
			CooldownSeconds: 3,
			LastDailyReset:  now,
			LastWeeklyReset: now,
			LastMonthlyReset: now,
			TotalMessages:   0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		_, err = db.Groups.InsertOne(ctx, group)
		if err != nil {
			return nil, err
		}
		return &group, nil
	} else if err != nil {
		return nil, err
	}

	// Update title or username if changed
	if group.Title != title || group.Username != username {
		_, _ = db.Groups.UpdateOne(ctx, bson.M{"chat_id": chatID}, bson.M{
			"$set": bson.M{
				"title":      title,
				"username":   username,
				"updated_at": now,
			},
		})
		group.Title = title
		group.Username = username
	}

	return &group, nil
}

// UpdateGroupSettings saves updated group preferences
func (db *DB) UpdateGroupSettings(ctx context.Context, group *GroupSettings) error {
	group.UpdatedAt = time.Now()
	_, err := db.Groups.ReplaceOne(ctx, bson.M{"chat_id": group.ChatID}, group, options.Replace().SetUpsert(true))
	return err
}

// GetGlobalLeaderboard fetches top players across all groups
func (db *DB) GetGlobalLeaderboard(ctx context.Context, limit int) ([]GlobalUser, error) {
	if limit <= 0 {
		limit = 10
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "global_xp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := db.GlobalUsers.Find(ctx, bson.M{"global_xp": bson.M{"$gt": 0}}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []GlobalUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetSystemStats computes system-wide analytics for admin panel
func (db *DB) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	totalGroups, err := db.Groups.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	totalUsers, err := db.GlobalUsers.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Sum total messages
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalMsgs", Value: bson.D{{Key: "$sum", Value: "$all_time_count"}}},
			{Key: "dailyMsgs", Value: bson.D{{Key: "$sum", Value: "$daily_count"}}},
			{Key: "weeklyMsgs", Value: bson.D{{Key: "$sum", Value: "$weekly_count"}}},
		}}},
	}

	cursor, err := db.Members.Aggregate(ctx, pipeline)
	var totalMsgs, dailyMsgs, weeklyMsgs int64
	if err == nil && cursor.Next(ctx) {
		var result struct {
			TotalMsgs  int64 `bson:"totalMsgs"`
			DailyMsgs  int64 `bson:"dailyMsgs"`
			WeeklyMsgs int64 `bson:"weeklyMsgs"`
		}
		if err := cursor.Decode(&result); err == nil {
			totalMsgs = result.TotalMsgs
			dailyMsgs = result.DailyMsgs
			weeklyMsgs = result.WeeklyMsgs
		}
	}

	return &SystemStats{
		TotalGroups:    totalGroups,
		TotalUsers:     totalUsers,
		TotalMessages:  totalMsgs,
		DailyMessages:  dailyMsgs,
		WeeklyMessages: weeklyMsgs,
	}, nil
}

// GetAllGroupChatIDs retrieves all group IDs for broadcasts
func (db *DB) GetAllGroupChatIDs(ctx context.Context) ([]int64, error) {
	cursor, err := db.Groups.Find(ctx, bson.M{"active": true}, options.Find().SetProjection(bson.M{"chat_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []int64
	for cursor.Next(ctx) {
		var g struct {
			ChatID int64 `bson:"chat_id"`
		}
		if err := cursor.Decode(&g); err == nil {
			ids = append(ids, g.ChatID)
		}
	}

	return ids, nil
}
