package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redhook/backend/src/models"
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

	userID, campaignID, dataLength := parseToken(token)
	dataPattern := "credentials"

	submission := models.SubmissionLog{
		UserID:        userID,
		CampaignID:    campaignID,
		TrackingToken: token,
		SubmittedAt:   time.Now(),
		IPAddress:     c.ClientIP(),
		DataLength:    dataLength,
		DataPattern:   dataPattern,
	}
	h.DB.Create(&submission)

	c.Redirect(http.StatusFound, "/train/"+token)
	return
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

	trainingHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Session Expired</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f5f5; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .box { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); max-width: 400px; text-align: center; }
        h2 { color: #333; font-size: 20px; margin-bottom: 15px; }
        p { color: #666; font-size: 14px; line-height: 1.6; }
        .btn { display: inline-block; padding: 10px 24px; background: #0067b8; color: white; text-decoration: none; border-radius: 4px; margin-top: 15px; }
    </style>
</head>
<body>
    <div class="box">
        <h2>Session Expired</h2>
        <p>Your session has timed out. Please close this browser and try again.</p>
        <a href="https://login.microsoft.com" class="btn">Go to Login</a>
    </div>
</body>
</html>`

	c.Data(200, "text/html; charset=utf-8", []byte(trainingHTML))
}

func (h *PhishingHandler) ShowLandingPreview(c *gin.Context) {
	idStr := c.Param("id")
	var page models.LandingPage
	if err := h.DB.First(&page, idStr).Error; err != nil {
		c.String(http.StatusNotFound, "Page not found")
		return
	}

	html := page.HTMLContent
	c.Data(200, "text/html; charset=utf-8", []byte(html))
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
		dataPattern := "credentials"

		if email := c.PostForm("email"); email != "" {
			dataLength += len(email)
		}
		if password := c.PostForm("password"); password != "" {
			dataLength += len(password)
		}
		if username := c.PostForm("username"); username != "" {
			dataLength += len(username)
		}

		userID, campaignID, _ := parseToken(token)

		submission := models.SubmissionLog{
			UserID:        userID,
			CampaignID:    campaignID,
			TrackingToken: token,
			SubmittedAt:   time.Now(),
			IPAddress:     c.ClientIP(),
			DataLength:    dataLength,
			DataPattern:   dataPattern,
		}
		h.DB.Create(&submission)

		c.Redirect(http.StatusFound, "/train/"+token)
		return
	}

	html := page.HTMLContent

	trackURL := fmt.Sprintf("/track/click/%s", token)
	html = strings.Replace(html, `<form method="POST">`, `<form method="POST" action="`+trackURL+`">`, 1)

	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

func (h *PhishingHandler) ReportPhishing(c *gin.Context) {
	c.String(http.StatusOK, "This page has been reported. Thank you.")
}

func parseToken(token string) (userID, campaignID uint, dataLength int) {
	fmt.Sscanf(token, "%d:%d:%d", &userID, &campaignID, &dataLength)
	return
}
