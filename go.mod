go mod init goncord


Добавьте зависимости

go get github.com/gorilla/websocket
go get github.com/go-redis/redis/v8
go get github.com/golang-jwt/jwt/v5
go get github.com/prometheus/client_golang/prometheus
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/http-swagger
go get github.com/swaggo/files


Полный go.mod после этого будет примерно таким:

module goncord

go 1.20

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/gorilla/websocket v1.5.1
	github.com/prometheus/client_golang v1.17.0
	github.com/swaggo/files v0.0.0-20240114170216-2196fb80fa64
	github.com/swaggo/http-swagger v1.4.5
	github.com/swaggo/swag v1.16.2
)
