package controllers

import (
	controller "user-service/controllers/user"
	"user-service/services"
)

type Registry struct {
	service services.IServiceRegistry
}

type IControllerRegistry interface {
	GetUserController() controller.IUserController
}

func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}

func (u *Registry) GetUserController() controller.IUserController {
	return controller.NewUserController(u.service)
}
