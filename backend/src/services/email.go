package services

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/redhook/backend/src/config"
	"github.com/redhook/backend/src/models"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/resend/resend-go/v3"
)

type EmailService struct {
	DB     *gorm.DB
	Cfg    *config.Config
	SES    *ses.SES
	Resend *resend.Client
}

func NewEmailService(db *gorm.DB, cfg *config.Config) *EmailService {
	svc := &EmailService{
		DB:  db,
		Cfg: cfg,
	}

	if cfg.AWS.Enabled {
		svc.initSES()
	}

	if cfg.Resend.Enabled && cfg.Resend.APIKey != "" {
		svc.Resend = resend.NewClient(cfg.Resend.APIKey)
	}

	return svc
}

func (svc *EmailService) initSES() error {
	if svc.Cfg.AWS.AccessKeyID != "" && svc.Cfg.AWS.SecretKey != "" {
		awsSession, err := session.NewSession(&aws.Config{
			Region:      aws.String(svc.Cfg.AWS.Region),
			Credentials: credentials.NewStaticCredentials(svc.Cfg.AWS.AccessKeyID, svc.Cfg.AWS.SecretKey, ""),
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS session: %w", err)
		}
		svc.SES = ses.New(awsSession)
	} else {
		awsSession, err := session.NewSessionWithOptions(session.Options{
			Profile: svc.Cfg.AWS.Profile,
			Config: aws.Config{
				Region: aws.String(svc.Cfg.AWS.Region),
			},
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS session: %w", err)
		}
		svc.SES = ses.New(awsSession)
	}

	return nil
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
			"TrackLink": fmt.Sprintf("%s/track/click/%s", svc.Cfg.App.PhishingURL, token),
			"OpenLink":  fmt.Sprintf("%s/track/open/%s", svc.Cfg.App.PhishingURL, token),
		})

		subject := campaign.Template.Subject
		fromName := campaign.Template.FromName
		if fromName == "" {
			fromName = "RedHook"
		}

		emailLog := models.EmailLog{
			UserID:         user.ID,
			CampaignID:     campaign.ID,
			TemplateID:     campaign.TemplateID,
			RecipientEmail: user.Email,
			TrackingToken:  token,
		}
		svc.DB.Create(&emailLog)

		var err error
		if svc.Cfg.Resend.Enabled && svc.Resend != nil {
			err = svc.sendEmailResend(user.Email, fromName, subject, body)
		} else if svc.Cfg.AWS.Enabled && svc.SES != nil {
			err = svc.sendEmailAWS(user.Email, fromName, subject, body)
		} else {
			err = svc.sendEmailSMTP(smtpConfig, user.Email, fromName, subject, body)
		}

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

func (svc *EmailService) sendEmailResend(to, fromName, subject, body string) error {
	if svc.Resend == nil {
		return fmt.Errorf("Resend not initialized")
	}

	sender := svc.Cfg.Resend.FromEmail

	params := &resend.SendEmailRequest{
		From:    sender,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	sent, err := svc.Resend.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	_ = sent
	return nil
}

func (svc *EmailService) sendEmailAWS(to, fromName, subject, body string) error {
	if svc.SES == nil {
		return fmt.Errorf("AWS SES not initialized")
	}

	sender := svc.Cfg.AWS.SESFromEmail
	if sender == "" {
		sender = svc.Cfg.SMTP.From
	}

	charSet := "UTF-8"
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(to)},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(fmt.Sprintf("%s <%s>", fromName, sender)),
	}

	result, err := svc.SES.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				return fmt.Errorf("message rejected: %s", aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				return fmt.Errorf("sender not verified: %s", aerr.Error())
			default:
				return fmt.Errorf("SES error: %s", aerr.Error())
			}
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Printf("Email sent successfully: %s\n", *result.MessageId)
	return nil
}

func (svc *EmailService) sendEmailSMTP(cfg models.SMTPConfig, to, fromName, subject, body string) error {
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

	err := smtp.SendMail(addr, auth, cfg.FromEmail, []string{to}, msg)
	return err
}

func (svc *EmailService) TestConnection() error {
	if svc.Cfg.AWS.Enabled && svc.SES != nil {
		_, err := svc.SES.GetSendQuota(&ses.GetSendQuotaInput{})
		return err
	}

	smtpConfig := models.SMTPConfig{}
	if err := svc.DB.Where("is_active = ?", true).First(&smtpConfig).Error; err != nil {
		return fmt.Errorf("no SMTP config found")
	}

	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)
	var auth smtp.Auth
	if smtpConfig.Username != "" {
		auth = smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	}

	err := smtp.SendMail(addr, auth, smtpConfig.FromEmail, []string{smtpConfig.FromEmail}, []byte("Test message"))
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
