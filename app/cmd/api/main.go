package main

import (
	"log"

	_ "tili/app/docs"
	"tili/app/internal/account"
	"tili/app/internal/license"
	"tili/app/internal/profile"
	"tili/app/internal/store"
	"tili/app/pkg/db"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Tili API
// @version         1.0
// @host            localhost:8080
// @basePath        /

// @securityDefinitions.apikey AccountToken
// @in header
// @name Authorization
// @description JWT obtenu après login account (POST /account/login)

// @securityDefinitions.apikey ProfileToken
// @in header
// @name Authorization
// @description JWT obtenu après login profil avec PIN (POST /profile/login/pin)
func main() {
	db := db.NewDb()

	profileRepo := profile.NewRepository(db)
	profileService := profile.NewService(profileRepo)
	profileHandler := profile.NewHandler(profileService)

	storeRepo := store.NewRepository(db)
	storeService := store.NewService(storeRepo)
	storeHandler := store.NewHandler(storeService, profileService)

	licenseRepo := license.NewRepository(db)
	licenseService := license.NewService(licenseRepo)
	licenseHandler := license.NewHandler(licenseService)

	accountRepo := account.NewRepository(db)
	accountService := account.NewService(accountRepo, storeService, profileService, licenseService)
	accountHandler := account.NewHandler(accountService)

	r := gin.Default()

	profileHandler.RegisterRoutes(r)
	storeHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)
	accountHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:8080")
	r.Run(":8080")
}
