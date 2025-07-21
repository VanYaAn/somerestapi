package models

type Ad struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	UserID      int     `json:"user_id"`
	Login       string  `json:"login"`
	CreatedAt   string  `json:"created_at"`
	IsOwner     bool    `json:"is_owner,omitempty"`
}
