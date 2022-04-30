resource "elasticsearch_opensearch_ism_policy_mapping" "test" {
  policy_id = "policy_1"
  indexes   = "test_index"
  state     = "delete"
}
