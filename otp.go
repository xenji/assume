package main

import (
	"github.com/pquerna/otp/totp"
	"time"
)

func otpFromSecret(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
