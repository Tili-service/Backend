package email

import (
	"github.com/google/uuid"
)

// MockGetNewLicenseActiveEmailContent temporarily replaces the email content function
// for testing purposes
func MockGetNewLicenseActiveEmailContent(mockFunc func(string) (string, error)) func() {
	original := getNewLicenseActiveEmailContent
	getNewLicenseActiveEmailContent = mockFunc
	return func() {
		getNewLicenseActiveEmailContent = original
	}
}

// MockGetNewPaymentLinkEmailContent temporarily replaces the email content function
// for testing purposes
func MockGetNewPaymentLinkEmailContent(mockFunc func(string, string) (string, error)) func() {
	original := getNewPaymentLinkEmailContent
	getNewPaymentLinkEmailContent = mockFunc
	return func() {
		getNewPaymentLinkEmailContent = original
	}
}

// MockGetWelcomeEmailContent temporarily replaces the email content function
// for testing purposes
func MockGetWelcomeEmailContent(mockFunc func(string, string) (string, error)) func() {
	original := getWelcomeEmailContent
	getWelcomeEmailContent = mockFunc
	return func() {
		getWelcomeEmailContent = original
	}
}

// MockGetNewProfileCreatedEmailContent temporarily replaces the email content function
// for testing purposes
func MockGetNewProfileCreatedEmailContent(mockFunc func(uuid.UUID, string, string) (string, error)) func() {
	original := getNewProfileCreatedEmailContent
	getNewProfileCreatedEmailContent = mockFunc
	return func() {
		getNewProfileCreatedEmailContent = original
	}
}

// MockGetNewStoreCreatedEmailContent temporarily replaces the email content function
// for testing purposes
func MockGetNewStoreCreatedEmailContent(mockFunc func(string, uuid.UUID) (string, error)) func() {
	original := getNewStoreCreatedEmailContent
	getNewStoreCreatedEmailContent = mockFunc
	return func() {
		getNewStoreCreatedEmailContent = original
	}
}
