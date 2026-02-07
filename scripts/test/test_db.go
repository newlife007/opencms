package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Files struct {
	ID       uint   `gorm:"primarykey"`
	Title    string `gorm:"column:title"`
	Name     string `gorm:"column:name"`
	Type     int    `gorm:"column:type"`
	Status   int    `gorm:"column:status"`
}

func (Files) TableName() string {
	return "ow_files"
}

func main() {
	dsn := "openwan:openwan123@tcp(172.21.0.2:3306)/openwan_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var count int64
	if err := db.Model(&Files{}).Count(&count).Error; err != nil {
		log.Fatal("Failed to count files:", err)
	}
	fmt.Printf("Total files: %d\n", count)

	var files []Files
	if err := db.Limit(5).Find(&files).Error; err != nil {
		log.Fatal("Failed to query files:", err)
	}

	fmt.Printf("First 5 files:\n")
	for _, f := range files {
		fmt.Printf("  ID=%d, Title=%s, Name=%s, Type=%d, Status=%d\n",
			f.ID, f.Title, f.Name, f.Type, f.Status)
	}
}
