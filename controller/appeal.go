package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetBanList(c *gin.Context) {
	users, err := model.GetBannedUsers()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

type SubmitAppealRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=1000"`
}

func SubmitAppeal(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if user.Status != common.UserStatusDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "您的账户未被封禁，无需申诉",
		})
		return
	}

	if model.HasPendingAppeal(userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "您已有待处理的申诉，请等待管理员审核",
		})
		return
	}

	var req SubmitAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "申诉理由需要10-1000个字符",
		})
		return
	}

	appeal := &model.Appeal{
		UserId: userId,
		Reason: req.Reason,
	}
	if err := appeal.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "申诉已提交，请等待管理员审核",
	})
}

func GetMyAppeals(c *gin.Context) {
	userId := c.GetInt("id")
	appeals, err := model.GetAppealsByUserId(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    appeals,
	})
}

func GetAllAppeals(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	appeals, total, err := model.GetAllAppeals(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": appeals,
			"total": total,
			"page":  page,
		},
	})
}

type ReviewAppealRequest struct {
	Note string `json:"note"`
}

func ApproveAppeal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的申诉ID",
		})
		return
	}

	var req ReviewAppealRequest
	c.ShouldBindJSON(&req)

	adminId := c.GetInt("id")
	if err := model.ApproveAppeal(id, adminId, req.Note); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "申诉已通过，用户已解封",
	})
}

func RejectAppeal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的申诉ID",
		})
		return
	}

	var req ReviewAppealRequest
	c.ShouldBindJSON(&req)

	adminId := c.GetInt("id")
	if err := model.RejectAppeal(id, adminId, req.Note); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "申诉已拒绝",
	})
}
