package main

import (
	"report/router"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	router.IntiateRoutes(r)

	r.Run(":8000")
}
