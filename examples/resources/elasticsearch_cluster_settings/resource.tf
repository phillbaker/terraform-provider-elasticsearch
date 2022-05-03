resource "elasticsearch_cluster_settings" "global" {
  cluster_max_shards_per_node = 10
  action_auto_create_index    = "my-index-000001,index10,-index1*,+ind*"
}
