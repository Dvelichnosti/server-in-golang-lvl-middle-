run:
	go run main.go

build:
	go build -o goncord main.go

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
