package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/models"
	"gorm.io/gorm"
)

type PhishingHandler struct {
	DB *gorm.DB
}

func NewPhishingHandler(db *gorm.DB) *PhishingHandler {
	return &PhishingHandler{DB: db}
}

func (h *PhishingHandler) TrackOpen(c *gin.Context) {
	token := c.Param("token")

	var emailLog models.EmailLog
	if err := h.DB.Where("tracking_token = ?", token).First(&emailLog).Error; err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}

	if !emailLog.Opened {
		now := time.Now()
		emailLog.Opened = true
		emailLog.OpenedAt = &now
		h.DB.Save(&emailLog)
	}

	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.String(http.StatusOK, "GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
}

func (h *PhishingHandler) TrackClick(c *gin.Context) {
	token := c.Param("token")

	var emailLog models.EmailLog
	if err := h.DB.Where("tracking_token = ?", token).First(&emailLog).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	userID, campaignID, _ := parseToken(token)

	clickLog := models.ClickLog{
		UserID:        userID,
		CampaignID:    campaignID,
		TrackingToken: token,
		ClickedAt:     time.Now(),
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
	}
	h.DB.Create(&clickLog)

	c.Redirect(http.StatusFound, "/landing/"+token)
}

func (h *PhishingHandler) TrackSubmit(c *gin.Context) {
	token := c.Param("token")

	var emailLog models.EmailLog
	if err := h.DB.Where("tracking_token = ?", token).First(&emailLog).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	dataLength := 0
	dataPattern := "unknown"

	if email := c.PostForm("email"); email != "" {
		dataLength = len(email)
		dataPattern = "email"
	}
	if password := c.PostForm("password"); password != "" {
		dataLength += len(password)
		dataPattern = "credentials"
	}
	if username := c.PostForm("username"); username != "" {
		dataLength += len(username)
		dataPattern = "credentials"
	}

	userID, _, _ := parseToken(token)

	submission := models.SubmissionLog{
		UserID:        userID,
		CampaignID:    emailLog.CampaignID,
		TrackingToken: token,
		SubmittedAt:   time.Now(),
		IPAddress:     c.ClientIP(),
		DataLength:    dataLength,
		DataPattern:   dataPattern,
		TrainingShown: true,
	}
	h.DB.Create(&submission)

	c.Redirect(http.StatusFound, "/train/"+token)
}

func (h *PhishingHandler) ShowLanding(c *gin.Context) {
	token := c.Param("token")

	var emailLog models.EmailLog
	if err := h.DB.Where("tracking_token = ?", token).First(&emailLog).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	var campaign models.Campaign
	if err := h.DB.Preload("LandingPage").First(&campaign, emailLog.CampaignID).Error; err != nil {
		c.String(http.StatusNotFound, "Campaign not found")
		return
	}

	if campaign.LandingPage == nil {
		c.String(http.StatusNotFound, "Landing page not found")
		return
	}

	page := campaign.LandingPage

	if c.Request.Method == "POST" {
		dataLength := 0
		dataPattern := "unknown"

		if email := c.PostForm("email"); email != "" {
			dataLength = len(email)
			dataPattern = "email"
		}
		if password := c.PostForm("password"); password != "" {
			dataLength += len(password)
			dataPattern = "credentials"
		}
		if username := c.PostForm("username"); username != "" {
			dataLength += len(username)
			dataPattern = "credentials"
		}

		userID, _, _ := parseToken(token)

		submission := models.SubmissionLog{
			UserID:        userID,
			CampaignID:    emailLog.CampaignID,
			TrackingToken: token,
			SubmittedAt:   time.Now(),
			IPAddress:     c.ClientIP(),
			DataLength:    dataLength,
			DataPattern:   dataPattern,
			TrainingShown: true,
		}
		h.DB.Create(&submission)

		c.Redirect(http.StatusFound, "/train/"+token)
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, page.HTMLContent)
}

func (h *PhishingHandler) ShowTraining(c *gin.Context) {
	token := c.Param("token")

	var emailLog models.EmailLog
	if err := h.DB.Where("tracking_token = ?", token).First(&emailLog).Error; err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	var campaign models.Campaign
	if err := h.DB.Preload("LandingPage").First(&campaign, emailLog.CampaignID).Error; err != nil {
		c.String(http.StatusNotFound, "Campaign not found")
		return
	}

	trainingContent := `<!DOCTYPE html>
<html>
<head>
    <title>Phishing Awareness Training</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; background: #f5f5f5; }
        .alert { background: #fff3cd; border: 1px solid #ffc107; padding: 20px; border-radius: 5px; }
        .alert h2 { color: #856404; margin-top: 0; }
        .tips { background: #d4edda; border: 1px solid #28a745; padding: 20px; border-radius: 5px; margin-top: 20px; }
        .tips h3 { color: #155724; margin-top: 0; }
        .tips ul { color: #155724; line-height: 1.8; }
        .btn { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="alert">
        <h2>You have just participated in a phishing simulation!</h2>
        <p>This email was a simulated phishing test to help improve organizational security awareness.</p>
    </div>
    <div class="tips">
        <h3>How to identify phishing emails:</h3>
        <ul>
            <li>Check the sender's email address carefully</li>
            <li>Hover over links before clicking to see the actual URL</li>
            <li>Look for urgent or threatening language</li>
            <li>Verify requests through official channels</li>
            <li>When in doubt, report to IT security</li>
        </ul>
    </div>
    <a href="/" class="btn">Back to Safe Page</a>
</body>
</html>`

	if campaign.LandingPage != nil && campaign.LandingPage.TrainingContent != "" {
		trainingContent = campaign.LandingPage.TrainingContent
	}

	userID, _, _ := parseToken(token)

	submission := models.TrainingCompletion{
		UserID:      userID,
		CampaignID:  emailLog.CampaignID,
		ModuleName:  "Post-simulation training",
		CompletedAt: time.Now(),
	}
	h.DB.Create(&submission)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, trainingContent)
}

func (h *PhishingHandler) ReportPhishing(c *gin.Context) {
	token := c.Query("token")

	if token != "" {
		userID, campaignID, _ := parseToken(token)
		if userID > 0 && campaignID > 0 {
			report := models.Report{
				UserID:     userID,
				CampaignID: campaignID,
				ReportedAt: time.Now(),
			}
			h.DB.Create(&report)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thank you for reporting. This was a simulated phishing test.",
		"status":  "received",
	})
}

func parseToken(token string) (uint, uint, error) {
	parts := strings.Split(token, ":")
	if len(parts) < 2 {
		return 0, 0, nil
	}

	var userID, campaignID int64
	_, err := fmt.Sscanf(parts[0], "%d", &userID)
	if err != nil {
		return 0, 0, err
	}
	_, err = fmt.Sscanf(parts[1], "%d", &campaignID)
	if err != nil {
		return 0, 0, err
	}

	return uint(userID), uint(campaignID), nil
}
