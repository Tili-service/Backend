package email

import (
	"log"
	"os"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/resend/resend-go/v2"
	"github.com/wneessen/go-mail"
)

type Sender interface {
	SendEmail(to string, subject string, body string) error
}

type MailpitSender struct {
	client *mail.Client
}

type ResendSender struct {
	client *resend.Client
}

func NewEmailSender() (Sender, error) {
	env := os.Getenv("APP_ENV")

	if env == "production" {
		log.Println("Using Resend for email sending in production environment")
		apiKey := os.Getenv("RESEND_API_KEY")
		client := resend.NewClient(apiKey)
		return &ResendSender{client: client}, nil
	} else {
		log.Println("Using Mailpit for email sending in development environment")
		host := os.Getenv("TILI_SMTP_HOST")
		portStr := os.Getenv("TILI_SMTP_PORT")
		user := os.Getenv("TILI_SMTP_USER")
		pass := os.Getenv("TILI_SMTP_PASS")

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}

		options := []mail.Option{
			mail.WithPort(port),
		}

		if user != "" && pass != "" {
			options = append(options,
				mail.WithSMTPAuth(mail.SMTPAuthPlain),
				mail.WithUsername(user),
				mail.WithPassword(pass),
			)
		}

		if port == 1025 {
			options = append(options, mail.WithTLSPolicy(mail.NoTLS))
		} else {
			options = append(options, mail.WithSSLPort(true))
		}

		client, err := mail.NewClient(host, options...)
		if err != nil {
			return nil, err
		}

		return &MailpitSender{client: client}, nil
	}
}

var getWelcomeEmailContent = func(name string, email string) (string, error) {
	tpl, err := pongo2.FromFile("html/welcome_email.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"name":    name,
		"email":   email,
		"app_url": os.Getenv("APP_URL") + "/admin",
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

var getNewPaymentLinkEmailContent = func(offer string, paymentLinkURL string) (string, error) {
	tpl, err := pongo2.FromFile("html/new_payment_link.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"offer":            offer,
		"payment_link_url": paymentLinkURL,
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

var getNewLicenseActiveEmailContent = func(licenseLink string) (string, error) {
	tpl, err := pongo2.FromFile("html/new_license_active.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"license_link": licenseLink,
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

var getNewProfileCreatedEmailContent = func(profileStoreID int, profileName string, profilePIN string) (string, error) {
	tpl, err := pongo2.FromFile("html/new_profile_created.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"profile_store_id":    profileStoreID,
		"profile_name":        profileName,
		"profile_pin":         profilePIN,
		"manage_profile_link": os.Getenv("APP_URL") + "/admin/shop/" + strconv.Itoa(profileStoreID) + "/profils",
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

var getNewStoreCreatedEmailContent = func(storeName string, storeID int) (string, error) {
	tpl, err := pongo2.FromFile("html/new_store_created.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"store_name": storeName,
		"store_link": os.Getenv("APP_URL") + "/admin/shop/" + strconv.Itoa(storeID) + "/dashboard",
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

// Public wrapper functions that call the mockable variables
func GetWelcomeEmailContent(name string, email string) (string, error) {
	return getWelcomeEmailContent(name, email)
}

func GetNewPaymentLinkEmailContent(offer string, paymentLinkURL string) (string, error) {
	return getNewPaymentLinkEmailContent(offer, paymentLinkURL)
}

func GetNewLicenseActiveEmailContent(licenseLink string) (string, error) {
	return getNewLicenseActiveEmailContent(licenseLink)
}

func GetNewProfileCreatedEmailContent(profileStoreID int, profileName string, profilePIN string) (string, error) {
	return getNewProfileCreatedEmailContent(profileStoreID, profileName, profilePIN)
}

func GetNewStoreCreatedEmailContent(storeName string, storeID int) (string, error) {
	return getNewStoreCreatedEmailContent(storeName, storeID)
}

func (m *MailpitSender) SendEmail(to string, subject string, body string) error {
	msg := mail.NewMsg()
	msg.From("tili-service@local.dev")
	msg.To(to)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)
	return m.client.DialAndSend(msg)
}

func (r *ResendSender) SendEmail(to string, subject string, body string) error {
	params := &resend.SendEmailRequest{
		From:    "Tili-Services<no-reply@nduboi.fr>", // Todo change when we have tili domain and verify it with resend(default is domain nduboi.fr)
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	r.client.Emails.Send(params)
	return nil
}
