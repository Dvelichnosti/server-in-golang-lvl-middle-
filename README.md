# Название не придумал, это что то типо Гонкорда (Goncord)

**Goncord** — это минималистичный, но мощный real-time WebSocket-сервер на Go с поддержкой JWT-авторизации, Redis-хранилища, Prometheus-метрик и Swagger-документации.

![Goncord](https://img.shields.io/badge/Golang-Realtime-blue?style=flat-square)

---

## 🚀 Возможности

- 📡 WebSocket-соединения с поддержкой комнат
- 🔐 JWT-авторизация
- 🧠 Redis для хранения последних 100 сообщений
- 📊 Prometheus метрики (`/metrics`)
- 📘 Swagger-документация (`/swagger/index.html`)
- 🧪 Минималистичный HTML-клиент

---

## ⚙️ Быстрый старт

### 1. Клонируй репозиторий
```bash
git clone https://github.com/yourname/goncord.git
cd goncord
```

### 2. Создай `go.mod`
```bash
go mod init goncord
```

### 3. Установи зависимости
```bash
go get github.com/gorilla/websocket

go get github.com/go-redis/redis/v8

go get github.com/golang-jwt/jwt/v5

go get github.com/prometheus/client_golang/prometheus

go get github.com/swaggo/swag/cmd/swag

go get github.com/swaggo/http-swagger

go get github.com/swaggo/files
```

### 4. Сгенерируй Swagger
```bash
swag init
```

После этого появится папка `./docs` с Swagger-файлами.

### 5. Запусти с Docker
```bash
make docker-up
```

Открой в браузере:
- http://localhost:8080/swagger/index.html — Swagger
- http://localhost:8080/metrics — Prometheus метрики
- `client.html` — HTML-интерфейс

---

## 📁 Структура проекта
```
.
├── main.go             # Основной сервер
├── client.html         # UI для WebSocket
├── docker-compose.yml  # Redis + сервер
├── Makefile            # Утилиты
├── docs/               # Swagger документация
└── README.md           # Это файл
```

---

## 🔐 Генерация JWT
Пример токена:
```go
import (
  "github.com/golang-jwt/jwt/v5"
  "time"
)

func generateToken(userID string) string {
  claims := jwt.MapClaims{
    "user_id": userID,
    "exp": time.Now().Add(time.Hour).Unix(),
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenStr, _ := token.SignedString([]byte("secret"))
  return tokenStr
}
```

---
## 📫 Контакты

**Разработчик:**

- GitHub: [@Dvelichnosti](https://github.com/Dvelichnosti)  
- Telegram: [@Dvelichnosti](https://t.me/Dve_lichnosti)
Если хочешь, я помогу развернуть проект в облаке или дописать `login` и `refresh-token` endpoints!
