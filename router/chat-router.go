package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetChatRouter(router *gin.Engine) {
	chatRouter := router.Group("/api/chat")
	chatRouter.Use(middleware.GlobalAPIRateLimit())
	{
		chatRouter.GET("/config", controller.GetChatRoomConfig)
		chatRouter.GET("/count", controller.GetChatRoomMessageCount)
		chatRouter.GET("/ws", middleware.ChatWsAuth(), controller.ChatRoomWs)
		chatRouter.POST("/images", middleware.UserAuth(), controller.UploadChatRoomImage)
		chatRouter.GET("/images/:date/:name", controller.GetChatRoomImage)
		chatRouter.POST("/announcement", middleware.AdminAuth(), controller.SetChatRoomAnnouncement)
		chatRouter.GET("/messages", middleware.UserAuth(), controller.ListChatRoomMessages)
		chatRouter.POST("/messages", middleware.UserAuth(), controller.PostChatRoomMessage)
	}
}
