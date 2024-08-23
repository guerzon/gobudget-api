build:
	go build -o ./bin/gobudget-api ./

run: build
	./bin/gobudget-api

test:
	go test -v -cover -short ./... -count=1

migrateup:
	migrate -path db/migration --database "postgresql://app:Supers3cret@localhost:5432/appdb?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration --database "postgresql://app:Supers3cret@localhost:5432/appdb?sslmode=disable" -verbose down

migrateup1:
	migrate -path db/migration --database "postgresql://app:Supers3cret@localhost:5432/appdb?sslmode=disable" -verbose up 1

migratedown1:
	migrate -path db/migration --database "postgresql://app:Supers3cret@localhost:5432/appdb?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate -f ./db/sqlc.yaml

mock:
	mockgen -package mockdb -destination pkg/mock/store.go github.com/guerzon/gobudget-api/pkg/db Store
	mockgen -package mockdb -destination pkg/mock/worker.go github.com/guerzon/gobudget-api/pkg/worker TaskDistributor

docs:
	swag fmt && swag init --parseDependency

sec:
	gosec -exclude=G404 ./...

.PHONY: build run migrateup migratedown migrateup1 migratedown1 sqlc mock docs sec
