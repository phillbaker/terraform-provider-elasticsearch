provider "elasticsearch" {
    /*
    url = "${var.elasticsearch_cluster_url_protocol}${var.elasticsearch_cluster_url}" # Don't include port at the end for aws
    username = "${data.aws_ssm_parameter.cluster_username.value}"
    password = "${data.aws_ssm_parameter.cluster_password.value}"
    sniff = false # Sniffing won't work on clusters where nodes are not routable

    //insecure = true # to bypass certificate check
    */
    url = "http://localhost:9210"
    username = "elastic"
    password = "elastic"
    sniff = false
}
/*
data "aws_ssm_parameter" "cluster_username" {
  name = "/${var.env}/infra/logstack/xpack_management_username"
}

data "aws_ssm_parameter" "cluster_password" {
    name  = "/${var.env}/infra/logstack/xpack_management_password"
}
*/
/*
resource "elasticsearch_xpack_role_mapping" "role-mapping-test" {
  role_mapping_name = "testName2"
  roles = [
      "admin",
      "user",
  ]
  rules = <<-EOF
  {
    "any": [
      {
        "field": {
          "username": "esadmin"
        }
      },
      {
        "field": {
          "groups": "cn=admins,dc=example,dc=com"
        }
      }
    ]
  }
  EOF
  enabled = true
}
*/

resource "elasticsearch_xpack_role_mapping" "role-mapping-test" {
  role_mapping_name = "testName2"
  roles = [
      "admin",
      "user",
  ]
  enabled = true
  rules = "test"
}