package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/anil-wu/spark-x/internal/config"
	"github.com/anil-wu/spark-x/internal/model"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("f", "etc/sparkx-api.yaml", "the config file")
	email      = flag.String("u", "admin@example.com", "admin email")
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

	// md5 hash
	sum := md5.Sum([]byte(*password))
	passHash := hex.EncodeToString(sum[:])

	username := *email
	if at := strings.Index(username, "@"); at > 0 {
		username = username[:at]
	}

	isSuper := *role == "super_admin"

	if err := db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Raw(
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'is_super'",
		).Scan(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Exec("ALTER TABLE `users` ADD COLUMN `is_super` TINYINT(1) NOT NULL DEFAULT 0").Error; err != nil {
				return err
			}
		}

		if isSuper {
			if err := tx.Model(&model.Users{}).Where("is_super = ?", true).Update("is_super", false).Error; err != nil {
				return err
			}
		}

		var existingUser model.Users
		findErr := tx.Where("email = ?", *email).First(&existingUser).Error
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}

		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return tx.Create(&model.Users{
				Username:     username,
				Email:        *email,
				PasswordHash: passHash,
				IsSuper:      isSuper,
			}).Error
		}

		return tx.Model(&model.Users{}).Where("id = ?", existingUser.Id).Updates(map[string]interface{}{
			"username":      username,
			"password_hash": passHash,
			"is_super":      isSuper,
		}).Error
	}); err != nil {
		log.Fatalf("failed to ensure user: %v", err)
	}

	fmt.Printf("Admin created successfully!\n")
	fmt.Printf("Email: %s\n", *email)
	fmt.Printf("Role: %s\n", *role)
}
