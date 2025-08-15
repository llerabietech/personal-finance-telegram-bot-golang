package state

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type UserState string

const (
	Idle                     UserState = "idle"
	AwaitingCategoryName     UserState = "awaiting_category_name"
	AwaitingCategoryLimit    UserState = "awaiting_category_limit"
	AwaitingLimitUpdate      UserState = "awaiting_limit_update"
	AwaitingNewLimitValue    UserState = "awaiting_new_limit_value"
	AwaitingCategoryToDelete UserState = "awaiting_category_to_delete"
	ConfirmDeleteCategory    UserState = "confirm_delete_category"
	MainMenu                 UserState = "main_menu"
	CategoriesMenu           UserState = "categories_menu"
	LimitsMenu               UserState = "limits_menu"
	StateChoosingLanguage              = "choosing_language"
	StateMainMenu                      = "main_menu"
)

const StateTTL = 5 * time.Minute

func SetState(ctx context.Context, redis *redis.Client, chatID int64, state UserState) error {
	key := stateKey(chatID)
	return redis.Set(ctx, key, string(state), StateTTL).Err()
}

func GetState(ctx context.Context, redis *redis.Client, chatID int64) (UserState, error) {
	key := stateKey(chatID)
	val, err := redis.Get(ctx, key).Result()
	if err == nil {
		return UserState(val), nil
	}
	return Idle, err
}

func SetTempData(ctx context.Context, redis *redis.Client, chatID int64, data string) error {
	key := tempKey(chatID)
	return redis.Set(ctx, key, data, StateTTL).Err()
}

func GetTempData(ctx context.Context, redis *redis.Client, chatID int64) (string, error) {
	key := tempKey(chatID)
	return redis.Get(ctx, key).Result()
}

func Clear(ctx context.Context, redis *redis.Client, chatID int64) {
	redis.Del(ctx, stateKey(chatID))
	redis.Del(ctx, tempKey(chatID))
}

func stateKey(chatID int64) string {
	return "user:state:" + strconv.FormatInt(chatID, 10)
}

func tempKey(chatID int64) string {
	return "user:temp:" + strconv.FormatInt(chatID, 10)
}

func SetUserLanguage(ctx context.Context, redis *redis.Client, chatID int64, lang string) error {
	return redis.Set(ctx, "user:lang:"+strconv.FormatInt(chatID, 10), lang, 0).Err()
}

func GetUserLanguage(ctx context.Context, redis *redis.Client, chatID int64) (string, error) {
	return redis.Get(ctx, "user:lang:"+strconv.FormatInt(chatID, 10)).Result()
}
