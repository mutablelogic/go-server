# Terraform test configuration

Manual testing for the `terraform-kaiak-provider`.

## Prerequisites

1. A running kaiak server:

```bash
go run ./cmd/go-server run
```

1. Build and install the provider:

```bash
go install ./cmd/terraform-kaiak-provider
```

1. Point Terraform at the local binary (one-off):

```bash
export TF_CLI_CONFIG_FILE="$PWD/.terraformrc"
```

Edit `.terraformrc` and replace `${GOBIN}` with the output of `go env GOBIN`
(or `go env GOPATH`/bin if GOBIN is empty).

## Usage

```bash
cd etc/tf

# See what Terraform will do
terraform plan

# Apply changes to the running server
terraform apply

# Tear everything down
terraform destroy
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `KAIAK_ENDPOINT` | `http://localhost:8084/api` | Base URL of the kaiak server API |
| `TF_CLI_CONFIG_FILE` | `~/.terraformrc` | Path to the CLI config with dev overrides |
