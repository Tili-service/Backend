package main

import (
	"log"

	_ "tili/app/docs"
	"tili/app/internal/user"
	"tili/app/pkg/db"
	"tili/app/internal/license"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	db := db.NewDb()
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	licenseRepo := license.NewRepository(db)
	licenseService := license.NewService(licenseRepo, userService)
	licenseHandler := license.NewHandler(licenseService)
	r := gin.Default()

	userHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:8080")
	r.Run(":8080")
}
