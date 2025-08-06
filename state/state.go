package state

import (
	"personal-finance/db"
	"time"
	"strconv"
)

type UserState string

const (
	Idle                  UserState = "idle"
	AwaitingCategoryName  UserState = "awaiting_category_name"
	AwaitingCategoryLimit UserState = "awaiting_category_limit"
	AwaitingLimitUpdate   UserState = "awaiting_limit_update"
)

const StateTTL = 5 * time.Minute // Время жизни состояния (защита от "зависаний")

// SetState — устанавливает состояние пользователя
func SetState(chatID int64, state UserState) error {
	key := stateKey(chatID)
	return db.RedisClient.Set(db.Ctx, key, string(state), StateTTL).Err()
}

// GetState — получает состояние пользователя
func GetState(chatID int64) (UserState, error) {
	key := stateKey(chatID)
	val, err := db.RedisClient.Get(db.Ctx, key).Result()
	if err == nil {
		return UserState(val), nil
	}
	return Idle, err // если нет — возвращаем Idle
}

// SetTempData — временные данные (например, имя категории)
func SetTempData(chatID int64, data string) error {
	key := tempKey(chatID)
	return db.RedisClient.Set(db.Ctx, key, data, StateTTL).Err()
}

// GetTempData — получить временные данные
func GetTempData(chatID int64) (string, error) {
	key := tempKey(chatID)
	return db.RedisClient.Get(db.Ctx, key).Result()
}

// Clear — очистить состояние и временные данные
func Clear(chatID int64) {
	db.RedisClient.Del(db.Ctx, stateKey(chatID))
	db.RedisClient.Del(db.Ctx, tempKey(chatID))
}

// Вспомогательные функции для ключей
func stateKey(chatID int64) string {
    return "user:state:" + strconv.FormatInt(chatID, 10)
}

func tempKey(chatID int64) string {
    return "user:temp:" + strconv.FormatInt(chatID, 10)
}
