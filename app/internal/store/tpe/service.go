package tpe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"

	"tili/app/internal/store"
)

type Service struct {
	storeService *store.Service
	httpClient   *http.Client
	sumupBaseURL string
}

func NewService(storeServices ...*store.Service) *Service {
	var ss *store.Service
	if len(storeServices) > 0 {
		ss = storeServices[0]
	}
	return &Service{
		storeService: ss,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		sumupBaseURL: "https://api.sumup.com",
	}
}

type SumupReadersResponse struct {
	Items []*TPE `json:"items"`
}

func (s *Service) getTPE(ctx context.Context, st *store.Store) ([]*TPE, error) {
	if st == nil {
		return nil, errors.New("store is nil")
	}

	if st.SumupAccessToken == "" && st.SumupRefreshToken == "" {
		return nil, errors.New("no SumUp credentials available for this store")
	}

	// Premier essai : utiliser l'access token actuel
	if st.SumupAccessToken != "" {
		tpes, statusCode, err := s.fetchReadersFromSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken)
		if err == nil {
			return tpes, nil
		}
		// Si ce n'est pas une erreur d'authentification (401/403) et qu'on n'a pas de refresh token
		if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden && st.SumupRefreshToken == "" {
			return nil, err
		}
	}

	// Si l'access token est invalide/expiré ou vide, rafraîchir avec le refresh token
	if st.SumupRefreshToken == "" {
		return nil, errors.New("access token invalid or expired, and no refresh token available")
	}

	newToken, err := s.refreshToken(ctx, st.SumupRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh SumUp token: %w", err)
	}

	// Mettre à jour les tokens dans la DB si le service store est renseigné
	if s.storeService != nil {
		updatedStore, updateErr := s.storeService.UpdateSumupTokens(ctx, st.StoreID, newToken.AccessToken, newToken.RefreshToken)
		if updateErr == nil && updatedStore != nil {
			st.SumupAccessToken = updatedStore.SumupAccessToken
			if updatedStore.SumupRefreshToken != "" {
				st.SumupRefreshToken = updatedStore.SumupRefreshToken
			}
		} else {
			st.SumupAccessToken = newToken.AccessToken
			if newToken.RefreshToken != "" {
				st.SumupRefreshToken = newToken.RefreshToken
			}
		}
	} else {
		st.SumupAccessToken = newToken.AccessToken
		if newToken.RefreshToken != "" {
			st.SumupRefreshToken = newToken.RefreshToken
		}
	}

	// Réessayer la requête auprès de SumUp avec le nouvel access token
	tpes, _, err := s.fetchReadersFromSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TPE after token refresh: %w", err)
	}

	return tpes, nil
}

func (s *Service) fetchReadersFromSumup(ctx context.Context, merchantCode string, accessToken string) ([]*TPE, int, error) {
	baseURL := s.sumupBaseURL
	if baseURL == "" {
		baseURL = "https://api.sumup.com"
	}

	var apiURL string
	if merchantCode != "" {
		apiURL = fmt.Sprintf("%s/v0.1/merchants/%s/readers", baseURL, merchantCode)
	} else {
		apiURL = fmt.Sprintf("%s/v0.1/me/readers", baseURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, resp.StatusCode, errors.New("unauthorized")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("sumup API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Tenter de désérialiser au format { "items": [...] }
	var readersResp SumupReadersResponse
	if err := json.Unmarshal(bodyBytes, &readersResp); err == nil && readersResp.Items != nil {
		return readersResp.Items, http.StatusOK, nil
	}

	// Tenter de désérialiser au format liste directe [...]
	var readerList []*TPE
	if err := json.Unmarshal(bodyBytes, &readerList); err == nil {
		return readerList, http.StatusOK, nil
	}

	return []*TPE{}, http.StatusOK, nil
}

func (s *Service) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	tokenURL := "https://api.sumup.com/token"
	if s.sumupBaseURL != "" && s.sumupBaseURL != "https://api.sumup.com" {
		tokenURL = s.sumupBaseURL + "/token"
	}

	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID_SUMUP_OAUTH"),
		ClientSecret: os.Getenv("CLIENT_SECRET_SUMUP_OAUTH"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.sumup.com/authorize",
			TokenURL: tokenURL,
		},
	}

	t := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := oauthConfig.TokenSource(ctx, t)
	return tokenSource.Token()
}

func (s *Service) initiatePayment(ctx context.Context, st *store.Store, paymentRequest TPEPaymentRequest) (*ReaderCheckoutData, error) {
	if st == nil {
		return nil, errors.New("store is nil")
	}

	if st.SumupAccessToken == "" && st.SumupRefreshToken == "" {
		return nil, errors.New("no SumUp credentials available for this store")
	}

	if paymentRequest.TotalAmount.Value <= 0 {
		return nil, errors.New("total amount must be greater than zero")
	}

	if paymentRequest.TotalAmount.Currency == "" {
		paymentRequest.TotalAmount.Currency = "EUR"
	}

	if paymentRequest.TotalAmount.MinorUnit == 0 {
		paymentRequest.TotalAmount.MinorUnit = 2
	}

	if paymentRequest.Description == "" {
		paymentRequest.Description = "Payment"
	}

	if paymentRequest.TPEID == "" {
		return nil, errors.New("TPE ID is required for payment initiation")
	}

	// Premier essai avec l'access token actuel
	if st.SumupAccessToken != "" {
		data, statusCode, err := s.sendReaderCheckoutToSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken, paymentRequest)
		if err == nil {
			return data, nil
		}
		if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden && st.SumupRefreshToken == "" {
			return nil, err
		}
	}

	// Rafraîchir le token si expiré ou vide
	if st.SumupRefreshToken == "" {
		return nil, errors.New("access token invalid or expired, and no refresh token available")
	}

	newToken, err := s.refreshToken(ctx, st.SumupRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh SumUp token: %w", err)
	}

	// Mettre à jour les tokens dans la DB
	if s.storeService != nil {
		updatedStore, updateErr := s.storeService.UpdateSumupTokens(ctx, st.StoreID, newToken.AccessToken, newToken.RefreshToken)
		if updateErr == nil && updatedStore != nil {
			st.SumupAccessToken = updatedStore.SumupAccessToken
			if updatedStore.SumupRefreshToken != "" {
				st.SumupRefreshToken = updatedStore.SumupRefreshToken
			}
		} else {
			st.SumupAccessToken = newToken.AccessToken
			if newToken.RefreshToken != "" {
				st.SumupRefreshToken = newToken.RefreshToken
			}
		}
	} else {
		st.SumupAccessToken = newToken.AccessToken
		if newToken.RefreshToken != "" {
			st.SumupRefreshToken = newToken.RefreshToken
		}
	}

	// Réessayer la demande de paiement auprès de SumUp
	data, _, err := s.sendReaderCheckoutToSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken, paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send payment request after token refresh: %w", err)
	}

	return data, nil
}

