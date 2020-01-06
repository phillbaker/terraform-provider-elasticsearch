provider "elasticsearch" {
    url = "localhost:9200"
}

resource "elasticsearch_ingest_pipeline" "filebeat_nginx_log" {
  name = "filebeat-nginx-log-v1"
  body = <<EOF
{
	"description": "Ingest pipeline for nginx log lines",
  "version": 1,
  "processors" : [
    {
      "grok": {
        "field": "message",
        "patterns": [
          """%{IPORHOST:clientip} %{USER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "%{WORD:verb} %{DATA:request} HTTP/%{NUMBER:httpversion}" %{NUMBER:response:int} (?:-|%{NUMBER:bytes:int}) %{QS:referrer} %{QS:agent}"""
        ]
      }
    },
    {
      "date": {
        "field": "timestamp",
        "formats": [
          "dd/MMM/YYYY:HH:mm:ss Z"
        ]
      }
    }
  ]
}
EOF
}

resource "elasticsearch_ingest_pipeline" "filebeat_nginx_error" {
  name = "filebeat-nginx-error-v1"
  body = <<EOF
{
	"description": "Ingest pipeline for nginx error lines",
  "version": 1,
  "processors" : [
    {
      "grok": {
        "field": "message",
        "patterns": [
          """^(?<timestamp>%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}) \[%{LOGLEVEL:severity}\] %{POSINT:pid}#%{NUMBER:threadid}\:( \*%{NUMBER:connectionid})? %{DATA:message}(,|$)( client: %{IPORHOST:client})?(, server: %{IPORHOST:server})?(, request: "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion}))")?(, upstream: "%{DATA:upstream}")?(, host: "%{IPORHOST:vhost}")?"""
        ]
      }
    },
    {
      "date": {
        "field": "timestamp",
        "formats": [
          "YYYY/MM/dd HH:mm:ss"
        ]
      }
    }
  ]
}
EOF
}
