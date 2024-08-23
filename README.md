# gobudget-api

Inspired by YNAB API (because we love reinventing the wheel, for learning purposes of course).

Work in progress.

## Configuration

`app.env`:

```config
DB_CONNSTRING=postgresql://app:Supers3cret@localhost:5432/appdb?sslmode=disable
DB_MIGRATION_FILES=file://db/migration
APP_URL=http://localhost:8080/beta
ENVIRONMENT=local
LISTEN_ADDR=0.0.0.0
LISTEN_PORT=8080
SECRET_KEY=SuperS3cretJwtKey4DevelopmentUs@ge
ACCESS_TOKEN_DURATION=60m
REFRESH_TOKEN_DURATION=24h
REDIS_ADDRESS=127.0.0.1:6379
EMAIL_SENDER_NAME=
GMAIL_SENDER_ADDRESS=
GMAIL_SENDER_PASSWORD=
MAILHOG_HOST=localhost:1025
MAILHOG_SENDER_ADDRESS=gobudgetapi@localdomain.lcl
```

## Developer setup

Install Docker.

```bash
sudo apt install build-essential

go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install go.uber.org/mock/mockgen@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```
