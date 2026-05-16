package middleware

import (
	"net/http"
	"strings"

	"go-im-server/common"
	"go-im-server/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.ErrUnauthorized, common.GetErrMsg(common.ErrUnauthorized))
			c.Abort()
			return
		}

		c.Set("uid", int64(uidFloat))
		c.Next()
	}
}

func ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return 0, err
	}
	return int64(uidFloat), nil
}
