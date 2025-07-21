package user

import "time"

// User - модель пользователя
type User struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// Ad - модель объявления
type Ad struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`   // Связь с автором
	Title     string    `json:"title"`     // Ограничение: 100 символов
	Text      string    `json:"text"`      // Ограничение: 1000 символов
	ImageURL  string    `json:"image_url"` // Просто строка с URL
	Price     float64   `json:"price"`     // DECIMAL в БД
	CreatedAt time.Time `json:"created_at"`
}

// AuthToken - JWT токен
type AuthToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
