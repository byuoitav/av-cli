terraform {
  backend "s3" {
    bucket     = "terraform-state-storage-586877430255"
    lock_table = "terraform-state-lock-586877430255"
    region     = "us-west-2"

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

module "api" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "cli-api"
  image          = "docker.pkg.github.com/byuoitav/av-cli/api-dev"
  image_version  = "23a091f"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/av-cli"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["cli.av.byu.edu"]
  health_check      = false
  container_env     = {}
  container_args = [
    "--port", "8080",
    "--log-level", "0",
    "--auth-addr", "idk",
    "--auth-token", "idk",
    "--disable-auth"
  ]
  ingress_annotations = {
    "nginx.ingress.kubernetes.io/backend-protocol" = "GRPC"
    "nginx.ingress.kubernetes.io/server-snippet"   = <<EOF
		grpc_read_timeout 3600s;
		EOF
  }
}

module "slack_cli" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "slack-cli"
  image          = "docker.pkg.github.com/byuoitav/av-cli/slack-dev"
  image_version  = "23a091f"
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
    "--avcli-token", "put-token-here"
  ]
  ingress_annotations = {}
}
