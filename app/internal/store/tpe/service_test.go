package tpe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"tili/app/internal/store"
	"tili/app/pkg/db"
)

func setupMockDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	assert.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	return bunDB, mock
}

func TestGetTPE_NoCredentials(t *testing.T) {
	svc := NewService()
	st := &store.Store{
		StoreID: uuid.New(),
	}

	readers, err := svc.getTPE(context.Background(), st)
	assert.Error(t, err)
	assert.Nil(t, readers)
	assert.Equal(t, "no SumUp credentials available for this store", err.Error())
}

func TestGetTPE_SuccessFirstTry(t *testing.T) {
	// Mock SumUp Readers Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v0.1/merchants/M123/readers", r.URL.Path)
		assert.Equal(t, "Bearer valid_access_token", r.Header.Get("Authorization"))

		readerName := "Solo Terminal 1"
		readerID := "rdr_123"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SumupReadersResponse{
			Items: []*TPE{
				{ID: &readerID, Name: &readerName},
			},
		})
	}))
	defer server.Close()

	svc := NewService()
	svc.sumupBaseURL = server.URL

	st := &store.Store{
		StoreID:           uuid.New(),
		SumupMerchantCode: "M123",
		SumupAccessToken:  "valid_access_token",
	}

	readers, err := svc.getTPE(context.Background(), st)
	assert.NoError(t, err)
	assert.Len(t, readers, 1)
	assert.Equal(t, "Solo Terminal 1", *readers[0].Name)
	assert.Equal(t, "rdr_123", *readers[0].ID)
}

func TestGetTPE_TokenExpired_AutoRefreshSuccess(t *testing.T) {
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			assert.Equal(t, "POST", r.Method)
			_ = r.ParseForm()
			assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
			assert.Equal(t, "old_refresh_token", r.Form.Get("refresh_token"))

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "new_access_token",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "new_refresh_token",
			})
			return
		}

		if r.URL.Path == "/v0.1/merchants/M123/readers" {
			attempt++
			if attempt == 1 {
				assert.Equal(t, "Bearer expired_access_token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			assert.Equal(t, "Bearer new_access_token", r.Header.Get("Authorization"))
			readerName := "Refreshed Terminal"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(SumupReadersResponse{
				Items: []*TPE{
					{Name: &readerName},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "sumup_merchant_code", "sumup_access_token", "sumup_refresh_token", "siret"}).
		AddRow(storeID, "My Store", buyerID, uuid.New(), time.Now(), "", "M123", "expired_access_token", "old_refresh_token", "")

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	storeRepo := store.NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(storeRepo)

	svc := NewService(storeSvc)
	svc.sumupBaseURL = server.URL

	st := &store.Store{
		StoreID:           storeID,
		SumupMerchantCode: "M123",
		SumupAccessToken:  "expired_access_token",
		SumupRefreshToken: "old_refresh_token",
	}

	readers, err := svc.getTPE(context.Background(), st)
	assert.NoError(t, err)
	assert.Len(t, readers, 1)
	assert.Equal(t, "Refreshed Terminal", *readers[0].Name)
	assert.Equal(t, "new_access_token", st.SumupAccessToken)
	assert.Equal(t, "new_refresh_token", st.SumupRefreshToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInitiatePayment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v0.1/merchants/M123/readers/rdr_123/checkout", r.URL.Path)
		assert.Equal(t, "Bearer valid_access_token", r.Header.Get("Authorization"))
		assert.Equal(t, "POST", r.Method)

		var reqBody struct {
			TotalAmount TotalAmount `json:"total_amount"`
			Description string      `json:"description"`
		}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, int64(1500), reqBody.TotalAmount.Value)
		assert.Equal(t, "EUR", reqBody.TotalAmount.Currency)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":     "chk_123",
				"status": "PENDING",
			},
		})
	}))
	defer server.Close()

	svc := NewService()
	svc.sumupBaseURL = server.URL

	st := &store.Store{
		StoreID:           uuid.New(),
		SumupMerchantCode: "M123",
		SumupAccessToken:  "valid_access_token",
	}

	req := TPEPaymentRequest{
		TotalAmount: TotalAmount{
			Value:    1500,
			Currency: "EUR",
		},
		Description: "Test Payment",
		TPEID:       "rdr_123",
	}

	checkoutData, err := svc.initiatePayment(context.Background(), st, req)
	assert.NoError(t, err)
	assert.NotNil(t, checkoutData)
	assert.Equal(t, "chk_123", checkoutData.ID)
}
