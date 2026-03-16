package main

import (
	"log"
	"os"

	_ "tili/app/docs"
	"tili/app/internal/account"
	"tili/app/internal/catalog"
	"tili/app/internal/categorie"
	"tili/app/internal/item"
	"tili/app/internal/license"
	"tili/app/internal/payementmethod"
	"tili/app/internal/profile"
	"tili/app/internal/store"

	"tili/app/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v84"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Tili API
// @version         0.1

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
	stripe.Key = os.Getenv("STRIPE_API_KEY")

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

	catalogRepo := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepo)
	catalogHandler := catalog.NewHandler(catalogService)

	itemRepo := item.NewRepository(db)
	itemService := item.NewService(itemRepo)
	itemHandler := item.NewHandler(itemService)

	categorieRepo := categorie.NewRepository(db)
	categorieService := categorie.NewService(categorieRepo)
	categorieHandler := categorie.NewHandler(categorieService)

	payementmethodRepo := payementmethod.NewRepository(db)
	payementmethodService := payementmethod.NewService(payementmethodRepo)
	payementmethodHandler := payementmethod.NewHandler(payementmethodService)

	r := gin.Default()

	profileHandler.RegisterRoutes(r)
	storeHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)
	accountHandler.RegisterRoutes(r)
	catalogHandler.RegisterRoutes(r)
	itemHandler.RegisterRoutes(r)
	categorieHandler.RegisterRoutes(r)
	payementmethodHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:8000")
	r.Run(":8000")
}
