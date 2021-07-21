# Assume

## Disclaimer

I've developed this tool together with my valued colleagues in my time at Breuninger.

## Example Config

This config goes to `$HOME/.assume.yml`

```yaml
profiles:
  - profile: dev
    role_to_assume: a_developer
    aws_main_account_name: some-iam
    aws_main_account_number: 5678900
    aws_main_account_user: foobar
    aws_target_account_name: some-dev
    aws_target_account_number: 123456789
    mfa_token: XXXXX
```
## Usage

    assume profile $youProfileNameFromTheConfig
