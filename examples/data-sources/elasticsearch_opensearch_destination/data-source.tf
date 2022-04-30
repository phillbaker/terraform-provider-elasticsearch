# Example destination in other terraform plan
# resource "elasticsearch_opensearch_destination" "test" {
#   body = <<EOF
# {
#   "name": "my-destination",
#   "type": "slack",
#   "slack": {
#     "url": "http://www.example.com"
#   }
# }
# EOF
# }

data "elasticsearch_opensearch_destination" "test" {
  name = "my-destination"
}