func (s *Service) sendReaderCheckoutToSumup(ctx context.Context, merchantCode string, accessToken string, req TPEPaymentRequest) (*ReaderCheckoutData, int, error) {
	baseURL := s.sumupBaseURL
	if baseURL == "" {
		baseURL = "https://api.sumup.com"
	}

	var paymentURL string
	if merchantCode != "" {
		paymentURL = fmt.Sprintf("%s/v0.1/merchants/%s/readers/%s/checkout", baseURL, merchantCode, req.TPEID)
	} else {
		paymentURL = fmt.Sprintf("%s/v0.1/me/readers/%s/checkout", baseURL, req.TPEID)
	}

	bodyData := struct {
		TotalAmount TotalAmount `json:"total_amount"`
		Description string      `json:"description,omitempty"`
	}{
		TotalAmount: req.TotalAmount,
		Description: req.Description,
	}

	payloadBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, paymentURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create payment request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, resp.StatusCode, errors.New("unauthorized")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, resp.StatusCode, fmt.Errorf("sumup API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var checkoutResp ReaderCheckoutResponse
	if err := json.Unmarshal(bodyBytes, &checkoutResp); err == nil && checkoutResp.Data.ID != "" {
		return &checkoutResp.Data, resp.StatusCode, nil
	}

	var directData ReaderCheckoutData
	if err := json.Unmarshal(bodyBytes, &directData); err == nil && directData.ID != "" {
		return &directData, resp.StatusCode, nil
	}

	fmt.Print("Payment request sent successfully. ResponseBody: ", string(bodyBytes), "\n")

	return &ReaderCheckoutData{Status: "PENDING"}, resp.StatusCode, nil
}

func (s *Service) getPaymentStatus(ctx context.Context, st *store.Store, tpeID string, checkoutID string) (*ReaderCheckoutData, error) {
	if st == nil {
		return nil, errors.New("store is nil")
	}

	if tpeID == "" || checkoutID == "" {
		return nil, errors.New("tpe_id and checkout_id are required")
	}

	if st.SumupAccessToken != "" {
		data, statusCode, err := s.fetchCheckoutStatusFromSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken, tpeID, checkoutID)
		if err == nil {
			return data, nil
		}
		if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden && st.SumupRefreshToken == "" {
			return nil, err
		}
	}

	if st.SumupRefreshToken == "" {
		return nil, errors.New("access token invalid or expired, and no refresh token available")
	}

	newToken, err := s.refreshToken(ctx, st.SumupRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh SumUp token: %w", err)
	}

	if s.storeService != nil {
		updatedStore, updateErr := s.storeService.UpdateSumupTokens(ctx, st.StoreID, newToken.AccessToken, newToken.RefreshToken)
		if updateErr == nil && updatedStore != nil {
			st.SumupAccessToken = updatedStore.SumupAccessToken
		} else {
			st.SumupAccessToken = newToken.AccessToken
		}
	} else {
		st.SumupAccessToken = newToken.AccessToken
	}

	data, _, err := s.fetchCheckoutStatusFromSumup(ctx, st.SumupMerchantCode, st.SumupAccessToken, tpeID, checkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch checkout status after token refresh: %w", err)
	}

	return data, nil
}

func (s *Service) fetchCheckoutStatusFromSumup(ctx context.Context, merchantCode string, accessToken string, tpeID string, checkoutID string) (*ReaderCheckoutData, int, error) {
	baseURL := s.sumupBaseURL
	if baseURL == "" {
		baseURL = "https://api.sumup.com"
	}

	var statusURL string
	if merchantCode != "" {
		statusURL = fmt.Sprintf("%s/v0.1/merchants/%s/readers/%s/checkout/%s", baseURL, merchantCode, tpeID, checkoutID)
	} else {
		statusURL = fmt.Sprintf("%s/v0.1/me/readers/%s/checkout/%s", baseURL, tpeID, checkoutID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, resp.StatusCode, errors.New("unauthorized")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("sumup API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var checkoutResp ReaderCheckoutResponse
	if err := json.Unmarshal(bodyBytes, &checkoutResp); err == nil && checkoutResp.Data.ID != "" {
		return &checkoutResp.Data, resp.StatusCode, nil
	}

	var directData ReaderCheckoutData
	if err := json.Unmarshal(bodyBytes, &directData); err == nil && directData.ID != "" {
		return &directData, resp.StatusCode, nil
	}

	return nil, resp.StatusCode, fmt.Errorf("unable to parse checkout status response")
}
