package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	userHome                  = os.Getenv("HOME")
	assumeConfig              = userHome + "/.assume.yml"
	awsCredentials            = userHome + "/.aws/credentials"
	selectedAwsCredentialFile string
	selectedConfigFile        string
)

func main() {

	app := &cli.App{
		Name:  "assume",
		Usage: "assume roles in different AWS accounts and write the obtained credentials in your AWS-cli config",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       assumeConfig,
				Destination: &selectedConfigFile,
				Usage:       "Load configuration from `FILE`",
			},
			&cli.StringFlag{
				Name:        "aws-credentials",
				Aliases:     []string{"ac"},
				Value:       awsCredentials,
				Destination: &selectedAwsCredentialFile,
				Usage:       "Load aws credentials from `FILE`",
			},
		},

		Commands: []*cli.Command{
			{
				Name:      "profile",
				Aliases:   []string{"p"},
				ArgsUsage: "PROFILE [MFA_TOPT]",
				Usage:     "Try to assume the role in the profile, using the config in " + assumeConfig,
				Action: func(context *cli.Context) error {
					cfg := readConfig(context.String("config"))

					if context.NArg() == 0 {
						msg := fmt.Sprintf("You need to pass a profile name, defined in your %s\n\nConfigured profiles:\n\n%s", assumeConfig, flattenProfileNames(cfg.Profiles))
						return errors.New(msg)
					}

					selectedProfile := context.Args().First()
					return assumeProfile(context, cfg, selectedProfile)
				},
			},
			{
				Name:      "watch-profile",
				ArgsUsage: "PROFILE",
				Usage:     "Try to assume the role in the profile, using the config in " + assumeConfig + " and keep the credentials alive by auto-refreshing them",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "refreshIntervalSeconds",
						Aliases: []string{"r"},
						Usage:   "",
						Value:   1800,
					},
				},
				Action: func(context *cli.Context) error {
					sigs := make(chan os.Signal, 1)

					signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
					cfg := readConfig(context.String("config"))

					if context.NArg() == 0 {
						msg := fmt.Sprintf("You need to pass a profile name, defined in your %s\n\nConfigured profiles:\n\n%s", assumeConfig, flattenProfileNames(cfg.Profiles))
						return errors.New(msg)
					}

					selectedProfile := context.Args().First()

					refreshInterval := context.Int("refreshIntervalSeconds")
					ticker := time.NewTicker(time.Second * time.Duration(refreshInterval))

					err := assumeProfile(context, cfg, selectedProfile)
					if err != nil {
						return err
					}
					log.Printf("waiting %d seconds for refresh", refreshInterval)
					for {
						select {
						case <-ticker.C:
							err := assumeProfile(context, cfg, selectedProfile)
							if err != nil {
								ticker.Stop()
								return err
							}
							log.Printf("waiting %d seconds for refresh", refreshInterval)
						case <-sigs:
							ticker.Stop()
							return errors.New("signal received")
						}
					}
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func assumeProfile(context *cli.Context, cfg AssumeConfig, selectedProfile string) error {
	var profile AssumeProfile
	profileFound := false
	for _, v := range cfg.Profiles {
		if v.Profile == selectedProfile {
			profile = v
			profileFound = true
		}
	}

	if !profileFound {
		return errors.New(fmt.Sprintf("profile %s not found in config\n\nConfigured profiles:\n\n%s", selectedProfile, flattenProfileNames(cfg.Profiles)))
	}

	log.Println("Found profile in assume.yml")

	awsCredentialFile, err := readAwsConfigFileGkit(selectedAwsCredentialFile)
	if err != nil {
		return err
	}

	log.Println("Found aws credentials file, requesting temporary credentials")
	creds, err := requestCredentials(profile, context.NArg() == 2, context.Args().Get(1))
	if err != nil {
		return err
	}

	log.Println("Received credentials, updating credentials in the credential file")
	configureProfile(
		profile.AwsTargetAccountName,
		creds.Credentials.AccessKeyId,
		creds.Credentials.SecretAccessKey,
		creds.Credentials.SessionToken,
		awsCredentialFile)

	credsFile, err := os.OpenFile(context.String("aws-credentials"), os.O_RDWR|os.O_CREATE, 600)

	if err != nil {
		return err
	}
	defer credsFile.Close()

	_, err = awsCredentialFile.WriteTo(credsFile)

	if err != nil {
		return err
	}

	log.Println("Saved credentials to " + context.String("aws-credentials"))
	return nil
}
