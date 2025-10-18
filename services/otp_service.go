package services

import (
	"fmt"
	"gocars-api/config"
	"gocars-api/models"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gopkg.in/mail.v2"
)

// GenerateAndSendOtp generates a 4-digit OTP, saves it in DB, and sends email
func GenerateAndSendOtp(email string) error {
	// Generate OTP
	verificationCode := fmt.Sprintf("%04d", rand.Intn(10000))

	// Save to DB
	otp := models.Otp{
		Email:            email,
		VerificationCode: verificationCode,
		ExpiresAt:        time.Now().Add(10 * time.Minute),
	}
	if err := config.DB.Create(&otp).Error; err != nil {
		return err
	}

	// Send email
	return sendOtpEmail(email, verificationCode)
}

// sendOtpEmail sends the OTP email using SMTP
func sendOtpEmail(email, code string) error {
	username := os.Getenv("SMTP_USERNAME")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	m := mail.NewMessage()
	m.SetHeader("From", "Go Cars LLC <"+username+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Verification Code 🔒")

	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body>
			<p>Hello Dear,</p>
			<p>Your verification code is: <strong>%s</strong></p>
			<p>This code will expire in 10 minutes.</p>
			<p>Thank you,<br>Go Cars LLC</p>
		</body>
		</html>
	`, code)

	m.SetBody("text/html", body)

	d := mail.NewDialer(os.Getenv("SMTP_HOST"), port, username, os.Getenv("SMTP_PASSWORD"))
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send OTP email: %v", err)
	}

	return nil
}

// VerifyOtp checks if the OTP is correct and not expired
func VerifyOtp(email, verification_code string) error {
	var otp models.Otp
	if err := config.DB.Where("email = ? AND verification_code = ?", email, verification_code).First(&otp).Error; err != nil {
		return fmt.Errorf("invalid OTP")
	}

	if time.Now().After(otp.ExpiresAt) {
		return fmt.Errorf("OTP expired")
	}

	// Delete OTP after successful verification
	if err := config.DB.Delete(&otp).Error; err != nil {
		log.Printf("Failed to delete OTP: %v", err)
	}

	return nil
}

// CleanExpiredOtps deletes expired OTPs periodically
func CleanExpiredOtps() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			result := config.DB.Where("expires_at < ?", time.Now()).Delete(&models.Otp{})
			if result.Error != nil {
				log.Printf("Failed to clean expired OTPs: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("Cleaned %d expired OTP(s)", result.RowsAffected)
			}
		}
	}()
}

// ResendOtp generates a new OTP and sends it if allowed by rate-limit
func ResendOtp(email string) error {
	// Check if user exists and is not verified
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return fmt.Errorf("user not found")
	}
	if user.IsVerified {
		return fmt.Errorf("user already verified")
	}

	// Rate limit: allow only 1 OTP per minute
	var lastOtp models.Otp
	if err := config.DB.Where("email = ?", email).Order("created_at desc").First(&lastOtp).Error; err == nil {
		if time.Since(lastOtp.CreatedAt) < time.Minute {
			return fmt.Errorf("please wait before requesting a new OTP")
		}
	}

	// Generate new OTP
	otpCode := fmt.Sprintf("%04d", rand.Intn(10000))

	// Save OTP in DB
	newOtp := models.Otp{
		Email:            email,
		VerificationCode: otpCode,
		ExpiresAt:        time.Now().Add(10 * time.Minute),
	}
	if err := config.DB.Create(&newOtp).Error; err != nil {
		return fmt.Errorf("failed to save OTP: %v", err)
	}

	// Send OTP email
	if err := sendOtpEmail(email, otpCode); err != nil {
		log.Printf("Failed to send OTP email: %v", err)
		return fmt.Errorf("failed to send OTP email")
	}

	return nil
}
