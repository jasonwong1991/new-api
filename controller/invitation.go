package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetAllInvitationCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	codes, total, err := model.GetAllInvitationCodes(pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func SearchInvitationCodes(c *gin.Context) {
	keyword := c.Query("keyword")
	pageInfo := common.GetPageQuery(c)
	codes, total, err := model.SearchInvitationCodes(keyword, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

type CreateInvitationCodeRequest struct {
	Name        string `json:"name"`
	Count       int    `json:"count"`
	MaxUses     int    `json:"max_uses"`
	ExpiredTime int64  `json:"expired_time"`
}

func CreateInvitationCode(c *gin.Context) {
	var req CreateInvitationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "一次最多生成 100 个邀请码",
		})
		return
	}

	createdBy := c.GetInt("id")

	// Default to single use if not specified
	if req.MaxUses == 0 {
		req.MaxUses = 1
	}

	// Default to never expire if not specified
	if req.ExpiredTime == 0 {
		req.ExpiredTime = -1
	}

	codes, err := model.BatchCreateInvitationCodes(req.Name, req.Count, req.MaxUses, req.ExpiredTime, createdBy)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    codes,
	})
}

func UpdateInvitationCode(c *gin.Context) {
	var code model.InvitationCode
	if err := c.ShouldBindJSON(&code); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if code.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邀请码 ID 不能为空",
		})
		return
	}

	if err := code.Update(); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func DeleteInvitationCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if err := model.DeleteInvitationCode(id); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func GetInvitationCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	code, err := model.GetInvitationCodeById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    code,
	})
}

// ValidateInvitationCode is a public endpoint to check if an invitation code is valid
func ValidateInvitationCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邀请码不能为空",
		})
		return
	}

	if err := model.CheckInvitationCode(code); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
