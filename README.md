# Go Homework 4 Blog API

A simple blog backend using Gin, GORM, SQLite, and JWT authentication.

## Requirements

- Go 1.20+ (recommended)

## Setup

```bash
go mod tidy
```

## Run

```bash
go run .
```

The server listens on `:8080`. The SQLite database file `blog.db` will be created in the project root.

## Web UI

Open `http://localhost:8080/` after starting the server. The UI is served from the `web` directory and lets you register, login, create posts, and comment.

## Environment

- `JWT_SECRET` (optional): secret key for signing JWT tokens. Defaults to `dev_secret_change_me`.

## API Overview

### Auth

- `POST /api/register`
- `POST /api/login`

### Posts

- `POST /api/posts` (auth)
- `GET /api/posts`
- `GET /api/posts/:id`
- `PUT /api/posts/:id` (auth, author only)
- `DELETE /api/posts/:id` (auth, author only)

### Comments

- `POST /api/posts/:id/comments` (auth)
- `GET /api/posts/:id/comments`

## Sample Requests

Register:

```bash
curl -X POST http://localhost:8080/api/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret","email":"alice@example.com"}'
```

Login:

```bash
curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret"}'
```

Create post (replace TOKEN):

```bash
curl -X POST http://localhost:8080/api/posts \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Hello","content":"First post"}'
```

List posts:

```bash
curl http://localhost:8080/api/posts
```

Create comment (replace TOKEN):

```bash
curl -X POST http://localhost:8080/api/posts/1/comments \
  -H 'Authorization: Bearer TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{"content":"Nice post"}'
```
