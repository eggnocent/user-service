package services

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"
	"user-service/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repositories.IRepositoryRegistry
}

type IUserService interface {
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(context.Context, *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Update(context.Context, *dto.UpdateRequest, string) (*dto.UserResponse, error)
	GetUserLogin(ctx context.Context) (*dto.UserResponse, error)
	GetUserByUUD(ctx context.Context, uuid string) (*dto.UserResponse, error)
}

type Claims struct {
	User *dto.UserResponse
	jwt.RegisteredClaims
}

func NewUserService(repository repositories.IRepositoryRegistry) IUserService {
	return &UserService{repository: repository}
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.repository.GetUser().FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(time.Duration(config.Config.JwtExpirationTime) * time.Minute).Unix()
	data := &dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Role:        strings.ToLower(user.Role.Code),
	}

	claims := &Claims{
		User: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	response := &dto.LoginResponse{
		User:  *data,
		Token: tokenString,
	}

	return response, nil
}

func (u *UserService) isUsernameExist(ctx context.Context, username string) bool {
	user, err := u.repository.GetUser().FindByUsername(ctx, username)
	if err != nil {
		return false
	}

	if user != nil {
		return true
	}

	return false
}

func (u *UserService) isEmailExist(ctx context.Context, email string) bool {
	user, err := u.repository.GetUser().FindByEmail(ctx, email)
	if err != nil {
		return false
	}

	if user != nil {
		return true
	}

	return false
}

func (u *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if u.isUsernameExist(ctx, req.Username) {
		return nil, errConstant.ErrUsernameExist
	}

	if u.isEmailExist(ctx, req.Email) {
		return nil, errConstant.ErrEmailExist
	}

	if req.Password != req.ConfirmPassword {
		return nil, errConstant.ErrPasswordDoesNotMatch
	}

	user, err := u.repository.GetUser().Register(ctx, &dto.RegisterRequest{
		Name:        req.Name,
		Username:    req.Username,
		Password:    string(hashPassword),
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		RoleID:      uint(constants.Customer),
	})

	if err != nil {
		return nil, err
	}

	response := &dto.RegisterResponse{
		User: dto.UserResponse{
			UUID:        user.UUID,
			Name:        user.Name,
			Username:    user.Username,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		},
	}

	return response, nil
}

func (u *UserService) Update(ctx context.Context, req *dto.UpdateRequest, uuid string) (*dto.UserResponse, error) {
	var (
		password                  string
		checkUsername, checkEmail *models.User
		hashPassword              []byte
		user, userResult          *models.User
		err                       error
		data                      dto.UserResponse
	)

	user, err = u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	isUsernameExist := u.isUsernameExist(ctx, req.Username)
	if isUsernameExist && user.Username != req.Username {
		checkUsername, err = u.repository.GetUser().FindByUsername(ctx, req.Username)
		if err != nil {
			return nil, err
		}

		if checkUsername != nil {
			return nil, errConstant.ErrUsernameExist
		}
	}

	isEmailExist := u.isEmailExist(ctx, req.Email)
	if isEmailExist && user.Email != req.Email {
		checkEmail, err = u.repository.GetUser().FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}

		if checkEmail != nil {
			return nil, errConstant.ErrEmailExist
		}
	}

	if req.Password != nil {
		if *req.Password != *req.ConfirmPassword {
			return nil, errConstant.ErrPasswordDoesNotMatch
		}

		hashPassword, err = bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		password = string(hashPassword)
	}

	userResult, err = u.repository.GetUser().Update(ctx, &dto.UpdateRequest{
		Name:        req.Name,
		Username:    req.Username,
		Password:    &password,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}, uuid)

	if err != nil {
		return nil, err
	}

	data = dto.UserResponse{
		UUID:        userResult.UUID,
		Name:        userResult.Name,
		Username:    userResult.Username,
		PhoneNumber: userResult.PhoneNumber,
		Email:       userResult.Email,
	}

	return &data, nil
}

func (u *UserService) GetUserLogin(ctx context.Context) (*dto.UserResponse, error) {
	logrus.Info("📦 UserService.GetUserLogin called")

	userRaw := ctx.Value(constants.UserLogin)
	if userRaw == nil {
		logrus.Warn("⚠️ no user found in context (possible missing/misconfigured middleware)")
		return nil, fmt.Errorf("user not authenticated")
	}

	userLogin, ok := userRaw.(*dto.UserResponse)
	if !ok {
		logrus.Error("🔴 failed to cast context value to *dto.UserResponse")
		return nil, fmt.Errorf("invalid user data")
	}

	logrus.Infof("🔍 user found in context: uuid=%s, username=%s, role=%s", userLogin.UUID.String(), userLogin.Username, userLogin.Role)

	data := dto.UserResponse{
		UUID:        userLogin.UUID,
		Name:        userLogin.Name,
		Username:    userLogin.Username,
		PhoneNumber: userLogin.PhoneNumber,
		Email:       userLogin.Email,
		Role:        userLogin.Role,
	}

	return &data, nil
}

func (u *UserService) GetUserByUUD(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	user, err := u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	data := dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	return &data, nil
}
