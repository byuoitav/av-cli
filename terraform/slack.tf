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
  image_version  = "v0.1.4-alpha"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/av-cli"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["slack-cli.av.byu.edu"]
  health_check      = false
  container_env     = {}
  container_args = [
    "--port", "8080",
    "--log-level", "0",
    "--avcli-api", "cli.av.byu.edu:443", // TODO just point at cli-api
    "--avcli-token", data.aws_ssm_parameter.avcli_token.value,
    "--signing-secret", data.aws_ssm_parameter.slack_signing_secret.value,
    "--slack-token", data.aws_ssm_parameter.slack_token.value,
  ]
  ingress_annotations = {}
}
