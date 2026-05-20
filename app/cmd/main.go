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
	"tili/app/internal/sale"
	"tili/app/internal/store"

	"tili/app/pkg/db"
	"tili/app/pkg/email"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v84"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"context"
)

type storeRepositoryAdapter struct {
	repo *store.Repository
}

func (a *storeRepositoryAdapter) FindByID(ctx context.Context, id int) (*profile.StoreData, error) {
	store, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, nil
	}
	return &profile.StoreData{BuyerID: store.BuyerID}, nil
}

type accountRepositoryAdapter struct {
	repo *account.Repository
}

func (a *accountRepositoryAdapter) FindByID(ctx context.Context, id int) (*profile.AccountData, error) {
	acc, err := a.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, nil
	}
	return &profile.AccountData{Email: acc.Email}, nil
}

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
	emailClient, err := email.NewEmailSender()
	if err != nil {
		log.Fatalf("Erreur lors de la création du client email: %s", err)
	}
	db := db.NewDb()
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	accountRepo := account.NewRepository(db)
	profileRepo := profile.NewRepository(db)
	storeRepo := store.NewRepository(db)

	// Create adapters for profile service interfaces
	storeAdapter := &storeRepositoryAdapter{repo: storeRepo}
	accountAdapter := &accountRepositoryAdapter{repo: accountRepo}

	profileService := profile.NewServiceWithEmail(profileRepo, storeAdapter, accountAdapter, emailClient)
	profileHandler := profile.NewHandler(profileService)

	storeService := store.NewService(storeRepo)
	storeHandler := store.NewHandler(storeService, profileService)

	licenseRepo := license.NewRepository(db)
	licenseService := license.NewService(licenseRepo, emailClient)
	licenseHandler := license.NewHandler(licenseService)

	accountService := account.NewService(accountRepo, storeService, profileService, licenseService, emailClient)
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

	saleRepo := sale.NewRepository(db)
	saleService := sale.NewService(saleRepo)
	saleHandler := sale.NewHandler(saleService)

	r := gin.Default()

	profileHandler.RegisterRoutes(r)
	storeHandler.RegisterRoutes(r)
	licenseHandler.RegisterRoutes(r)
	accountHandler.RegisterRoutes(r)
	catalogHandler.RegisterRoutes(r)
	itemHandler.RegisterRoutes(r)
	categorieHandler.RegisterRoutes(r)
	payementmethodHandler.RegisterRoutes(r)
	saleHandler.RegisterRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	log.Println("Serveur démarré sur http://localhost:" + os.Getenv("BACKEND_PORT"))
	r.Run(":" + os.Getenv("BACKEND_PORT"))
}
