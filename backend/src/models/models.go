package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Department   string    `gorm:"size:100" json:"department"`
	Role         string    `gorm:"size:50;default:user" json:"role"`
	Password     string    `gorm:"-" json:"-"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	EmailLogs           []EmailLog           `gorm:"foreignKey:UserID" json:"-"`
	ClickLogs           []ClickLog           `gorm:"foreignKey:UserID" json:"-"`
	SubmissionLogs      []SubmissionLog      `gorm:"foreignKey:UserID" json:"-"`
	Reports             []Report             `gorm:"foreignKey:UserID" json:"-"`
	TrainingCompletions []TrainingCompletion `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) TableName() string {
	return "users"
}

type Campaign struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	TemplateID    uint           `gorm:"index" json:"template_id"`
	Template      *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	LandingPageID uint           `gorm:"index" json:"landing_page_id"`
	LandingPage   *LandingPage   `gorm:"foreignKey:LandingPageID" json:"landing_page,omitempty"`
	Status        string         `gorm:"size:50;default:draft" json:"status"`
	Difficulty    int            `gorm:"default:1" json:"difficulty"`
	ScheduleTime  *time.Time     `json:"schedule_time"`
	LaunchTime    *time.Time     `json:"launch_time"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CreatedByID   uint           `gorm:"index" json:"created_by_id"`
	CreatedBy     *User          `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	EmailLogs      []EmailLog      `gorm:"foreignKey:CampaignID" json:"-"`
	ClickLogs      []ClickLog      `gorm:"foreignKey:CampaignID" json:"-"`
	SubmissionLogs []SubmissionLog `gorm:"foreignKey:CampaignID" json:"-"`
	Reports        []Report        `gorm:"foreignKey:CampaignID" json:"-"`
}

func (c *Campaign) TableName() string {
	return "campaigns"
}

type EmailTemplate struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	Subject         string    `gorm:"size:500;not null" json:"subject"`
	BodyHTML        string    `gorm:"type:text;not null" json:"body_html"`
	DifficultyScore float64   `gorm:"default:1" json:"difficulty_score"`
	Category        string    `gorm:"size:100" json:"category"`
	FromName        string    `gorm:"size:255" json:"from_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Campaigns []Campaign `gorm:"foreignKey:TemplateID" json:"-"`
}

func (e *EmailTemplate) TableName() string {
	return "templates"
}

type LandingPage struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"size:255;not null" json:"name"`
	HTMLContent        string    `gorm:"type:text;not null" json:"html_content"`
	RedirectURL        string    `gorm:"size:500" json:"redirect_url"`
	CaptureCredentials bool      `gorm:"default:false" json:"capture_credentials"`
	TrainingContent    string    `gorm:"type:text" json:"training_content"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	Campaigns []Campaign `gorm:"foreignKey:LandingPageID" json:"-"`
}

func (l *LandingPage) TableName() string {
	return "landing_pages"
}

type EmailLog struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"index;not null" json:"user_id"`
	User           *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CampaignID     uint           `gorm:"index;not null" json:"campaign_id"`
	Campaign       *Campaign      `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	TemplateID     uint           `gorm:"index" json:"template_id"`
	Template       *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	RecipientEmail string         `gorm:"size:255;not null" json:"recipient_email"`
	TrackingToken  string         `gorm:"uniqueIndex;size:255;not null" json:"tracking_token"`
	SentAt         time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"sent_at"`
	OpenedAt       *time.Time     `json:"opened_at"`
	Opened         bool           `gorm:"default:false" json:"opened"`
}

func (e *EmailLog) TableName() string {
	return "email_logs"
}

type ClickLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CampaignID    uint      `gorm:"index" json:"campaign_id"`
	Campaign      *Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	TrackingToken string    `gorm:"size:255;index" json:"tracking_token"`
	ClickedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"clicked_at"`
	IPAddress     string    `gorm:"size:50" json:"ip_address"`
	UserAgent     string    `gorm:"type:text" json:"user_agent"`
	LandingPageID uint      `gorm:"index" json:"landing_page_id"`
}

func (c *ClickLog) TableName() string {
	return "click_logs"
}

type SubmissionLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	User          *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CampaignID    uint      `gorm:"index" json:"campaign_id"`
	Campaign      *Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	TrackingToken string    `gorm:"size:255;index" json:"tracking_token"`
	SubmittedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"submitted_at"`
	IPAddress     string    `gorm:"size:50" json:"ip_address"`
	DataLength    int       `json:"data_length"`
	DataPattern   string    `gorm:"size:100" json:"data_pattern"`
	TrainingShown bool      `gorm:"default:true" json:"training_shown"`
}

func (s *SubmissionLog) TableName() string {
	return "submission_logs"
}

type Report struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index" json:"user_id"`
	User       *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CampaignID uint           `gorm:"index" json:"campaign_id"`
	Campaign   *Campaign      `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	TemplateID uint           `gorm:"index" json:"template_id"`
	Template   *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	ReportedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"reported_at"`
}

func (r *Report) TableName() string {
	return "reports"
}

type TrainingCompletion struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"index;not null" json:"user_id"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CampaignID       uint      `gorm:"index" json:"campaign_id"`
	Campaign         *Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	ModuleName       string    `gorm:"size:255;not null" json:"module_name"`
	CompletedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"completed_at"`
	TimeSpentSeconds int       `json:"time_spent_seconds"`
}

func (t *TrainingCompletion) TableName() string {
	return "training_completions"
}

type SMTPConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Host      string    `gorm:"size:255;not null" json:"host"`
	Port      int       `gorm:"not null" json:"port"`
	Username  string    `gorm:"size:255" json:"username"`
	Password  string    `gorm:"-" json:"-"`
	FromEmail string    `gorm:"size:255;not null" json:"from_email"`
	FromName  string    `gorm:"size:255" json:"from_name"`
	UseTLS    bool      `gorm:"default:true" json:"use_tls"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *SMTPConfig) TableName() string {
	return "smtp_configs"
}

type UserRiskScore struct {
	UserID           uint      `gorm:"primaryKey" json:"user_id"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TotalClicks      int       `json:"total_clicks"`
	TotalSubmissions int       `json:"total_submissions"`
	TotalReports     int       `json:"total_reports"`
	RiskScore        int       `json:"risk_score"`
	RiskLevel        string    `gorm:"size:20" json:"risk_level"`
	LastUpdated      time.Time `json:"last_updated"`
}
