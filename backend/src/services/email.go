package services

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/phishguard/backend/src/config"
	"github.com/phishguard/backend/src/models"
	"gorm.io/gorm"
)

type EmailService struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewEmailService(db *gorm.DB, cfg *config.Config) *EmailService {
	return &EmailService{DB: db, Cfg: cfg}
}

type SendRequest struct {
	CampaignID uint
	Recipients []uint
}

func (svc *EmailService) SendCampaign(req SendRequest) error {
	var campaign models.Campaign
	if err := svc.DB.Preload("Template").Preload("LandingPage").First(&campaign, req.CampaignID).Error; err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}

	if campaign.Template == nil || campaign.LandingPage == nil {
		return fmt.Errorf("template or landing page not found")
	}

	var users []models.User
	if err := svc.DB.Where("id IN ?", req.Recipients).Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	smtpConfig := models.SMTPConfig{}
	svc.DB.Where("is_active = ?", true).First(&smtpConfig)

	var sentCount int
	var failedCount int

	for _, user := range users {
		token := generateTrackingToken(user.ID, campaign.ID)

		body := replaceVariables(campaign.Template.BodyHTML, map[string]string{
			"Name":      user.Name,
			"Email":     user.Email,
			"TrackLink": fmt.Sprintf("%s/track/click/%s", svc.Cfg.App.PhishingDomain, token),
			"OpenLink":  fmt.Sprintf("%s/track/open/%s", svc.Cfg.App.PhishingDomain, token),
		})

		subject := campaign.Template.Subject
		fromName := campaign.Template.FromName
		if fromName == "" {
			fromName = "PhishGuard"
		}

		emailLog := models.EmailLog{
			UserID:         user.ID,
			CampaignID:     campaign.ID,
			TemplateID:     campaign.TemplateID,
			RecipientEmail: user.Email,
			TrackingToken:  token,
		}
		svc.DB.Create(&emailLog)

		err := svc.sendEmail(smtpConfig, user.Email, fromName, subject, body)
		if err != nil {
			failedCount++
			continue
		}
		sentCount++
	}

	if campaign.Status == "draft" {
		svc.DB.Model(&models.Campaign{}).Where("id = ?", campaign.ID).
			Update("status", "sent")
	}

	return nil
}

func (svc *EmailService) sendEmail(cfg models.SMTPConfig, to, fromName, subject, body string) error {
	from := cfg.FromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, from)
	}

	headers := fmt.Sprintf("From: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n", from, subject)

	msg := []byte(headers + "\r\n" + body)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	var err error
	err = smtp.SendMail(addr, auth, cfg.FromEmail, []string{to}, msg)

	return err
}

func generateTrackingToken(userID, campaignID uint) string {
	timestamp := time.Now().UnixNano()
	raw := fmt.Sprintf("%d:%d:%d", userID, campaignID, timestamp)
	encoded := strings.Replace(raw, "/", "_", -1)
	return encoded
}

func replaceVariables(html string, vars map[string]string) string {
	result := html
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
