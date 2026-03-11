package main

import (
	"log"

	_ "tili/app/docs"
	"tili/app/internal/catalogue"
	"tili/app/internal/license"
	"tili/app/internal/store"
	"tili/app/internal/user"
	"tili/app/pkg/db"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	db := db.NewDb()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	storeRepo := store.NewRepository(db)
	storeService := store.NewService(storeRepo)
	storeHandler := store.NewHandler(storeService)

	licenseRepo := license.NewRepository(db)
	licenseService := license.NewService(licenseRepo, userService, storeService)
	licenseHandler := license.NewHandler(licenseService)

	catalogueRepo := catalogue.NewRepository(db)
	catalogueService := catalogue.NewService(catalogueRepo)
	catalogueHandler := catalogue.NewHandler(catalogueService)


	r := gin.Default()

	userHandler.RegisterRoutes(r)
	storeHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)
	catalogueHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:8080")
	r.Run(":8080")
}
