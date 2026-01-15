package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() error {
	conn, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := conn.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		return err
	}

	db = conn
	log.Println("database initialized")
	return nil
}
