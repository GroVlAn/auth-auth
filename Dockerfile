FROM golang:1.25.1 as builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o auth ./cmd/main.go 

FROM scratch
COPY --from=builder /app/auth /auth
COPY .env .
COPY ./configs ./configs
EXPOSE 9081 9011

ENTRYPOINT ["/auth"]