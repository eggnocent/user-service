package seeders

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user-service/domain/models"
)

func RunUserSeeder(db *gorm.DB) {
	password, _ := bcrypt.GenerateFromPassword([]byte("p@ssw0rd"), bcrypt.DefaultCost)
	user := models.User{
		UUID:        uuid.New(),
		Name:        "Administrator",
		Username:    "admin",
		Password:    string(password),
		PhoneNumber: "0987654321",
		Email:       "admin@gmail.com",
		RoleID:      1,
	}

	err := db.FirstOrCreate(&user, &models.User{Username: user.Username}).Error
	if err != nil {
		logrus.Errorf("Failed to create user: %v", err)
		panic(err)
	}

	logrus.Infof("Created user: %v successfully", user)
}
