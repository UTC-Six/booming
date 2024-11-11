package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// ... 其他路由 ...

	router.POST("/landmine", CreateLandmine)
	router.POST("/check_proximity", CheckProximity)

	// ... 其他路由 ...
}
