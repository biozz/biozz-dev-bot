FROM golang:1.24-alpine AS backend
WORKDIR /app/
COPY . .
RUN CGO_ENABLED=0 go build -o bin/app main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/app-arm64 .

FROM alpine:3.19
WORKDIR /app/
COPY --from=backend /app/bin/app .
COPY --from=backend /app/bin/app-arm64 .
COPY entrypoint.sh .

RUN chmod +x /app/entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["serve", "--http=0.0.0.0:8080", "--dev"]
