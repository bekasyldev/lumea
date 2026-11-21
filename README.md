# Lumea

## About
Lumea is a cosmetics brand of my friend who owns the business. The goal was to learn the basics of Go and help my friend with his business.

## Why I built it
I want to learn Go with hands-on practice, so it is more interesting than just theory. I like the idea of combining learning with helping my friend who has a business in cosmetics. So I built a REST API in Go for Lumea cosmetics.

## API docs
Bruno based API documentation: https://github.com/bekasyldev/lumea_api_docs

## Features

### Domain
- Products and variants
- Categories
- Orders with items and status updates
- Atomic order creation with stock locking; stock is returned when an order is deleted or cancelled
- Image upload to S3 storage (RustFS, an alternative to MinIO)

### Platform
- JWT auth with refresh tokens and logout (revoked via Redis)
- Role-based access: user and admin
- Token bucket based rate limiting: 100 capacity, 10/s per IP, returns 429 with a `Retry-After` header
- Cursor-based pagination on products
- Prometheus metrics at `/metrics` (admin only): request duration, request count, orders created
- Graceful shutdown with a 15s drain

## Tech stack
- **Go**
- **chi**
- **sqlc + pgx**
- **goose** (for migrations)
- **RustFS** (docker image)
- **Redis**
- **Docker Compose**

### Why these choices
I chose chi as a framework because it is widely used and I had already created a router before with net/http, so I wanted to speed up some steps along the way. Also it has great built-in middleware. I chose sqlc + pgx because it also helps me speed up writing type-safe code. I added Redis for storing refresh tokens so I can revoke them any time, like in production. I wanted to focus more on creating a real Go REST API.

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
4. Set env vars by running `set -a; source .env; set +a`
5. Run the API (`go run ./cmd`)

## Roadmap and what I learned
Everything I planned to build, plus notes on what I learned along the way.

- [x] Tests (unit and integration)
- [x] Atomic order creation with transactions and stock locking
- [x] Context, timeouts, and graceful shutdown
- [x] Structured logging and metrics
    - Understand the difference between metric types and when to use them
    - Create several custom metrics and monitor them
    - Run Prometheus locally and work with it
- [x] Rate limiting, validation, and pagination
    - Understand the difference between cursor and offset based pagination
    - [x] Implement cursor based pagination

      I chose cursor-based pagination because it is more complex and widely used in modern APIs (especially public ones). I know offset-based pagination would suit my application better and that this is a bit of overengineering, but I wanted to learn it.
    - Implement token bucket algorithm
- [x] Review the code with an AI agent and work on bugs
    - [x] There is no authorization, only authentication

      Authorization checks what access a user has, usually based on roles. The simple solution is to wrap the routes in a router group and protect it with an admin-required middleware.
    - [x] Write one integration test against CreateOrder to test the race condition
    - [x] Refresh token works as an access token
    - [x] DeleteOrder should return the stock and run in a transaction (like DeleteOrderItem)
    - [x] Fix validation of the /register endpoint, so an empty username and password can't create a user
