# Go Homework 4 Blog API

简单的博客API实现，使用 Gin, GORM, SQLite, 和 JWT authentication.

## 依赖 

- Go 1.20+ (recommended)

## 配置

```bash
go mod tidy
```

## 执行

```bash
go run .
```

The server listens on `:8080`. The SQLite database file `blog.db` will be created in the project root.

## Web UI

Open `http://localhost:8080/` after starting the server. The UI is served from the `web` directory and lets you register, login, create posts, and comment.

## 环境

- `JWT_SECRET` (optional): secret key for signing JWT tokens. Defaults to `dev_secret_change_me`.

## API 概览

### 验证

- `POST /api/register`
- `POST /api/login`

### 帖子

- `POST /api/posts` (auth)
- `GET /api/posts`
- `GET /api/posts/:id`
- `PUT /api/posts/:id` (auth, author only)
- `DELETE /api/posts/:id` (auth, author only)

### 评论

- `POST /api/posts/:id/comments` (auth)
- `GET /api/posts/:id/comments`

## 请求示例

注册:

```bash
curl -X POST http://localhost:8080/api/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret","email":"alice@example.com"}'
```

登陆:

```bash
curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret"}'
```

创建帖子 (replace TOKEN):

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
