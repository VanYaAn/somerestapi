## REST API

## Задание
**Написать приложение, реализующее REST-API условного маркетплейса**:


 -авторизация пользователей
 
 -регистрация пользователей

 -размещение нового объявления

  -отображение ленты объявлений.

 ---
## Требования
API должен быть реализован в формате REST + JSON. Реализовать приложение на языке Go. Допускается использование фреймворков. База данных любая.
Приложение, включая БД, должно быть упаковано в Docker контейнер.
---

## Запуск 

make docker_run

---
##Реализация

Приложение (Go приложение), БД (PostgreSQL) и Docker-compose файл, для управления контейнерами.
```yaml

services:
  db:
    image: postgres:latest
    restart: always
    environment:
      POSTGRES_USER: your_username
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: your_database_name
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - JWT_KEY=vk_please_give_me_internship
      - DATABASE_HOST=db
      - DATABASE_PORT=5432
      - DATABASE_USERNAME=your_username
      - DATABASE_PASSWORD=your_password
      - DATABASE_NAME=your_database_name
    depends_on:
      - db
    restart: always

volumes:
  db_data:
```

```Dockerfile
FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /restapi ./cmd/api/main.go

FROM alpine:latest


RUN apk --no-cache add ca-certificates

COPY --from=builder /restapi /restapi

COPY config.yaml /config.yaml

EXPOSE 8080

ENV JWT_KEY=vk_please_give_me_internship

CMD ["/restapi"]
```
Есть 4 endpoint'а:
  -"api/v1/register" (POST)
  -"api/v1/login" (POST)
  -"api/v1/ads" (GET)
  -"api/v1/ads" (POST)


Использовал классическую библиотеку для роутингка gorila/mux.
```golang
	r := mux.NewRouter()

	h := handlers.NewHandler(svc)
	//не нужен jwt token
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/register", h.RegisterHandler).Methods("POST")
	public.HandleFunc("/login", h.LoginHandler).Methods("POST")
	public.HandleFunc("/ads", h.GetAdsHandler).Methods("GET")

	//нужен jwt token
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(logger, JWTKey))
	protected.HandleFunc("/ads", h.CreateAdHandler).Methods("POST")
```
## /api/v1/register
request 


POST 
```json
{"login":"examplename","password":"examplepassword"}
```
response
```json{
    "message": "User created successfully",
    "user_id": 12
    }
```
если еще раз делать запрос 


request 

POST 
```json
{"login":"examplename","password":"examplepassword"}
```
response 
```json
{
    "error": "Failed to create user"
}
```

## /api/v1/login

request


POST 
```json
{"login":"examplename","password":"examplepassword"}
```
response
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTMxOTU1MDIsInVzZXJfaWQiOjEyfQ.iLzuXy9Y0X9q-FldhnIZhZcm721btj2mU6fK61D7UyY"
}
```

Использовал JWT для авторизации. Надо передать его в заголовки(header) запроса, чтобы он проходил AuthMiddleware 
```go
func AuthMiddleware(logger *zap.SugaredLogger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header missing"})
				return
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.Errorf("Unexpected signing method: %v", token.Header["alg"])
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token claims"})
				return
			}
			userID, ok := claims["user_id"].(float64)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID in token"})
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", int(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

### /api/v1/ads

В загаловке лежит JWT токен.

request


POST 
```json
{
    "title": "Новая доска",
    "description": "новая доска для школы 2м х 2м",
    "image_url": "http://example.com/image.jpg",
    "price": 1337
}
```
response
```json
{
    "id": 4
}
```

Если это будет GET-запрос,то есть параметры:page,page_size,min_price,max_price,sort_by,sort_order.

```json
[
    {
        "id": 4,
        "title": "Новая доска",
        "description": "новая доска для школы 2м х 2м",
        "image_url": "http://example.com/image.jpg",
        "price": 1337,
        "user_id": 11,
        "login": "dsfsd",
        "created_at": "2025-07-21T14:51:01Z"
    },
    {
        "id": 3,
        "title": "Новый телефон",
        "description": "Смартфон в отличном состоянии",
        "image_url": "http://example.com/image.jpg",
        "price": 500,
        "user_id": 6,
        "login": "limon",
        "created_at": "2025-07-19T22:58:34Z"
    },
  ...
]
```
Если нет параметров в URL, то применяется сортрировка по времени(самые новые в начале).

##Migration
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS ads (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL CHECK (char_length(title) >= 3),
    description TEXT NOT NULL CHECK (char_length(description) >= 10),
    image_url VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2)  NOT NULL CHECK (price >= 0),
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```


email: ivanantoshin176@gmail.com
