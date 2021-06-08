package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	gookit "github.com/gookit/ini/v2"
	"log"
	"os"
	"strings"
	"time"
)

func readAwsConfigFileGkit(file string) (*gookit.Ini, error) {
	iniData := gookit.New()
	err := iniData.LoadExists(file)
	if err != nil {
		return nil, err
	}
	return iniData, nil

}

func requestCredentials(profile AssumeProfile, useToken bool, token string) (*sts.AssumeRoleOutput, error) {
	_ = os.Setenv("AWS_PROFILE", profile.AwsMainAccountName)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	svc := sts.NewFromConfig(cfg)

	roleArn := fmt.Sprintf("arn:aws:iam::%d:role/%s", profile.AwsTargetAccountNumber, profile.Role)
	roleSessionName := fmt.Sprintf(
		"%d-%s-%s",
		profile.AwsTargetAccountNumber,
		profile.AwsTargetAccountName,
		strings.Replace(profile.Role, "\\", "_", -1))

	serialNumber := fmt.Sprintf(
		"arn:aws:iam::%d:mfa/%s",
		profile.AwsMainAccountNumber,
		profile.AwsMainAccountUser)

	var tokenCode string

	if useToken {
		log.Println("using TOPT passed as argument")
		tokenCode = token
	} else {
		if profile.MfaToken != "" {
			log.Println("using mfa_token from config to generate TOTP")
			tokenCode, err = otpFromSecret(profile.MfaToken)
		}
	}

	assumeRoleInput := sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &roleSessionName,
	}

	if tokenCode != "" {
		assumeRoleInput.TokenCode = &tokenCode
		assumeRoleInput.SerialNumber = &serialNumber
	}

	if err != nil {
		return nil, err
	}

	return svc.AssumeRole(ctx, &assumeRoleInput)
}

func configureProfile(targetAccountName string, accessKey, secretKey, token *string, file *gookit.Ini) {

	section := file.Section(targetAccountName)

	section["aws_access_key_id"] = *accessKey
	section["aws_secret_access_key"] = *secretKey
	section["aws_session_token"] = *token
}
