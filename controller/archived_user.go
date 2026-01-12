package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type CleanupPreviewResponse struct {
	Count   int64 `json:"count"`
	MinDays int   `json:"min_days"`
	StartId int   `json:"start_id"`
	EndId   int   `json:"end_id"`
}

type CleanupRequest struct {
	MinDays int    `json:"min_days"`
	StartId int    `json:"start_id"`
	EndId   int    `json:"end_id"`
	Reason  string `json:"reason"`
}

type ArchivedUserPublicInfo struct {
	Id              int    `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	Quota           int    `json:"quota"`
	UsedQuota       int    `json:"used_quota"`
	RequestCount    int    `json:"request_count"`
	ArchivedAt      int64  `json:"archived_at"`
	LinuxDOUsername string `json:"linux_do_username"`
}

func CheckArchivedUser(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请输入查询关键词",
		})
		return
	}

	user, err := model.FindArchivedUserByKeyword(keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"found":   false,
			"message": "未找到被清理的账号",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"found":   true,
		"data": ArchivedUserPublicInfo{
			Id:              user.Id,
			Username:        user.Username,
			DisplayName:     user.DisplayName,
			Quota:           user.Quota,
			UsedQuota:       user.UsedQuota,
			RequestCount:    user.RequestCount,
			ArchivedAt:      user.ArchivedAt,
			LinuxDOUsername: user.LinuxDOUsername,
		},
	})
}

type RecoverQuotaRequest struct {
	ArchivedId int `json:"archived_id" binding:"required"`
}

func RecoverQuotaFromArchived(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	var req RecoverQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	if req.ArchivedId <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的归档用户ID",
		})
		return
	}

	recoveredQuota, err := model.RecoverQuotaToUser(userId, req.ArchivedId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"message":         "额度恢复成功",
		"recovered_quota": recoveredQuota,
	})
}

func PreviewInactiveUsers(c *gin.Context) {
	minDaysStr := c.DefaultQuery("min_days", "7")
	minDays, err := strconv.Atoi(minDaysStr)
	if err != nil || minDays < 1 {
		minDays = 7
	}

	startIdStr := c.DefaultQuery("start_id", "0")
	startId, _ := strconv.Atoi(startIdStr)

	endIdStr := c.DefaultQuery("end_id", "0")
	endId, _ := strconv.Atoi(endIdStr)

	count, err := model.CountInactiveUsers(minDays, startId, endId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, CleanupPreviewResponse{
		Count:   count,
		MinDays: minDays,
		StartId: startId,
		EndId:   endId,
	})
}

func GetInactiveUsers(c *gin.Context) {
	minDaysStr := c.DefaultQuery("min_days", "7")
	minDays, err := strconv.Atoi(minDaysStr)
	if err != nil || minDays < 1 {
		minDays = 7
	}

	startIdStr := c.DefaultQuery("start_id", "0")
	startId, _ := strconv.Atoi(startIdStr)

	endIdStr := c.DefaultQuery("end_id", "0")
	endId, _ := strconv.Atoi(endIdStr)

	users, err := model.GetInactiveUsers(minDays, startId, endId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, users)
}

func CleanupInactiveUsers(c *gin.Context) {
	var req CleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}

	if req.MinDays < 1 {
		req.MinDays = 7
	}
	if req.Reason == "" {
		req.Reason = "不活跃用户清理"
	}

	adminId := c.GetInt("id")
	count, err := model.BatchArchiveInactiveUsers(req.MinDays, req.StartId, req.EndId, adminId, req.Reason)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"cleaned_count": count,
		"message":       "清理完成",
	})
}

func GetAllArchivedUsers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	users, total, err := model.GetAllArchivedUsers(pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(users)

	common.ApiSuccess(c, pageInfo)
}

func SearchArchivedUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	pageInfo := common.GetPageQuery(c)

	users, total, err := model.SearchArchivedUsers(keyword, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(users)

	common.ApiSuccess(c, pageInfo)
}

func GetArchivedUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	user, err := model.GetArchivedUserById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, user)
}

func RestoreArchivedUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	if err := model.RestoreArchivedUser(id); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"message": "用户恢复成功",
	})
}

func DeleteArchivedUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的ID",
		})
		return
	}

	archived, err := model.GetArchivedUserById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if err := archived.Delete(); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"message": "归档记录已永久删除",
	})
}
