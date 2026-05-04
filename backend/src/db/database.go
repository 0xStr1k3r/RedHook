package db

import (
	"fmt"
	"log"
	"time"

	"github.com/phishguard/backend/src/config"
	"github.com/phishguard/backend/src/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()
	log.Printf("Connecting to database: %s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return db, nil
}

func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Campaign{},
		&models.EmailTemplate{},
		&models.LandingPage{},
		&models.EmailLog{},
		&models.ClickLog{},
		&models.SubmissionLog{},
		&models.Report{},
		&models.TrainingCompletion{},
		&models.SMTPConfig{},
		&models.UserRiskScore{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func Seed(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	adminPassword := "$2a$10$rQEY8yV7E2zVJJJJJJJJJ.QQQEEEEEEAAAAAAADDDDDGGGGIIIIIIII"
	adminUser := models.User{
		Email:        "admin@phishguard.local",
		Name:         "System Admin",
		Department:   "IT",
		Role:         "admin",
		PasswordHash: adminPassword,
	}
	db.Create(&adminUser)

	templates := []models.EmailTemplate{
		{
			Name:     "IT Password Reset",
			Subject:  "Urgent: Your account password will expire in 24 hours",
			BodyHTML: `<html><body><h2>Password Reset Required</h2><p>Dear {{.Name}},</p><p>Your account password will expire in 24 hours. Please click the link below to reset your password:</p><p><a href="{{.TrackLink}}">Reset Password</a></p><p>If you did not request this, please ignore this email.</p></body></html>`,
			Category: "IT Support",
			FromName: "IT Support Team",
		},
		{
			Name:     "HR Benefits Update",
			Subject:  "Action Required: Update Your Benefits Information",
			BodyHTML: `<html><body><h2>Benefits Update</h2><p>Dear {{.Name}},</p><p>Please review and update your benefits information by clicking the link below:</p><p><a href="{{.TrackLink}}">Update Benefits</a></p></body></html>`,
			Category: "HR",
			FromName: "Human Resources",
		},
		{
			Name:     "Invoice Payment",
			Subject:  "Invoice #INV-2024-001 - Payment Due",
			BodyHTML: `<html><body><h2>Invoice Due</h2><p>Dear {{.Name}},</p><p>Invoice #INV-2024-001 is due for payment. Please review and submit payment:</p><p><a href="{{.TrackLink}}">View Invoice</a></p></body></html>`,
			Category: "Finance",
			FromName: "Accounts Payable",
		},
		{
			Name:     "Microsoft 365 Login",
			Subject:  "Sign in to Microsoft 365",
			BodyHTML: `<html><body><h2>Sign in to Microsoft 365</h2><p>Dear {{.Name}},</p><p>Please sign in to your Microsoft 365 account:</p><p><a href="{{.TrackLink}}">Sign In</a></p></body></html>`,
			Category: "IT Support",
			FromName: "Microsoft",
		},
		{
			Name:     "Shipping Notification",
			Subject:  "Your package has shipped!",
			BodyHTML: `<html><body><h2>Package Shipped</h2><p>Dear {{.Name}},</p><p>Your package has been shipped. Track your delivery:</p><p><a href="{{.TrackLink}}">Track Package</a></p></body></html>`,
			Category: "Shipping",
			FromName: "Shipping Team",
		},
	}
	for i := range templates {
		db.Create(&templates[i])
	}

	landingPages := []models.LandingPage{
		{
			Name:               "Microsoft Login",
			HTMLContent:        `<!DOCTYPE html><html><head><title>Sign in</title></head><body><h2>Sign in to your account</h2><form method="POST"><input type="email" name="email" placeholder="Email" required><input type="password" name="password" placeholder="Password" required><button type="submit">Sign in</button></form></body></html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>You were part of a phishing simulation!</h2><p>This email was a simulated phishing test. Here are signs you missed:</p><ul><li>Urgency</li><li>Suspicious link</li></ul>`,
		},
		{
			Name:               "Generic Login",
			HTMLContent:        `<!DOCTYPE html><html><head><title>Login</title></head><body><h2>Login</h2><form method="POST"><input type="text" name="username" placeholder="Username" required><input type="password" name="password" placeholder="Password" required><button type="submit">Login</button></form></body></html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>This was a phishing simulation!</h2><p>Always verify the sender and URL before entering credentials.</p>`,
		},
		{
			Name:               "Password Reset",
			HTMLContent:        `<!DOCTYPE html><html><head><title>Reset Password</title></head><body><h2>Reset Your Password</h2><form method="POST"><input type="email" name="email" placeholder="Enter your email" required><button type="submit">Send Reset Link</button></form></body></html>`,
			CaptureCredentials: false,
			TrainingContent:    `<h2>This was a phishing simulation!</h2><p>Never enter your credentials on unfamiliar pages.</p>`,
		},
	}
	for i := range landingPages {
		db.Create(&landingPages[i])
	}

	smtpConfig := models.SMTPConfig{
		Name:      "Default SMTP",
		Host:      "localhost",
		Port:      587,
		Username:  "",
		FromEmail: "noreply@phishguard.local",
		FromName:  "PhishGuard",
		UseTLS:    true,
		IsActive:  true,
	}
	db.Create(&smtpConfig)

	log.Println("Database seeded successfully")
	return nil
}
