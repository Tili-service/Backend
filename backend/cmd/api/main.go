package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"tili/backend/internal/user"
	"tili/backend/pkg/db"
	_ "tili/backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	db := db.NewDb()
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	r := gin.Default()

	userHandler.RegisterRoutes(r)

	swaggerHandler := ginSwagger.WrapHandler(swaggerfiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		swaggerHandler(c)
	})

	log.Println("Serveur démarré sur http://localhost:8080")
	r.Run(":8080")
}
