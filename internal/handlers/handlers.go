package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"restapi/internal/service"
	"strconv"
	"strings"
	"time"
)

var defaultTimeForCancel = 5

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AdRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type AdResponse struct {
	ID int `json:"id"`
}

// Handler представляет структуру для обработки HTTP-запросов
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterHandler обрабатывает регистрацию пользователя
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}
	if len(req.Login) < 3 || len(req.Login) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Login must be between 3 and 50 characters"})
		return
	}
	if len(req.Password) < 6 || len(req.Password) > 255 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Password must be between 6 and 255 characters"})
		return
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, req.Login); !matched {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Login must contain only letters, numbers, or underscores"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(defaultTimeForCancel)*time.Second)
	defer cancel()
	userID, err := h.svc.RegisterUser(ctx, req.Login, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "login already exists") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Login already exists"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "User created successfully", "user_id": userID})
}

// LoginHandler обрабатывает аутентификацию пользователя
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Login and password are required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(defaultTimeForCancel)*time.Second)
	defer cancel()
	token, err := h.svc.LoginUser(ctx, req.Login, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid login or password"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

// CreateAdHandler обрабатывает создание объявления
func (h *Handler) CreateAdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not authenticated"})
		return
	}
	var req AdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}
	if len(req.Title) < 3 || len(req.Title) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Title must be between 3 and 100 characters"})
		return
	}
	if len(req.Description) < 10 || len(req.Description) > 1000 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Description must be between 10 and 1000 characters"})
		return
	}
	if len(req.ImageURL) > 255 || !strings.HasPrefix(req.ImageURL, "http") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid image URL"})
		return
	}
	if req.Price < 0 || req.Price > 1000000 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Price must be between 0 and 1,000,000"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(defaultTimeForCancel)*time.Second)
	defer cancel()
	adID, err := h.svc.CreateAd(ctx, userID, req.Title, req.Description, req.ImageURL, req.Price)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create ad"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AdResponse{ID: adID})
}

// GetAdsHandler обрабатывает получение списка объявлений
func (h *Handler) GetAdsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page := 1
	pageSize := 10
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	minPrice := 0.0
	maxPrice := 1000000.0
	if p := r.URL.Query().Get("page"); p != "" {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 {
			pageSize = psInt
		}
	}
	if mp := r.URL.Query().Get("min_price"); mp != "" {
		if mpFloat, err := strconv.ParseFloat(mp, 64); err == nil {
			minPrice = mpFloat
		}
	}
	if mp := r.URL.Query().Get("max_price"); mp != "" {
		if mpFloat, err := strconv.ParseFloat(mp, 64); err == nil {
			maxPrice = mpFloat
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(defaultTimeForCancel)*time.Second)
	defer cancel()
	userID, _ := r.Context().Value("user_id").(int)
	ads, err := h.svc.GetAds(ctx, page, pageSize, sortBy, sortOrder, minPrice, maxPrice, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get ads"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ads)
}
