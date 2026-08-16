package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess       = 0
	CodeBadRequest    = 40000
	CodeNotFound      = 40400
	CodeConflict      = 40900
	CodeInternalError = 50000
)

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

func respondError(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
