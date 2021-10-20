package router

import "github.com/gin-gonic/gin"

func InitRouters() *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.GET("/users/", func(context *gin.Context) {
		//context.String(200, "get user infos")
		context.JSON(200, gin.H{"message": "get user infos"})
	})

	return ginRouter
}
