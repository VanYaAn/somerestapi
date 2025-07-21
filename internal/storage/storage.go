package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"restapi/internal/models"
	"strings"
	"time"
)

type Storage interface {
	RegisterUser(ctx context.Context, login, password string) (int, error)
	CheckUser(ctx context.Context, login, password string) (int, error)
	CreateAd(ctx context.Context, userID int, title, description, imageURL string, price float64) (int, error)
	GetAds(ctx context.Context, page, pageSize int, sortBy, sortOrder string, minPrice, maxPrice float64) ([]models.Ad, error)
	IsAdOwner(ctx context.Context, adID, userID int) (bool, error)
}
type StoragePostgresql struct {
	Database *sql.DB
}

func New(ctx context.Context, port string, username string, host string, DBname string, password string) *StoragePostgresql {
	// dsn := fmt.Sprintf("postgres://postgres:%s@%s:%s/%s", password, host, port, DBname)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, DBname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("DB connection is failed:%v", err)
		return nil
	}
	err = db.Ping()
	if err != nil {
		log.Println("bad connection:%v", err)
		return nil

	}

	return &StoragePostgresql{Database: db}
}
func (db *StoragePostgresql) RegisterUser(ctx context.Context, login, password string) (int, error) {
	var userID int
	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"
	err := db.Database.QueryRowContext(ctx, query, login, password).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to register user: %v", err)
	}
	return userID, nil
}
func (db *StoragePostgresql) CheckUser(ctx context.Context, login, password string) (int, error) {
	var userID int
	query := "SELECT id FROM users WHERE login = $1 AND password = $2"
	err := db.Database.QueryRowContext(ctx, query, login, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid login or password")
		}
		return 0, fmt.Errorf("failed to check user: %v", err)
	}
	return userID, nil
}

func (db *StoragePostgresql) CreateAd(ctx context.Context, userID int, title, description, imageURL string, price float64) (int, error) {
	var adID int
	query := "INSERT INTO ads (title, description, image_url, price, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := db.Database.QueryRowContext(ctx, query, title, description, imageURL, price, userID).Scan(&adID)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return 0, fmt.Errorf("user with ID %d does not exist", userID)
		}
		return 0, fmt.Errorf("failed to create ad: %v", err)
	}
	return adID, nil
}
func (db *StoragePostgresql) GetAds(ctx context.Context, page, pageSize int, sortBy, sortOrder string, minPrice, maxPrice float64) ([]models.Ad, error) {
	// Валидация параметров
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	if sortBy != "created_at" && sortBy != "price" {
		sortBy = "created_at"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
        SELECT a.id, a.title, a.description, a.image_url, a.price, a.user_id, u.login, a.created_at
        FROM ads a
        JOIN users u ON a.user_id = u.id
        WHERE a.price BETWEEN $1 AND $2
        ORDER BY a.%s %s
        LIMIT $3 OFFSET $4`, sortBy, sortOrder)

	rows, err := db.Database.QueryContext(ctx, query, minPrice, maxPrice, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get ads: %v", err)
	}
	defer rows.Close()

	var ads []models.Ad
	for rows.Next() {
		var ad models.Ad
		var createdAt time.Time
		if err := rows.Scan(&ad.ID, &ad.Title, &ad.Description, &ad.ImageURL, &ad.Price, &ad.UserID, &ad.Login, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan ad: %v", err)
		}
		ad.CreatedAt = createdAt.Format(time.RFC3339)
		ads = append(ads, ad)
	}
	return ads, nil
}

// IsAdOwner проверяет, является ли пользователь владельцем объявления
func (db *StoragePostgresql) IsAdOwner(ctx context.Context, adID, userID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM ads WHERE id = $1 AND user_id = $2)"
	err := db.Database.QueryRowContext(ctx, query, adID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check ad owner: %v", err)
	}
	return exists, nil
}
