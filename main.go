package main

import (
	config "report/base"
	"report/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	router.IntiateRoutes(r)
	config.LoadConfig()
	r.Run(":8000")
}
