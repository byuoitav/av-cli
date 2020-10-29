terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-586877430255"
    dynamodb_table = "terraform-state-lock-586877430255"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "av-cli.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-cluster-endpoint"
}

provider "kubernetes" {
  host = data.aws_ssm_parameter.eks_cluster_endpoint.value
}

data "aws_ssm_parameter" "cli_client_id" {
  name = "/env/avcli/client-id"
}

data "aws_ssm_parameter" "cli_client_secret" {
  name = "/env/avcli/client-secret"
}

data "aws_ssm_parameter" "db_address" {
  name = "/env/couch-address"
}
data "aws_ssm_parameter" "db_username" {
  name = "/env/couch-username"
}

data "aws_ssm_parameter" "db_password" {
  name = "/env/couch-password"
}

data "aws_ssm_parameter" "pi_password" {
  name = "/env/pi-password"
}

data "aws_ssm_parameter" "auth_token" {
  name = "/env/avcli/opa-token"
}

data "aws_ssm_parameter" "auth_addr" {
  name = "/env/auth-addr"
}

module "api" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "cli-api"
  image          = "docker.pkg.github.com/byuoitav/av-cli/api-dev"
  image_version  = "aba921a"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/av-cli"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["cli.av.byu.edu"]
  health_check      = false
  container_env = {
    "DB_ADDRESS"       = data.aws_ssm_parameter.db_address.value
    "DB_USERNAME"      = data.aws_ssm_parameter.db_username.value
    "DB_PASSWORD"      = data.aws_ssm_parameter.db_password.value
    "STOP_REPLICATION" = "true"
    "PI_PASSWORD"      = data.aws_ssm_parameter.pi_password.value
  }
  container_args = [
    "--port", "8080",
    "--log-level", "-1",
    "--auth-addr", data.aws_ssm_parameter.auth_addr.value,
    "--auth-token", data.aws_ssm_parameter.auth_token.value,
    "--gateway-addr", "api.byu.edu",
    "--client-id", data.aws_ssm_parameter.cli_client_id.value,
    "--client-secret", data.aws_ssm_parameter.cli_client_secret.value
  ]
  ingress_annotations = {
    "nginx.ingress.kubernetes.io/backend-protocol" = "GRPC"
    "nginx.ingress.kubernetes.io/server-snippet"   = <<EOF
		grpc_read_timeout 3600s;
		EOF
  }
}

data "aws_ssm_parameter" "avcli_token" {
  name = "/env/slack/avcli-token"
}

data "aws_ssm_parameter" "slack_signing_secret" {
  name = "/env/slack/slack-signing-secret"
}

data "aws_ssm_parameter" "slack_token" {
  name = "/env/slack/slack-token"
}

module "slack_cli" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "slack-cli"
  image          = "docker.pkg.github.com/byuoitav/av-cli/slack-dev"
  image_version  = "aba921a"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/av-cli"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["slack-cli.av.byu.edu"]
  container_env     = {}
  container_args = [
    "--port", "8080",
    "--log-level", "0",
    "--avcli-api", "cli.av.byu.edu:443",
    "--avcli-token", data.aws_ssm_parameter.avcli_token.value,
    "--signing-secret", data.aws_ssm_parameter.slack_signing_secret.value,
    "--slack-token", data.aws_ssm_parameter.slack_token.value,
  ]
  ingress_annotations = {}
}
