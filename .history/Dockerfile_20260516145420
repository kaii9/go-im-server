# ===== Stage 1: Build frontend =====
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ .
RUN npm run build

# ===== Stage 2: Build Go backend =====
FROM golang:1.25-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o im-server .

# ===== Stage 3: Final image =====
FROM alpine:3.19
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
COPY --from=backend /app/im-server .
COPY --from=frontend /app/web/dist ./web/dist
COPY config.yaml .

EXPOSE 8080

CMD ["./im-server"]
