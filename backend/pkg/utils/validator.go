package utils

import (
	"github.com/gin-gonic/gin"
)

// BindJSON 绑定 JSON 请求体并校验，失败时自动响应错误
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return false
	}
	return true
}

// BindQuery 绑定 Query 参数并校验
func BindQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return false
	}
	return true
}
