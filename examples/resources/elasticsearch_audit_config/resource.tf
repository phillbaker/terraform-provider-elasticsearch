resource "elasticsearch_opensearch_audit_config" "test" {
  enabled = true

  audit {
    enable_rest              = true
    disabled_rest_categories = ["GRANTED_PRIVILEGES", "AUTHENTICATED"]

    enable_transport              = true
    disabled_transport_categories = ["GRANTED_PRIVILEGES", "AUTHENTICATED"]

    resolve_bulk_requests = true
    log_request_body      = true
    resolve_indices       = true

    # Note: if set false, AWS OpenSearch will return HTTP 409 (Conflict)
    exclude_sensitive_headers = true

    ignore_users    = ["kibanaserver"]
    ignore_requests = ["SearchRequest", "indices:data/read/*", "/_cluster/health"]
  }

  compliance {
    enabled = true

    # Note: if both internal/external are set true, AWS OpenSearch will return HTTP 409 (Conflict)
    internal_config = true
    external_config = false

    read_metadata_only = true
    read_ignore_users  = ["read-ignore-1"]

    read_watched_field {
      index  = "read-index-1"
      fields = ["field-1", "field-2"]
    }

    read_watched_field {
      index  = "read-index-2"
      fields = ["field-3"]
    }

    write_metadata_only   = true
    write_log_diffs       = false
    write_watched_indices = ["write-index-1", "write-index-2", "log-*", "*"]
    write_ignore_users    = ["write-ignore-1"]
  }
}
