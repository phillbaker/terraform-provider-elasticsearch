project = "terraform-provider-project"
 
source_path = "../modules"
 
terragrunt = {
  include {
    source = "${discover("bootstrap", default(var.env, "default"), default(var.region, "us-east-1"))}"
  }

  assume_role = "${var.deploy_role_security}"
}