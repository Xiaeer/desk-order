package middleware

import (
	"net/http"
	"strings"

	"deskorder/pkg/auth"
	"deskorder/pkg/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 鉴权中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.FailWithStatus(c, http.StatusUnauthorized, 401, "未登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.FailWithStatus(c, http.StatusUnauthorized, 401, "Token 格式错误")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			utils.FailWithStatus(c, http.StatusUnauthorized, 401, "Token 无效或已过期")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RoleAuth 角色权限中间件，限制只有指定角色才能访问
func RoleAuth(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.FailWithStatus(c, http.StatusForbidden, 403, "无权限")
			c.Abort()
			return
		}
		roleStr, _ := role.(string)
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}
		utils.FailWithStatus(c, http.StatusForbidden, 403, "无权限")
		c.Abort()
	}
}
