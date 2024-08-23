FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . /app/

RUN go build -o gobudgetapi ./

FROM alpine:latest
LABEL org.opencontainers.image.source https://github.com/guerzon/gobudget-api
LABEL org.opencontainers.image.description "gobudget api"

RUN addgroup app && adduser -S app app

WORKDIR /app
COPY --from=builder /app/gobudgetapi .
COPY db/migration ./db/migration

RUN chown -R app:app /app
USER app

EXPOSE 8080
ENTRYPOINT [ "/app/gobudgetapi" ]