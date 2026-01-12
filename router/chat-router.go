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
		chatRouter.GET("/ws", middleware.ChatWsAuth(), controller.ChatRoomWs)
		chatRouter.POST("/images", middleware.UserAuth(), controller.UploadChatRoomImage)
		chatRouter.GET("/images/:date/:name", controller.GetChatRoomImage)
	}
}
