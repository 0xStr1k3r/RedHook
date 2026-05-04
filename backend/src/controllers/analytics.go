package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/models"
	"gorm.io/gorm"
)

type AnalyticsController struct {
	DB *gorm.DB
}

func NewAnalyticsController(db *gorm.DB) *AnalyticsController {
	return &AnalyticsController{DB: db}
}

type OverviewStats struct {
	TotalUsers      int64   `json:"total_users"`
	TotalCampaigns  int64   `json:"total_campaigns"`
	TotalEmailsSent int64   `json:"total_emails_sent"`
	OpenRate        float64 `json:"open_rate"`
	ClickRate       float64 `json:"click_rate"`
	SubmissionRate  float64 `json:"submission_rate"`
	ReportRate      float64 `json:"report_rate"`
	HighRiskUsers   int64   `json:"high_risk_users"`
}

func (ctrl *AnalyticsController) GetOverview(c *gin.Context) {
	var stats OverviewStats

	ctrl.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	ctrl.DB.Model(&models.Campaign{}).Count(&stats.TotalCampaigns)

	var totalSent int64
	ctrl.DB.Model(&models.EmailLog{}).Count(&totalSent)
	stats.TotalEmailsSent = totalSent

	var totalOpened int64
	ctrl.DB.Model(&models.EmailLog{}).Where("opened = ?", true).Count(&totalOpened)

	var totalClicks int64
	ctrl.DB.Model(&models.ClickLog{}).Count(&totalClicks)

	var totalSubmissions int64
	ctrl.DB.Model(&models.SubmissionLog{}).Count(&totalSubmissions)

	var totalReports int64
	ctrl.DB.Model(&models.Report{}).Count(&totalReports)

	if totalSent > 0 {
		stats.OpenRate = float64(totalOpened) / float64(totalSent) * 100
		stats.ReportRate = float64(totalReports) / float64(totalSent) * 100
	}
	if totalOpened > 0 {
		stats.ClickRate = float64(totalClicks) / float64(totalOpened) * 100
	}
	if totalClicks > 0 {
		stats.SubmissionRate = float64(totalSubmissions) / float64(totalClicks) * 100
	}

	ctrl.DB.Model(&models.UserRiskScore{}).Where("risk_level = ?", "high").Count(&stats.HighRiskUsers)

	c.JSON(http.StatusOK, stats)
}

type DepartmentStats struct {
	Department   string  `json:"department"`
	TotalUsers   int64   `json:"total_users"`
	TotalClicks  int64   `json:"total_clicks"`
	ClickRate    float64 `json:"click_rate"`
	AvgRiskScore float64 `json:"avg_risk_score"`
	RiskLevel    string  `json:"risk_level"`
}

func (ctrl *AnalyticsController) GetDepartment(c *gin.Context) {
	department := c.Param("dept")

	var stats DepartmentStats
	stats.Department = department

	ctrl.DB.Model(&models.User{}).Where("department = ?", department).Count(&stats.TotalUsers)

	var clicks int64
	ctrl.DB.Model(&models.ClickLog{}).
		Joins("JOIN users ON click_logs.user_id = users.id").
		Where("users.department = ?", department).
		Count(&clicks)
	stats.TotalClicks = clicks

	if stats.TotalUsers > 0 {
		avgRisk := float64(0)
		ctrl.DB.Model(&models.UserRiskScore{}).
			Joins("JOIN users ON user_risk_scores.user_id = users.id").
			Where("users.department = ?", department).
			Select("AVG(risk_score)").
			Scan(&avgRisk)
		stats.AvgRiskScore = avgRisk

		if avgRisk <= 5 {
			stats.RiskLevel = "low"
		} else if avgRisk <= 15 {
			stats.RiskLevel = "medium"
		} else {
			stats.RiskLevel = "high"
		}
	}

	c.JSON(http.StatusOK, stats)
}

func (ctrl *AnalyticsController) GetUserRisk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var riskScore models.UserRiskScore
	if err := ctrl.DB.Preload("User").First(&riskScore, id).Error; err != nil {
		clicks := int64(0)
		submissions := int64(0)
		reports := int64(0)

		ctrl.DB.Model(&models.ClickLog{}).Where("user_id = ?", id).Count(&clicks)
		ctrl.DB.Model(&models.SubmissionLog{}).Where("user_id = ?", id).Count(&submissions)
		ctrl.DB.Model(&models.Report{}).Where("user_id = ?", id).Count(&reports)

		score := int(clicks)*2 + int(submissions)*5 - int(reports)*3
		if score < 0 {
			score = 0
		}

		level := "low"
		if score > 15 {
			level = "high"
		} else if score > 5 {
			level = "medium"
		}

		riskScore = models.UserRiskScore{
			UserID:           uint(id),
			TotalClicks:      int(clicks),
			TotalSubmissions: int(submissions),
			TotalReports:     int(reports),
			RiskScore:        score,
			RiskLevel:        level,
		}
	} else {
		if riskScore.RiskScore < 0 {
			riskScore.RiskScore = 0
		}
		if riskScore.RiskLevel == "" {
			if riskScore.RiskScore > 15 {
				riskScore.RiskLevel = "high"
			} else if riskScore.RiskScore > 5 {
				riskScore.RiskLevel = "medium"
			} else {
				riskScore.RiskLevel = "low"
			}
		}
	}

	c.JSON(http.StatusOK, riskScore)
}

func (ctrl *AnalyticsController) GetCampaignAnalytics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	type CampaignAnalytics struct {
		CampaignID  uint    `json:"campaign_id"`
		Status      string  `json:"status"`
		Sent        int64   `json:"sent"`
		Opened      int64   `json:"opened"`
		Clicks      int64   `json:"clicks"`
		Submissions int64   `json:"submissions"`
		Reports     int64   `json:"reports"`
		OpenRate    float64 `json:"open_rate"`
		ClickRate   float64 `json:"click_rate"`
	}

	var campaign models.Campaign
	if err := ctrl.DB.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	var analytics CampaignAnalytics
	analytics.CampaignID = campaign.ID
	analytics.Status = campaign.Status

	ctrl.DB.Model(&models.EmailLog{}).Where("campaign_id = ?", id).Count(&analytics.Sent)
	ctrl.DB.Model(&models.EmailLog{}).Where("campaign_id = ? AND opened = ?", id, true).Count(&analytics.Opened)
	ctrl.DB.Model(&models.ClickLog{}).Where("campaign_id = ?", id).Count(&analytics.Clicks)
	ctrl.DB.Model(&models.SubmissionLog{}).Where("campaign_id = ?", id).Count(&analytics.Submissions)
	ctrl.DB.Model(&models.Report{}).Where("campaign_id = ?", id).Count(&analytics.Reports)

	if analytics.Sent > 0 {
		analytics.OpenRate = float64(analytics.Opened) / float64(analytics.Sent) * 100
	}
	if analytics.Opened > 0 {
		analytics.ClickRate = float64(analytics.Clicks) / float64(analytics.Opened) * 100
	}

	c.JSON(http.StatusOK, analytics)
}

func (ctrl *AnalyticsController) GetTrends(c *gin.Context) {
	type Trend struct {
		Date    string `json:"date"`
		Clicks  int64  `json:"clicks"`
		Submits int64  `json:"submits"`
		Reports int64  `json:"reports"`
	}

	var trends []Trend

	ctrl.DB.Model(&models.ClickLog{}).
		Select("DATE(clicked_at) as date, COUNT(*) as clicks").
		Group("DATE(clicked_at)").
		Order("date DESC").
		Limit(30).
		Scan(&trends)

	c.JSON(http.StatusOK, trends)
}
