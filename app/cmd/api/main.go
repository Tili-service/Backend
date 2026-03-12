package main

import (
	"log"

	_ "tili/app/docs"
	"tili/app/internal/catalog"
	"tili/app/internal/categorie"
	"tili/app/internal/item"
	"tili/app/internal/license"
	"tili/app/internal/store"
	"tili/app/internal/user"
	"tili/app/pkg/db"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type 'Bearer {token}' to authorize
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

	catalogRepo := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepo)
	catalogHandler := catalog.NewHandler(catalogService)

	itemRepo := item.NewRepository(db)
	itemService := item.NewService(itemRepo)
	itemHandler := item.NewHandler(itemService)

	categorieRepo := categorie.NewRepository(db)
	categorieService := categorie.NewService(categorieRepo)
	categorieHandler := categorie.NewHandler(categorieService)

	r := gin.Default()

	userHandler.RegisterRoutes(r)
	storeHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)
	catalogHandler.RegisterRoutes(r)
	itemHandler.RegisterRoutes(r)
	categorieHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:8080")
	r.Run(":8080")
}
