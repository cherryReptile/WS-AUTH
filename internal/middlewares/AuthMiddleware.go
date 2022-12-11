package middlewares

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pavel-one/GoStarter/api"
	"github.com/pavel-one/GoStarter/internal/helpers"
	"net/http"
	"strings"
)

func CheckAuthHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := helpers.CheckAuthHeader(c)
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth header is missing or value is required"})
			return
		}
		c.Set("token", token)
		c.Next()
	}
}

func CheckUserAndToken(checkAuthService api.CheckAuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := helpers.GetAndCastToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		res, err := checkAuthService.CheckAuth(context.Background(), &api.TokenRequest{Token: token})
		if err != nil || !res.Ok {
			e := strings.Split(err.Error(), "=")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": e[2]})
			return
		}

		//c.Next()
	}
}
