terragrunt = {
  include {
    path = "${find_in_parent_folders()}"
  }

  import_files "elasticsearch" {
    source              = "${var.outputs_folder}/infrastructure-logging/"
    import_into_modules = true
    files               = ["elasticsearch*"]
  }
}

