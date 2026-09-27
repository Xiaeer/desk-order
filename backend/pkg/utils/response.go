package utils

import (
	"net/http"

	"deskorder/internal/dto/response"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response.Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func SuccessPage(c *gin.Context, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, response.Response{
		Code: 0,
		Msg:  "success",
		Data: response.PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, response.Response{
		Code: code,
		Msg:  msg,
	})
}

func FailWithStatus(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, response.Response{
		Code: code,
		Msg:  msg,
	})
}
