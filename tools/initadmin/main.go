package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/anil-wu/spark-x/internal/config"
	"github.com/anil-wu/spark-x/internal/model"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("f", "etc/sparkx-api.yaml", "the config file")
	username   = flag.String("u", "admin", "admin username")
	password   = flag.String("p", "admin123", "admin password")
	role       = flag.String("r", "super_admin", "admin role (super_admin or admin)")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if c.MySQL.DSN == "" {
		log.Fatal("MySQL DSN is empty")
	}

	db, err := gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// check if admin already exists
	var existingAdmin model.Admins
	result := db.Where("username = ?", *username).First(&existingAdmin)
	if result.Error == nil {
		log.Fatalf("admin with username '%s' already exists", *username)
	}

	// md5 hash
	sum := md5.Sum([]byte(*password))
	passHash := hex.EncodeToString(sum[:])

	admin := &model.Admins{
		Username:     *username,
		PasswordHash: passHash,
		Role:         *role,
		Status:       "active",
	}

	result = db.Create(admin)
	if result.Error != nil {
		log.Fatalf("failed to create admin: %v", result.Error)
	}

	fmt.Printf("Admin created successfully!\n")
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Role: %s\n", *role)
}
