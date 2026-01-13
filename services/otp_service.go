package services

import (
	"context"
	"fmt"
	"gocars-api/database"
	"gocars-api/models"
	"gocars-api/utils"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"gopkg.in/mail.v2"
	"gorm.io/gorm"
)

// GenerateAndSendOtp generates a 4-digit OTP, upserts it in DB, and sends email
func GenerateAndSendOtp(email string) error {
	verificationCode := fmt.Sprintf("%04d", rand.Intn(10000))
	expiration := time.Now().Add(10 * time.Minute)

	var otp models.Otp

	// Check if OTP exists (including soft-deleted)
	err := database.DB.Unscoped().Where("email = ?", email).First(&otp).Error

	switch {
	case err == nil:
		// Record found — update existing OTP and revive if soft-deleted
		otp.VerificationCode = verificationCode
		otp.ExpiresAt = expiration
		otp.DeletedAt = gorm.DeletedAt{}
		if saveErr := database.DB.Save(&otp).Error; saveErr != nil {
			log.Printf("[OTP] Failed to update OTP for %s: %v", email, saveErr)
			return utils.NewInternalError("failed to update OTP")
		}
		log.Printf("[OTP] Updated OTP for %s", email)

	case err == gorm.ErrRecordNotFound:
		// Record not found — create new OTP
		otp = models.Otp{
			Email:            email,
			VerificationCode: verificationCode,
			ExpiresAt:        expiration,
		}
		if createErr := database.DB.Create(&otp).Error; createErr != nil {
			log.Printf("[OTP] Failed to create OTP for %s: %v", email, createErr)
			return utils.NewInternalError("failed to create OTP")
		}
		log.Printf("[OTP] Created new OTP for %s", email)

	default:
		// Any other DB error
		log.Printf("[OTP] DB error when generating OTP for %s: %v", email, err)
		return utils.NewInternalError("database error")
	}

	// Send email
	if err := sendOtpEmailUsingAPI(email, verificationCode); err != nil {
		log.Printf("[OTP] Failed to send OTP email to %s: %v", email, err)
		return err
	}
	log.Printf("[OTP] OTP email sent to %s", email)

	return nil
}

// sendOtpEmailUsingAPI sends the OTP email using Mailgun API
func sendOtpEmailUsingAPI(email, code string) error {
	apiKey := os.Getenv("MAILGUN_SENDING_API_KEY")
	domain := "support.gocars.mn"

	mg := mailgun.NewMailgun(domain, apiKey)

	// Create HTML email body
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

	// Use package-level NewMessage function
	m := mailgun.NewMessage(
		"Go Cars LLC <no-reply@support.gocars.mn>", // From
		"Your Verification Code",                   // Subject
		"",                                         // Plain text fallback (can be empty)
		email,                                      // To
	)

	// Set the HTML body
	m.SetHTML(body)

	// Send the email with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, id, err := mg.Send(ctx, m)
	if err != nil {
		return err
	}

	fmt.Println("Email sent successfully! ID:", id, "Response:", resp)
	return nil
}

func sendOtpEmail(email, code string) error {
	username := os.Getenv("SMTP_USERNAME")
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

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

	var lastErr error
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		m := mail.NewMessage()
		m.SetHeader("From", "Go Cars LLC <"+username+">")
		m.SetHeader("To", email)
		m.SetHeader("Subject", "Your Verification Code 🔒")
		m.SetBody("text/html", body)

		d := mail.NewDialer(os.Getenv("SMTP_HOST"), port, username, os.Getenv("SMTP_PASSWORD"))
		if err := d.DialAndSend(m); err != nil {
			lastErr = err
			log.Printf("[OTP] Attempt %d: Failed to send OTP to %s: %v", attempt, email, err)
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("[OTP] OTP email sent to %s on attempt %d", email, attempt)
		return nil
	}

	return fmt.Errorf("failed to send OTP email to %s after %d attempts: %v", email, maxRetries, lastErr)
}

// VerifyOtp checks if the OTP is correct and not expired
func VerifyOtp(email, verificationCode string) error {
	var otp models.Otp

	err := database.DB.Unscoped().
		Where("email = ? AND verification_code = ?", email, verificationCode).
		First(&otp).Error

	switch {
	case err != nil:
		log.Printf("[OTP] Verification failed for %s: invalid OTP", email)
		return utils.ErrInvalidOtp

	case time.Now().After(otp.ExpiresAt):
		log.Printf("[OTP] Verification failed for %s: OTP expired", email)
		return utils.ErrOtpExpired

	default:
		if err := database.DB.Delete(&otp).Error; err != nil {
			log.Printf("[OTP] Failed to delete OTP for %s: %v", email, err)
		}
		log.Printf("[OTP] OTP verified successfully for %s", email)
		return nil
	}
}

// ResendOtp generates a new OTP and sends it if allowed by rate-limit
func ResendOtp(email string) error {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("[OTP] Resend failed: user not found %s", email)
		return utils.ErrUserNotFound
	}
	if user.IsVerified {
		log.Printf("[OTP] Resend failed: user already verified %s", email)
		return utils.ErrAlreadyVerified
	}

	var lastOtp models.Otp
	err := database.DB.Where("email = ?", email).Order("updated_at desc").First(&lastOtp).Error
	if err == nil && time.Since(lastOtp.UpdatedAt) < time.Minute {
		log.Printf("[OTP] Resend rate limit: please wait before requesting new OTP for %s", email)
		return utils.ErrRateLimit
	}

	if err := GenerateAndSendOtp(user.Email); err != nil {
		log.Printf("[OTP] Failed to resend OTP for %s: %v", email, err)
		return err
	}

	log.Printf("[OTP] OTP resent successfully to %s", email)
	return nil
}

// CleanExpiredOtps deletes expired OTPs periodically
func CleanExpiredOtps() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			result := database.DB.Unscoped().Where("expires_at < ?", time.Now()).Delete(&models.Otp{})
			if result.Error != nil {
				log.Printf("[OTP] Failed to clean expired OTPs: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("[OTP] Cleaned %d expired OTP(s)", result.RowsAffected)
			}
		}
	}()
}
