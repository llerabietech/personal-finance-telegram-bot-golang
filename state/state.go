package state

import (
	"personal-finance/db"
	"strconv"
	"time"
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

func SetState(chatID int64, state UserState) error {
	key := stateKey(chatID)
	return db.RedisClient.Set(db.Ctx, key, string(state), StateTTL).Err()
}

func GetState(chatID int64) (UserState, error) {
	key := stateKey(chatID)
	val, err := db.RedisClient.Get(db.Ctx, key).Result()
	if err == nil {
		return UserState(val), nil
	}
	return Idle, err
}

func SetTempData(chatID int64, data string) error {
	key := tempKey(chatID)
	return db.RedisClient.Set(db.Ctx, key, data, StateTTL).Err()
}

func GetTempData(chatID int64) (string, error) {
	key := tempKey(chatID)
	return db.RedisClient.Get(db.Ctx, key).Result()
}

func Clear(chatID int64) {
	db.RedisClient.Del(db.Ctx, stateKey(chatID))
	db.RedisClient.Del(db.Ctx, tempKey(chatID))
}

func stateKey(chatID int64) string {
	return "user:state:" + strconv.FormatInt(chatID, 10)
}

func tempKey(chatID int64) string {
	return "user:temp:" + strconv.FormatInt(chatID, 10)
}

func SetUserLanguage(chatID int64, lang string) error {
    return db.RedisClient.Set(db.Ctx, "user:lang:"+strconv.FormatInt(chatID, 10), lang, 0).Err()
}

func GetUserLanguage(chatID int64) (string, error) {
    return db.RedisClient.Get(db.Ctx, "user:lang:"+strconv.FormatInt(chatID, 10)).Result()
}
