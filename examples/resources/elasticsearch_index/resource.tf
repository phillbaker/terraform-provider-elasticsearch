# Create a simple index
resource "elasticsearch_index" "test" {
  name               = "terraform-test"
  number_of_shards   = 1
  number_of_replicas = 1
  mappings           = <<EOF
{
  "people": {
    "_all": {
      "enabled": false
    },
    "properties": {
      "email": {
        "type": "text"
      }
    }
  }
}
EOF
}
