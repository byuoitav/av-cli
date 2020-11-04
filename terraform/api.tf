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
  image_version  = "v0.1.8-alpha"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/av-cli"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["cli.av.byu.edu"]
  health_check      = false
  container_env     = {}
  container_args = [
    "--port", "8080",
    "--log-level", "debug",
    "--auth-addr", data.aws_ssm_parameter.auth_addr.value,
    "--auth-token", data.aws_ssm_parameter.auth_token.value,
    "--gateway-addr", "api.byu.edu",
    "--client-id", data.aws_ssm_parameter.cli_client_id.value,
    "--client-secret", data.aws_ssm_parameter.cli_client_secret.value,
    "--db-address", data.aws_ssm_parameter.db_address.value,
    "--db-username", data.aws_ssm_parameter.db_username.value,
    "--db-password", data.aws_ssm_parameter.db_password.value,
    "--pi-password", data.aws_ssm_parameter.pi_password.value,
  ]
  ingress_annotations = {
    "nginx.ingress.kubernetes.io/backend-protocol" = "GRPC"
    "nginx.ingress.kubernetes.io/server-snippet"   = <<EOF
		grpc_read_timeout 3600s;
		EOF
  }
}
