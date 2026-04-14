package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

func GetDynamicRatio(c *gin.Context) {
	info := setting.GetDynamicRatioInfo()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}
