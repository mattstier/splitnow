# --- build stage ---
FROM golang:1.25-alpine AS build
WORKDIR /src

# Copy Go modules first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
RUN CGO_ENABLED=0 go build -o /app/splitnow .

# --- runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates 
COPY --from=build /app/splitnow /usr/local/bin/splitnow
EXPOSE 8080
CMD ["splitnow"]
