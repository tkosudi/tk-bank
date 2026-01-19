PATH := $(PATH):/home/tkosudi/.asdf/installs/golang/1.25.3/bin

postgres:
	docker run --name postgres-16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	docker exec -it postgres-16 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres-16 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc: 
	sqlc generate

test:
	go test -v -cover ./...

resetdb: dropdb createdb migrateup

server: 
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server resetdb