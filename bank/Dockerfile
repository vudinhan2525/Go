# Stage 1 : Build the app
FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY . . 
RUN go build -o main main.go
# Stage 2 :  Run app
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .

EXPOSE 5000
CMD [ "/app/main" ] 