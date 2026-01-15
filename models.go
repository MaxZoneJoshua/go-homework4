package main

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Email     string    `gorm:"unique;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Posts     []Post    `json:"-"`
	Comments  []Comment `json:"-"`
}

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"not null" json:"title"`
	Content   string    `gorm:"not null" json:"content"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"constraint:OnDelete:CASCADE" json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  []Comment `json:"-"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"not null" json:"content"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"constraint:OnDelete:CASCADE" json:"user"`
	PostID    uint      `json:"post_id"`
	Post      Post      `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
