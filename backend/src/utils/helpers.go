package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

func GenerateToken(userID uint, campaignID uint) string {
	timestamp := time.Now().Unix()
	raw := fmt.Sprintf("%d:%d:%d", userID, campaignID, timestamp)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func DecodeToken(token string) (userID uint, campaignID uint, err error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, 0, err
	}

	var userIDInt, campaignIDInt, timestamp int64
	_, err = fmt.Sscanf(string(decoded), "%d:%d:%d", &userIDInt, &campaignIDInt, &timestamp)
	if err != nil {
		return 0, 0, err
	}

	return uint(userIDInt), uint(campaignIDInt), nil
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func ParseTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func CalculateRiskScore(clicks int, submissions int, reports int) int {
	score := (clicks * 2) + (submissions * 5) - (reports * 3)
	if score < 0 {
		return 0
	}
	return score
}

func GetRiskLevel(score int) string {
	if score > 15 {
		return "high"
	}
	if score > 5 {
		return "medium"
	}
	return "low"
}

func ReplaceTemplateVariables(html string, vars map[string]string) string {
	result := html
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = ReplaceString(result, placeholder, value)
	}
	return result
}

func ReplaceString(s, old, new string) string {
	result := s
	for i := 0; i < len(result); i++ {
		if len(result)-i >= len(old) && result[i:i+len(old)] == old {
			result = result[:i] + new + result[i+len(old):]
		}
	}
	return result
}
