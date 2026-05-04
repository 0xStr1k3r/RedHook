package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redhook/backend/src/models"
	"gorm.io/gorm"
)

type CampaignController struct {
	DB *gorm.DB
}

func NewCampaignController(db *gorm.DB) *CampaignController {
	return &CampaignController{DB: db}
}

func (ctrl *CampaignController) List(c *gin.Context) {
	var campaigns []models.Campaign
	query := ctrl.DB.Preload("Template").Preload("LandingPage").Preload("CreatedBy")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&campaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaigns"})
		return
	}

	c.JSON(http.StatusOK, campaigns)
}

func (ctrl *CampaignController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := ctrl.DB.Preload("Template").Preload("LandingPage").Preload("CreatedBy").First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

type CreateCampaignRequest struct {
	Name          string     `json:"name" binding:"required"`
	Description   string     `json:"description"`
	TemplateID    uint       `json:"template_id" binding:"required"`
	LandingPageID uint       `json:"landing_page_id" binding:"required"`
	Difficulty    int        `json:"difficulty"`
	ScheduleTime  *time.Time `json:"schedule_time"`
}

func (ctrl *CampaignController) Create(c *gin.Context) {
	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	campaign := models.Campaign{
		Name:          req.Name,
		Description:   req.Description,
		TemplateID:    req.TemplateID,
		LandingPageID: req.LandingPageID,
		Difficulty:    req.Difficulty,
		Status:        "draft",
		CreatedByID:   userID,
		ScheduleTime:  req.ScheduleTime,
	}

	if err := ctrl.DB.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	c.JSON(http.StatusCreated, campaign)
}

type UpdateCampaignRequest struct {
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	TemplateID    uint       `json:"template_id"`
	LandingPageID uint       `json:"landing_page_id"`
	Difficulty    int        `json:"difficulty"`
	ScheduleTime  *time.Time `json:"schedule_time"`
	Status        string     `json:"status"`
}

func (ctrl *CampaignController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	var req UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var campaign models.Campaign
	if err := ctrl.DB.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	if req.Name != "" {
		campaign.Name = req.Name
	}
	if req.Description != "" {
		campaign.Description = req.Description
	}
	if req.TemplateID != 0 {
		campaign.TemplateID = req.TemplateID
	}
	if req.LandingPageID != 0 {
		campaign.LandingPageID = req.LandingPageID
	}
	if req.Difficulty != 0 {
		campaign.Difficulty = req.Difficulty
	}
	if req.ScheduleTime != nil {
		campaign.ScheduleTime = req.ScheduleTime
	}
	if req.Status != "" {
		campaign.Status = req.Status
	}

	if err := ctrl.DB.Save(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

func (ctrl *CampaignController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	if err := ctrl.DB.Delete(&models.Campaign{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign deleted"})
}

func (ctrl *CampaignController) Launch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := ctrl.DB.Preload("Template").Preload("LandingPage").First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	if campaign.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign already launched"})
		return
	}

	now := time.Now()
	campaign.Status = "running"
	campaign.LaunchTime = &now

	if err := ctrl.DB.Save(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to launch campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign launched", "campaign": campaign})
}

func (ctrl *CampaignController) GetStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	var campaign models.Campaign
	if err := ctrl.DB.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	var totalSent int64
	ctrl.DB.Model(&models.EmailLog{}).Where("campaign_id = ?", id).Count(&totalSent)

	var totalOpened int64
	ctrl.DB.Model(&models.EmailLog{}).Where("campaign_id = ? AND opened = ?", id, true).Count(&totalOpened)

	var totalClicks int64
	ctrl.DB.Model(&models.ClickLog{}).Where("campaign_id = ?", id).Count(&totalClicks)

	var totalSubmissions int64
	ctrl.DB.Model(&models.SubmissionLog{}).Where("campaign_id = ?", id).Count(&totalSubmissions)

	var totalReports int64
	ctrl.DB.Model(&models.Report{}).Where("campaign_id = ?", id).Count(&totalReports)

	openRate := float64(0)
	if totalSent > 0 {
		openRate = float64(totalOpened) / float64(totalSent) * 100
	}
	clickRate := float64(0)
	if totalOpened > 0 {
		clickRate = float64(totalClicks) / float64(totalOpened) * 100
	}
	submissionRate := float64(0)
	if totalClicks > 0 {
		submissionRate = float64(totalSubmissions) / float64(totalClicks) * 100
	}
	reportRate := float64(0)
	if totalSent > 0 {
		reportRate = float64(totalReports) / float64(totalSent) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign_id":       id,
		"total_sent":        totalSent,
		"total_opened":      totalOpened,
		"total_clicks":      totalClicks,
		"total_submissions": totalSubmissions,
		"total_reports":     totalReports,
		"open_rate":         openRate,
		"click_rate":        clickRate,
		"submission_rate":   submissionRate,
		"report_rate":       reportRate,
	})
}
