resource "elasticsearch_xpack_role_mapping" "test" {
  role_mapping_name = "test"
  roles = [
    "admin",
    "user",
  ]
  enabled = true
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
}
