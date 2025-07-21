package service

import (
	"context"
	"net/http"
	"os"
	"restapi/internal/models"
	"restapi/internal/storage"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type Service struct {
	Port        string
	Host        string
	logger      *zap.SugaredLogger
	StorageImpl storage.Storage
}

func NewService(port, host string, logger *zap.SugaredLogger, storage storage.Storage) *Service {
	return &Service{
		Port:        port,
		Host:        host,
		logger:      logger,
		StorageImpl: storage,
	}
}

var JWTKey = string(os.Getenv("JWT_KEY"))

// ListenAndServe запускает HTTP-сервер
func (s *Service) ListenAndServe(handler http.Handler) error {
	httpServer := &http.Server{
		Addr:         s.Host + s.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	s.logger.Infof("Starting server at %s%s", s.Host, s.Port)
	return httpServer.ListenAndServe()
}

// RegisterUser регистрирует нового пользователя
func (s *Service) RegisterUser(ctx context.Context, login, password string) (int, error) {
	s.logger.Infof("Registering user with login: %s", login)
	id, err := s.StorageImpl.RegisterUser(ctx, login, password)
	if err != nil {
		s.logger.Errorf("Failed to register user: %v", err)
		return 0, err
	}
	return id, nil
}

// LoginUser аутентифицирует пользователя и возвращает JWT-токен
func (s *Service) LoginUser(ctx context.Context, login, password string) (string, error) {
	s.logger.Infof("Authenticating user with login: %s", login)
	userID, err := s.StorageImpl.CheckUser(ctx, login, password)
	if err != nil {
		s.logger.Errorf("Failed to check user: %v", err)
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	s.logger.Debugf("Generated token claims: %v", token.Claims)
	tokenString, err := token.SignedString([]byte(JWTKey))
	if err != nil {
		s.logger.Errorf("Failed to generate token: %v", err)
		return "", err
	}
	return tokenString, nil
}

// CreateAd создает новое объявление
func (s *Service) CreateAd(ctx context.Context, userID int, title, description, imageURL string, price float64) (int, error) {
	s.logger.Infof("Creating ad for user ID: %d", userID)
	adID, err := s.StorageImpl.CreateAd(ctx, userID, title, description, imageURL, price)
	if err != nil {
		s.logger.Errorf("Failed to create ad: %v", err)
		return 0, err
	}
	return adID, nil
}

// GetAds возвращает список объявлений
func (s *Service) GetAds(ctx context.Context, page, pageSize int, sortBy, sortOrder string, minPrice, maxPrice float64, userID int) ([]models.Ad, error) {
	s.logger.Infof("Fetching ads for page: %d, pageSize: %d", page, pageSize)
	ads, err := s.StorageImpl.GetAds(ctx, page, pageSize, sortBy, sortOrder, minPrice, maxPrice)
	if err != nil {
		s.logger.Errorf("Failed to get ads: %v", err)
		return nil, err
	}
	if userID != 0 {
		for i := range ads {
			isOwner, err := s.StorageImpl.IsAdOwner(ctx, ads[i].ID, userID)
			if err != nil {
				s.logger.Errorf("Failed to check ad owner for ad ID %d: %v", ads[i].ID, err)
				continue
			}
			ads[i].IsOwner = isOwner
		}
	}
	return ads, nil
}
