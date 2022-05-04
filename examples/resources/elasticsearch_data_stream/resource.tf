resource "elasticsearch_composable_index_template" "foo" {
  name = "foo-template"
  body = <<EOF
{
  "index_patterns": ["foo-data-stream*"],
  "data_stream": {}
}
EOF
}

resource "elasticsearch_data_stream" "foo" {
  name       = "foo-data-stream"
  depends_on = [elasticsearch_composable_index_template.foo]
}
