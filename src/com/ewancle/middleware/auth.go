package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/utils/jwt"
)

// JWTAuth 中间件
func JWTAuth(requiredRoles ...string) gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer {token}"})
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 保存用户信息到上下文
		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles)

		// 权限检查
		if len(requiredRoles) > 0 {
			hasRole := false
			for _, r := range claims.Roles {
				for _, rr := range requiredRoles {
					if r == rr {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
