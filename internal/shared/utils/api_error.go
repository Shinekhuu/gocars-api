package utils

import "net/http"

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *ApiError) Error() string {
	return e.Message
}

var (
	ErrRateLimit       = &ApiError{Code: "RATE_LIMIT", Message: "Шинэ баталгаажуулах код хүсэхийн өмнө түр хүлээнэ үү.", Status: http.StatusTooManyRequests}
	ErrUserNotFound    = &ApiError{Code: "USER_NOT_FOUND", Message: "Хэрэглэгч олдсонгүй.", Status: http.StatusNotFound}
	ErrAlreadyVerified = &ApiError{Code: "ALREADY_VERIFIED", Message: "Хэрэглэгч бүртгэлтэй байна.", Status: http.StatusBadRequest}
	ErrInvalidOtp      = &ApiError{Code: "INVALID_OTP", Message: "Баталгаажуулах код буруу байна.", Status: http.StatusBadRequest}
	ErrOtpExpired      = &ApiError{Code: "OTP_EXPIRED", Message: "Баталгаажуулах кодны хугацаа дууссан байна.", Status: http.StatusBadRequest}
)

func NewInternalError(message string) *ApiError {
	return &ApiError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}
