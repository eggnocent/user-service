package seeders

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"user-service/domain/models"
)

func RunRoleSeeder(db *gorm.DB) {
	roles := []models.Role{
		{
			Code: "ADMIN",
			Name: "Administrator",
		},
		{
			Code: "CUSTOMER",
			Name: "Customer",
		},
	}

	for _, role := range roles {
		err := db.FirstOrCreate(&role, models.Role{Code: role.Code}).Error
		if err != nil {
			logrus.Errorf("Failed to create role: %v", err)
			panic(err)
		}
		logrus.Infof("Created role: %v successfuly", role)
	}
}
