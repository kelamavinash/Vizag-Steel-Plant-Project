package main

import (
	"fmt"
	"vizag-steel-delay-system/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("delay_system.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	var users []models.User
	db.Find(&users)
	for _, u := range users {
		fmt.Printf("Emp: %s, Role: %s\n", u.EmpNo, u.Role)
	}
}
