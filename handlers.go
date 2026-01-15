package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type postRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type commentRequest struct {
	Content string `json:"content" binding:"required"`
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func respondJSON(c *gin.Context, status int, payload interface{}) {
	c.JSON(status, payload)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	value := c.Param(name)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func registerHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var existing User
	err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existing).Error
	if err == nil {
		respondError(c, http.StatusConflict, "username or email already exists")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("register lookup failed: %v", err)
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}

	hashed, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("hash password failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := User{
		Username: req.Username,
		Password: hashed,
		Email:    req.Email,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("create user failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{"message": "user registered"})
}

func loginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusUnauthorized, "invalid username or password")
			return
		}
		log.Printf("login lookup failed: %v", err)
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}

	if err := checkPassword(user.Password, req.Password); err != nil {
		respondError(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := generateToken(user)
	if err != nil {
		log.Printf("token generation failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(c, http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func createPostHandler(c *gin.Context) {
	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	post := Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
	}

	if err := db.Create(&post).Error; err != nil {
		log.Printf("create post failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to create post")
		return
	}

	respondJSON(c, http.StatusCreated, post)
}

func listPostsHandler(c *gin.Context) {
	query := db.Preload("User").Order("created_at desc")

	if limit := c.Query("limit"); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err != nil || parsed <= 0 {
			respondError(c, http.StatusBadRequest, "invalid limit")
			return
		}
		query = query.Limit(parsed)
	}

	if offset := c.Query("offset"); offset != "" {
		parsed, err := strconv.Atoi(offset)
		if err != nil || parsed < 0 {
			respondError(c, http.StatusBadRequest, "invalid offset")
			return
		}
		query = query.Offset(parsed)
	}

	var posts []Post
	if err := query.Find(&posts).Error; err != nil {
		log.Printf("list posts failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch posts")
		return
	}

	respondJSON(c, http.StatusOK, posts)
}

func getPostHandler(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid post id")
		return
	}

	var post Post
	if err := db.Preload("User").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	respondJSON(c, http.StatusOK, post)
}

func updatePostHandler(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid post id")
		return
	}

	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post for update failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	userID := c.MustGet("userID").(uint)
	if post.UserID != userID {
		respondError(c, http.StatusForbidden, "not the post author")
		return
	}

	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	post.Title = req.Title
	post.Content = req.Content
	if err := db.Save(&post).Error; err != nil {
		log.Printf("update post failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to update post")
		return
	}

	respondJSON(c, http.StatusOK, post)
}

func deletePostHandler(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid post id")
		return
	}

	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post for delete failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	userID := c.MustGet("userID").(uint)
	if post.UserID != userID {
		respondError(c, http.StatusForbidden, "not the post author")
		return
	}

	if err := db.Delete(&post).Error; err != nil {
		log.Printf("delete post failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to delete post")
		return
	}

	respondJSON(c, http.StatusOK, gin.H{"message": "post deleted"})
}

func createCommentHandler(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid post id")
		return
	}

	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post for comment failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	var req commentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	comment := Comment{
		Content: req.Content,
		UserID:  userID,
		PostID:  post.ID,
	}

	if err := db.Create(&comment).Error; err != nil {
		log.Printf("create comment failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to create comment")
		return
	}

	respondJSON(c, http.StatusCreated, comment)
}

func listCommentsHandler(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid post id")
		return
	}

	if err := db.First(&Post{}, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post for comments failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch post")
		return
	}

	var comments []Comment
	if err := db.Preload("User").Where("post_id = ?", postID).Order("created_at asc").Find(&comments).Error; err != nil {
		log.Printf("list comments failed: %v", err)
		respondError(c, http.StatusInternalServerError, "failed to fetch comments")
		return
	}

	respondJSON(c, http.StatusOK, comments)
}
