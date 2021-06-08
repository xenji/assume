package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type AssumeProfile struct {
	Profile                string `yaml:"profile"`
	Role                   string `yaml:"role_to_assume"`
	AwsMainAccountName     string `yaml:"aws_main_account_name"`
	AwsMainAccountNumber   int    `yaml:"aws_main_account_number"`
	AwsMainAccountUser     string `yaml:"aws_main_account_user"`
	AwsTargetAccountName   string `yaml:"aws_target_account_name"`
	AwsTargetAccountNumber int    `yaml:"aws_target_account_number"`
	MfaToken               string `yaml:"mfa_token"`
}

type AssumeConfig struct {
	Profiles []AssumeProfile `yaml:"profiles"`
}

func readConfig(path string) AssumeConfig {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Could not read config file. Did you create one? (#%v)", err)
	}

	var assumeConfig AssumeConfig
	err = yaml.Unmarshal(yamlFile, &assumeConfig)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return assumeConfig
}

func flattenProfileNames(profiles []AssumeProfile) string {
	var profileNames []string
	for _, v := range profiles {
		profileNames = append(profileNames, v.Profile)
	}
	return strings.Join(profileNames, "\n")
}
