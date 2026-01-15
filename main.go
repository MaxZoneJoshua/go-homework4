package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})
	router.StaticFile("/app.js", "./web/app.js")
	router.Static("/web", "./web")
	router.StaticFile("/test", "./web/test.txt")

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	api.POST("/register", registerHandler)
	api.POST("/login", loginHandler)
	api.GET("/posts", listPostsHandler)
	api.GET("/posts/:id", getPostHandler)
	api.GET("/posts/:id/comments", listCommentsHandler)

	auth := api.Group("")
	auth.Use(authMiddleware(), gin.Logger(), gin.Recovery())
	auth.POST("/posts", createPostHandler)
	auth.PUT("/posts/:id", updatePostHandler)
	auth.DELETE("/posts/:id", deletePostHandler)
	auth.POST("/posts/:id/comments", createCommentHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
