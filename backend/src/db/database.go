package db

import (
	"fmt"
	"log"
	"time"

	"github.com/redhook/backend/src/config"
	"github.com/redhook/backend/src/models"

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

	adminPassword := "$2a$10$RmEFEQjawGf9mINqDcV6j.KZj41SRXDAiytFoANb9vJm7XzLiC.DK"
	adminUser := models.User{
		Email:        "admin@redhook.local",
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
		{
			Name:     "Microsoft - Security Alert",
			Subject:  "Unusual sign-in activity on your Microsoft account",
			BodyHTML: `<html><body><p>Dear {{.Name}},</p><p>We detected unusual sign-in activity.</p><p><a href="{{.TrackLink}}">Review Activity</a></p></body></html>`,
			Category: "Microsoft",
			FromName: "Microsoft Security",
		},
		{
			Name:     "Google - Password Changed",
			Subject:  "Your Google password was changed",
			BodyHTML: `<html><body><p>Dear {{.Name}},</p><p>Your password was changed.</p><p><a href="{{.TrackLink}}">Secure Account</a></p></body></html>`,
			Category: "Google",
			FromName: "Google Security",
		},
		{
			Name:     "LinkedIn - Profile Viewed",
			Subject:  "Your profile was viewed by a recruiter",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>Your profile was viewed.</p><p><a href="{{.TrackLink}}">View Profile</a></p></body></html>`,
			Category: "Social",
			FromName: "LinkedIn",
		},
		{
			Name:     "Amazon - Delivery",
			Subject:  "Your package is out for delivery",
			BodyHTML: `<html><body><p>Hello {{.Name}},</p><p>Your package is out for delivery.</p><p><a href="{{.TrackLink}}">Track Package</a></p></body></html>`,
			Category: "Shipping",
			FromName: "Amazon Shipping",
		},
		{
			Name:     "GitHub - Security Alert",
			Subject:  "Suspicious activity in your repository",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>Unusual activity detected.</p><p><a href="{{.TrackLink}}">View Activity</a></p></body></html>`,
			Category: "Development",
			FromName: "GitHub Security",
		},
		{
			Name:     "Facebook - New Login",
			Subject:  "New login to your Facebook account",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>New login detected.</p><p><a href="{{.TrackLink}}">Secure Account</a></p></body></html>`,
			Category: "Social",
			FromName: "Facebook Security",
		},
		{
			Name:     "Dropbox - File Shared",
			Subject:  "Someone shared a file with you",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>A file was shared with you.</p><p><a href="{{.TrackLink}}">View File</a></p></body></html>`,
			Category: "File Sharing",
			FromName: "Dropbox",
		},
		{
			Name:     "Netflix - Payment Failed",
			Subject:  "Your Netflix payment failed",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>Payment failed. Update to continue.</p><p><a href="{{.TrackLink}}">Update Payment</a></p></body></html>`,
			Category: "Entertainment",
			FromName: "Netflix",
		},
		{
			Name:     "PayPal - Verify Identity",
			Subject:  "Please verify your PayPal identity",
			BodyHTML: `<html><body><p>Dear {{.Name}},</p><p>Please verify your identity.</p><p><a href="{{.TrackLink}}">Verify Identity</a></p></body></html>`,
			Category: "Finance",
			FromName: "PayPal Security",
		},
		{
			Name:     "Zoom - Meeting Invite",
			Subject:  "You have been invited to a Zoom meeting",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>You've been invited to a meeting.</p><p><a href="{{.TrackLink}}">Join Meeting</a></p></body></html>`,
			Category: "Business",
			FromName: "Zoom",
		},
		{
			Name:     "Slack - Workspace Invite",
			Subject:  "You've been invited to join a workspace",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>You've been invited to join Slack.</p><p><a href="{{.TrackLink}}">Join Workspace</a></p></body></html>`,
			Category: "Business",
			FromName: "Slack",
		},
		{
			Name:     "Adobe - Verify Email",
			Subject:  "Verify your Adobe email",
			BodyHTML: `<html><body><p>Hi {{.Name}},</p><p>Verify your email to activate account.</p><p><a href="{{.TrackLink}}">Verify Email</a></p></body></html>`,
			Category: "Creative",
			FromName: "Adobe",
		},
		{
			Name:     "HR - Benefits Enrollment",
			Subject:  "Annual Benefits Enrollment Open",
			BodyHTML: `<html><body><p>Dear {{.Name}},</p><p>Benefits enrollment is now open.</p><p><a href="{{.TrackLink}}">Review Benefits</a></p></body></html>`,
			Category: "HR",
			FromName: "HR Department",
		},
		{
			Name:     "Finance - Invoice Due",
			Subject:  "Invoice Payment Due",
			BodyHTML: `<html><body><p>Dear {{.Name}},</p><p>Invoice payment is due.</p><p><a href="{{.TrackLink}}">View Invoice</a></p></body></html>`,
			Category: "Finance",
			FromName: "Accounts Payable",
		},
	}
	for i := range templates {
		db.Create(&templates[i])
	}

	landingPages := []models.LandingPage{
		{
			Name: "Microsoft 365",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Sign in to your account</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; background: #f2f2f2; margin: 0; padding: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .box { background: white; padding: 40px; border-radius: 4px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 400px; }
        .logo { text-align: center; margin-bottom: 20px; }
        .logo img { height: 24px; }
        h2 { color: #333; font-weight: 600; font-size: 24px; margin-bottom: 10px; }
        p { color: #666; font-size: 14px; margin-bottom: 20px; }
        input { width: 100%; padding: 6px; margin: 8px 0; border: 1px solid #999; border-radius: 2px; box-sizing: border-box; }
        button { width: 100%; padding: 8px; background: #0067b8; color: white; border: none; border-radius: 2px; font-size: 15px; cursor: pointer; }
        button:hover { background: #005a9e; }
        .link { color: #0067b8; text-decoration: none; font-size: 13px; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo"><img src="https://img-prod-cms-rt-microsoft-com.akamaized.net/cms/blb9QGBq2Z5TqE4F4.png" alt="Microsoft"></div>
        <h2>Sign in</h2>
        <p>Enter the email and password associated with your account.</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email, phone, or Skype" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>Your account security is important. Always verify:</p><ul><li>The URL should be the official domain</li><li>Check for HTTPS</li></ul>`,
		},
		{
			Name: "Google Login",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Sign in - Google</title>
    <style>
        body { font-family: Arial, sans-serif; background: #fff; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .box { border: 1px solid #dadce0; border-radius: 8px; padding: 48px 40px; width: 350px; }
        .logo { text-align: center; margin-bottom: 16px; }
        .logo span { color: #4285f4; font-size: 24px; font-weight: bold; }
        h2 { color: #202124; font-size: 24px; font-weight: 400; margin: 0 0 8px; }
        p { color: #5f6368; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 13px 15px; margin: 8px 0; border: 1px solid #dadce0; border-radius: 4px; box-sizing: border-box; font-size: 14px; }
        input:focus { outline: none; border: 1px solid #4285f4; }
        button { background: #1a73e8; color: white; border: none; padding: 10px 24px; border-radius: 4px; font-size: 14px; cursor: pointer; font-weight: 500; }
        button:hover { background: #1557b0; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo"><span>Google</span></div>
        <h2>Sign in</h2>
        <p>Use your Google Account</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email or phone number" required>
            <button type="submit">Next</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. Check:</p><ul><li>URL should be accounts.google.com</li><li>No legitimate login page has just "Next" button first</li></ul>`,
		},
		{
			Name: "LinkedIn",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Sign in to LinkedIn</title>
    <style>
        body { font-family: -apple-system, system-ui, sans-serif; background: #f3f2ee; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .box { background: white; padding: 32px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); width: 380px; }
        .logo { font-size: 28px; color: #0a66c2; margin-bottom: 20px; font-weight: bold; }
        h2 { font-size: 24px; color: rgba(0,0,0,0.9); margin: 0 0 4px; }
        p { color: rgba(0,0,0,0.6); font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 12px; margin: 8px 0; border: 1px solid rgba(0,0,0,0.3); border-radius: 4px; box-sizing: border-box; font-size: 14px; }
        button { width: 100%; padding: 10px; background: #0a66c2; color: white; border: none; border-radius: 24px; font-size: 16px; font-weight: 600; cursor: pointer; }
        button:hover { background: #004182; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">LinkedIn</div>
        <h2>Sign in</h2>
        <p>Stay updated on your professional world</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email or phone" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. Always check:</p><ul><li>Email domain matches the official domain</li><li>Direct URL to login page</li></ul>`,
		},
		{
			Name: "Amazon",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Amazon Sign-In</title>
    <style>
        body { font-family: Arial, sans-serif; background: white; margin: 0; display: flex; justify-content: center; padding: 20px; }
        .box { max-width: 350px; width: 100%; text-align: center; }
        .logo { margin-bottom: 20px; }
        .logo img { width: 103px; }
        .card { border: 1px solid #ddd; border-radius: 4px; padding: 20px; margin-bottom: 20px; }
        h2 { font-size: 24px; color: #111; margin: 0 0 10px; text-align: left; }
        label { font-size: 13px; font-weight: 700; color: #111; display: block; text-align: left; margin: 10px 0 5px; }
        input { width: 100%; padding: 7px; margin-bottom: 10px; border: 1px solid #a6a6a6; border-radius: 3px; box-sizing: border-box; font-size: 13px; }
        button { width: 100%; padding: 8px; background: #f0c14b; border: 1px solid #a88734; border-radius: 3px; font-size: 13px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo"><img src="https://upload.wikimedia.org/wikipedia/commons/a/a9/Amazon_logo.svg" alt="Amazon"></div>
        <div class="card">
            <h2>Sign in</h2>
            <label>Email or mobile phone number</label>
            <form method="POST">
                <input type="text" name="email" placeholder="Enter email" required>
                <label>Password</label>
                <input type="password" name="password" placeholder="Enter password" required>
                <button type="submit">Continue</button>
            </form>
        </div>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. Amazon:</p><ul><li>Always has the orange/yellow "Continue" button</li><li>URL is amazon.com</li></ul>`,
		},
		{
			Name: "GitHub",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign in to GitHub</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #0d1117; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; color: #c9d1d9; }
        .box { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 32px; width: 300px; }
        .logo { text-align: center; font-size: 40px; margin-bottom: 16px; }
        h2 { font-size: 24px; font-weight: 300; margin: 0 0 16px; text-align: center; }
        input { width: 100%; padding: 8px 12px; margin: 8px 0; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; color: #c9d1d9; font-size: 14px; box-sizing: border-box; }
        button { width: 100%; padding: 8px 16px; margin: 16px 0; background: #238636; border: 1px solid #238636; border-radius: 6px; color: white; font-size: 14px; font-weight: 500; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">⬛</div>
        <h2>Sign in to GitHub</h2>
        <form method="POST">
            <input type="text" name="username" placeholder="Username or email address" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. Check:</p><ul><li>URL must be github.com</li><li>GitHub uses dark theme consistently</li></ul>`,
		},
		{
			Name: "Facebook",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Log in to Facebook</title>
    <style>
        body { font-family: Helvetica, Arial, sans-serif; background: #f0f2f5; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .container { display: flex; gap: 20px; align-items: center; }
        .left { max-width: 500px; }
        .left h1 { color: #1877f2; font-size: 48px; margin: 0 0 10px; }
        .left p { color: #050505; font-size: 28px; margin: 0; }
        .box { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1), 0 8px 16px rgba(0,0,0,0.1); width: 380px; }
        input { width: 100%; padding: 14px 16px; margin: 10px 0; border: 1px solid #dddfe2; border-radius: 6px; font-size: 17px; box-sizing: border-box; }
        button { width: 100%; padding: 14px 16px; background: #1877f2; border: none; border-radius: 6px; color: white; font-size: 20px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <div class="left">
            <h1>facebook</h1>
            <p>Connect with friends and the world.</p>
        </div>
        <div class="box">
            <form method="POST">
                <input type="text" name="email" placeholder="Email or phone number" required>
                <input type="password" name="password" placeholder="Password" required>
                <button type="submit">Log in</button>
            </form>
        </div>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. Facebook never:</p><ul><li>Asks for password in a popup</li><li>Has this split layout</li></ul>`,
		},
		{
			Name: "Dropbox",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign in - Dropbox</title>
    <style>
        body { font-family: 'Proxima Nova', sans-serif; background: #f0f0f0; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }
        .box { background: white; padding: 40px; border-radius: 8px; width: 360px; text-align: center; }
        .logo { color: #0061ff; font-size: 40px; margin-bottom: 20px; }
        h2 { color: #333; font-size: 24px; margin: 0 0 10px; }
        p { color: #666; font-size: 14px; margin-bottom: 20px; }
        input { width: 100%; padding: 10px; margin: 8px 0; border: 1px solid #dbdbdb; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { width: 100%; padding: 10px; background: #0061ff; color: white; border: none; border-radius: 4px; font-size: 14px; font-weight: bold; cursor: pointer; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">◈</div>
        <h2>Sign in</h2>
        <p>Continue to Dropbox</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "Twitter / X",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Sign in to X</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #000; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; color: white; }
        .box { width: 350px; padding: 32px; text-align: center; }
        .logo { font-size: 36px; margin-bottom: 30px; }
        h2 { font-size: 31px; font-weight: 700; margin: 0 0 30px; }
        input { width: 100%; padding: 12px; margin: 10px 0; background: #000; border: 1px solid #536471; border-radius: 4px; color: white; font-size: 15px; box-sizing: border-box; }
        button { width: 100%; padding: 14px; background: white; color: black; border: none; border-radius: 999px; font-size: 15px; font-weight: bold; cursor: pointer; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">X</div>
        <h2>Sign in to X</h2>
        <form method="POST">
            <input type="text" name="phone" placeholder="Phone, email, or username" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "Netflix",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Netflix - Sign In</title>
    <style>
        body { background: #141414; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif; }
        .box { width: 100%; max-width: 440px; padding: 60px 68px 40px; background: rgba(0,0,0,0.75); border-radius: 4px; }
        h2 { color: white; font-size: 32px; margin: 0 0 28px; }
        input { width: 100%; padding: 16px 20px; margin: 16px 0; background: #333; border: none; border-radius: 4px; color: white; font-size: 16px; box-sizing: border-box; }
        button { width: 100%; padding: 16px; background: #e50914; color: white; border: none; border-radius: 4px; font-size: 16px; font-weight: bold; cursor: pointer; margin-top: 24px; }
    </style>
</head>
<body>
    <div class="box">
        <h2>Sign In</h2>
        <form method="POST">
            <input type="email" name="email" placeholder="Email or phone number" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign In</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "PayPal",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>PayPal: Log in to your account</title>
    <style>
        body { background: #f0f0f0; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'PayPal Sans', sans-serif; }
        .box { background: white; padding: 40px; border-radius: 8px; width: 380px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 24px; color: #003087; font-size: 28px; font-weight: bold; }
        h2 { color: #333; font-size: 26px; margin: 0 0 8px; }
        p { color: #666; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 14px; margin: 10px 0; border: 1px solid #a0a0a0; border-radius: 4px; font-size: 16px; box-sizing: border-box; }
        button { width: 100%; padding: 14px; background: #0070ba; color: white; border: none; border-radius: 25px; font-size: 16px; font-weight: bold; cursor: pointer; }
        button:hover { background: #005ea6; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">PayPal</div>
        <h2>Log in to your account</h2>
        <p>Send payments, track sales, and more</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email or mobile number" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Log In</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. PayPal:</p><ul><li>Does not ask to log in from email</li><li>URL must be paypal.com</li></ul>`,
		},
		{
			Name: "Adobe",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign in to Adobe</title>
    <style>
        body { background: #f5f5f5; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif; }
        .box { background: white; padding: 40px; border-radius: 8px; width: 400px; }
        .logo { color: #ff0000; font-size: 32px; font-weight: bold; margin-bottom: 24px; text-align: center; }
        h2 { color: #333; font-size: 24px; margin: 0 0 8px; }
        p { color: #666; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #bfbfbf; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #1473e6; color: white; border: none; border-radius: 4px; font-size: 16px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">Adobe</div>
        <h2>Sign in</h2>
        <p>Get access to all your Creative Cloud apps</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email address" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "Slack",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign in to Slack</title>
    <style>
        body { background: #f8f8f8; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Lato', sans-serif; }
        .box { background: white; padding: 44px 40px; border-radius: 6px; width: 400px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .logo { color: #36c5f0; font-size: 34px; margin-bottom: 24px; text-align: center; font-weight: bold; }
        h2 { color: #1d1c1d; font-size: 22px; margin: 0 0 8px; }
        p { color: #616061; font-size: 15px; margin-bottom: 24px; }
        input { width: 100%; padding: 12px 16px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; font-size: 15px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #611f69; color: white; border: none; border-radius: 4px; font-size: 15px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">💬</div>
        <h2>Sign in to Slack</h2>
        <p>We couldn't find an account with that email address.</p>
        <form method="POST">
            <input type="email" name="email" placeholder="name@workspace.com" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign in</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "Zoom",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign In - Zoom</title>
    <style>
        body { background: #fafafa; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Inter', sans-serif; }
        .box { background: white; padding: 40px; border-radius: 8px; width: 400px; text-align: center; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .logo { color: #2d8cff; font-size: 36px; margin-bottom: 24px; }
        h2 { color: #333; font-size: 24px; margin: 0 0 8px; }
        p { color: #666; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #2d8cff; color: white; border: none; border-radius: 4px; font-size: 14px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo">Zoom</div>
        <h2>Sign In</h2>
        <p>Enter your credentials to continue</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email address" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Sign In</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
		},
		{
			Name: "IT Password Reset",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Password Reset - IT Support</title>
    <style>
        body { background: #f5f5f5; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif; }
        .box { background: white; padding: 40px; border-radius: 8px; width: 400px; border-top: 4px solid #d32f2f; }
        .icon { background: #ffebee; color: #d32f2f; width: 48px; height: 48px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 24px; margin-bottom: 20px; }
        h2 { color: #333; font-size: 24px; margin: 0 0 12px; }
        p { color: #666; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #d32f2f; color: white; border: none; border-radius: 4px; font-size: 14px; font-weight: bold; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="icon">🔐</div>
        <h2>Password Reset</h2>
        <p>Your IT administrator has initiated a password reset. Enter your credentials below.</p>
        <form method="POST">
            <input type="text" name="username" placeholder="Username" required>
            <input type="password" name="new_password" placeholder="New Password" required>
            <input type="password" name="confirm_password" placeholder="Confirm New Password" required>
            <button type="submit">Reset Password</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test. IT departments never:</p><ul><li>Email password reset links</li><li>Ask for passwords via email</li></ul>`,
		},
		{
			Name: "Office 365",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <title>Sign in to Office 365</title>
    <style>
        body { background: #fff; margin: 0; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Segoe UI', sans-serif; }
        .box { width: 400px; padding: 32px; }
        .logo { width: 108px; margin-bottom: 16px; }
        h2 { color: #323130; font-size: 24px; font-weight: 600; margin: 0 0 8px; }
        p { color: #605e5c; font-size: 14px; margin-bottom: 24px; }
        input { width: 100%; padding: 8px 12px; margin: 8px 0; border: 1px solid #8a8886; border-radius: 2px; font-size: 14px; box-sizing: border-box; }
        button { padding: 8px 24px; background: #0067b8; color: white; border: none; font-size: 14px; font-weight: 600; cursor: pointer; }
    </style>
</head>
<body>
    <div class="box">
        <div class="logo"><img src="https://img-prod-cms-rt-microsoft-com.akamaized.net/cms/blb9QGBq2Z5TqE4F4.png" width="108"></div>
        <h2>Sign in</h2>
        <p>Enter the email and password you use with your Microsoft account.</p>
        <form method="POST">
            <input type="email" name="email" placeholder="Email" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Next</button>
        </form>
    </div>
</body>
</html>`,
			CaptureCredentials: true,
			TrainingContent:    `<h2>Security Notice</h2><p>This was a simulated phishing test.</p>`,
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
		FromEmail: "noreply@redhook.local",
		FromName:  "RedHook",
		UseTLS:    true,
		IsActive:  true,
	}
	db.Create(&smtpConfig)

	log.Println("Database seeded successfully")
	return nil
}
