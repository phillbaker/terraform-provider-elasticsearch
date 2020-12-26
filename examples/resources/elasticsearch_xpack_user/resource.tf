resource "elasticsearch_xpack_user" "test" {
	username = "johndoe"
	fullname = "John DoDo"
	email    = "john@do.com"
	password = "secret"
	roles    = ["admin"]
  metadata = <<-EOF
  {
    "foo": "bar"
  }
  EOF
}
