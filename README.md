# Lumea

## About
Lumea is a cosmetics brand of my friend who owns the business. The goal was to learn the basics of Go and help my friend with his business.

## Why I built it
I want to learn Go with hands-on practice, so it is more interesting than just theory. I like the idea of combining learning with helping my friend who has a business in cosmetics. So I built a REST API in Go for Lumea cosmetics.

## Features
- JWT auth with refresh tokens
- Products and variants
- Categories
- Orders with items and status updates
- Image upload to S3 storage (RustFS, an alternative to MinIO)

## Tech stack
- **Go**
- **chi**
- **sqlc + pgx**
- **goose** (for migrations)
- **RustFS** (docker image)
- **Docker Compose**

### Why these choices
I chose chi as a framework because it is widely used and I had already created a router with net/http, so I wanted to speed up some steps along the way. Also it has great built-in middleware. I chose sqlc + pgx because it also helps me speed up writing type-safe code. I wanted to focus more on creating a real Go REST API.

## Project structure
- `cmd/` - contains the files the app needs to run
- `internal/` - contains a domain-based structure. Each domain has repository, service, and handler files.
  - **Repository** is responsible for dealing with the database.
  - **Service** is responsible for business logic.
  - **Handler** is responsible for transport logic (more low-level stuff)

## Getting started

### Prerequisites
- Go v1.26
- Docker

### Steps
1. Copy `.env.example` to `.env`
2. Start the services `docker compose up`
3. Run the migrations (`goose up`)
4. Run the API (`go run ./cmd`)

## To-do list
Here I will list the things I need to finish.

- [ ] Tests (unit and integration)
- [ ] Atomic order creation with transactions and stock locking
- [ ] Context, timeouts, and graceful shutdown
- [ ] Structured logging and metrics
- [ ] Rate limiting, validation, and pagination
