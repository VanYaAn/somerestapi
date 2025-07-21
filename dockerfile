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