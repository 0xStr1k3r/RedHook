# PhishGuard - Phishing Simulation & Awareness Platform

A controlled phishing simulation platform for measuring and improving organizational security awareness.

## Features

- **Campaign Management:** Create, launch, and track phishing campaigns
- **Template Library:** Pre-built templates (IT, HR, Finance, etc.)
- **Landing Pages:** Customizable fake login pages
- **Analytics:** Open rate, click rate, submission rate, reporting rate
- **Risk Scoring:** User risk classification (Low/Medium/High)
- **Training:** Post-click awareness training
- **NIST Phish Scale:** Difficulty rating integration

## Quick Start

### Prerequisites
- Go 1.26+
- PostgreSQL 15+
- Node.js 18+

### Backend Setup

```bash
cd backend
cp .env.example .env
# Edit .env with your settings

go build -o phishguard ./src
./phishguard
```

### Frontend

```bash
# Serve frontend/index.html
```

### Docker

```bash
docker-compose up -d
```

## Default Credentials

- Email: admin@phishguard.local
- Password: changeme123

## API Endpoints

### Auth
- POST /api/auth/login
- POST /api/auth/register
- GET /api/auth/me

### Campaigns
- GET /api/campaigns
- POST /api/campaigns
- POST /api/campaigns/:id/launch
- GET /api/campaigns/:id/stats

### Users
- GET /api/users
- POST /api/users
- GET /api/users/departments

### Templates
- GET /api/templates
- POST /api/templates

### Analytics
- GET /api/analytics/overview

## Tracking Endpoints

- GET /track/open/:token - Track email opens
- GET /track/click/:token - Track link clicks
- GET /landing/:token - Display landing page
- GET /train/:token - Show training

## Ethical Guidelines

This tool must only be used with:
- Written authorization from management
- Employee awareness of testing program
- No real credential storage
- Educational purpose only

## Technology Stack

- Backend: Go + Gin
- Database: PostgreSQL
- Frontend: Vanilla JS

## License

MIT - For educational purposes