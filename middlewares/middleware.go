package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"user-service/common/response"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	services "user-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

func HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Recovered from panic: %v", r)
				c.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
// 		if err != nil {
// 			c.JSON(http.StatusTooManyRequests, response.Response{
// 				Status:  constants.Error,
// 				Message: errConstant.ErrToManyRequests.Error(),
// 			})
// 			c.Abort()
// 		}
// 		c.Next()
// 	}
// }

func extractBearerToken(token string) string {
	arrayToken := strings.Split(token, " ")
	if len(arrayToken) == 2 {
		return arrayToken[1]
	}
	return ""
}

func responseUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	c.Abort()
}

func validateAPIKey(c *gin.Context) error {
	apiKey := c.GetHeader(constants.XApiKey)
	requestAt := c.GetHeader(constants.XRequestAt)
	serviceName := c.GetHeader(constants.XServiceName)
	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)
	hash := sha256.New()
	hash.Write([]byte(validateKey))
	resultHash := hex.EncodeToString(hash.Sum(nil))

	// 🔍 Tambahkan logrus:
	logrus.Infof("🔐 [validateAPIKey] serviceName: %s", serviceName)
	logrus.Infof("🔐 [validateAPIKey] signatureKey: %s", signatureKey)
	logrus.Infof("🔐 [validateAPIKey] requestAt: %s", requestAt)
	logrus.Infof("🔐 [validateAPIKey] raw string to hash: %s", validateKey)
	logrus.Infof("🔐 [validateAPIKey] expected hash: %s", resultHash)
	logrus.Infof("🔐 [validateAPIKey] received apiKey: %s", apiKey)

	if apiKey != resultHash {
		logrus.Warn("❌ [validateAPIKey] Invalid API Key")
		return errConstant.ErrUnauthorized
	}

	logrus.Info("✅ [validateAPIKey] API Key valid")
	return nil
}

func validateBearerToken(c *gin.Context, token string) error {
	logrus.Infof("🔐 [validateBearerToken] raw Authorization header: %s", token)

	if !strings.Contains(token, "Bearer") {
		logrus.Warn("❌ [validateBearerToken] Authorization header does not contain 'Bearer'")
		return errConstant.ErrUnauthorized
	}

	tokenString := extractBearerToken(token)
	logrus.Infof("🔐 [validateBearerToken] extracted token: %s", tokenString)

	if tokenString == "" {
		logrus.Warn("❌ [validateBearerToken] Token is empty after extraction")
		return errConstant.ErrUnauthorized
	}

	claims := &services.Claims{}
	tokenJwt, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			logrus.Warn("❌ [validateBearerToken] Invalid signing method")
			return nil, errConstant.ErrInvalidToken
		}

		jwtSecret := []byte(config.Config.JwtSecretKey)
		logrus.Infof("🔐 [validateBearerToken] Using jwtSecret: %s", string(jwtSecret))
		return jwtSecret, nil
	})

	if err != nil {
		logrus.Warnf("❌ [validateBearerToken] Token parsing error: %v", err)
		return errConstant.ErrUnauthorized
	}

	if !tokenJwt.Valid {
		logrus.Warn("❌ [validateBearerToken] Token is not valid")
		return errConstant.ErrUnauthorized
	}

	logrus.Infof("✅ [validateBearerToken] Token valid for user: %+v", claims.User)
	userLogin := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.UserLogin, claims.User))
	c.Request = userLogin
	c.Set(constants.Token, token)
	return nil
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		token := c.GetHeader(constants.Authorization)
		logrus.Infof("🔐 [Authenticate] Authorization header: %s", token)

		if token == "" {
			logrus.Warn("❌ [Authenticate] Missing Authorization header")
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		err = validateBearerToken(c, token)
		if err != nil {
			logrus.Warnf("❌ [Authenticate] validateBearerToken failed: %v", err)
			responseUnauthorized(c, err.Error())
			return
		}

		err = validateAPIKey(c)
		if err != nil {
			logrus.Warnf("❌ [Authenticate] validateAPIKey failed: %v", err)
			responseUnauthorized(c, err.Error())
			return
		}

		logrus.Info("✅ [Authenticate] Auth success")
		c.Next()
	}
}
