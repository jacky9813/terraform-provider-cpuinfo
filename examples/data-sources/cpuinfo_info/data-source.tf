terraform {
  required_providers {
    cpuinfo = {
      source  = "jacky9813/cpu-info"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>5.4.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~>2.4.0"
    }
  }
}

provider "aws" {
  profile = "my-aws-profile"
}

provider "archive" {}
provider "cpuinfo" {}

data "cpuinfo_info" "cpu_info" {}

resource "terraform_data" "lambda_zip_packager" {
  provisioner "local-exec" {
    command     = <<-EOT
      pushd ${path.module}
      [ -d .build-venv ] && rm -r .build-venv
      python3.10 -m venv .build-venv
      source .build-venv/bin/activate
      [ -d build ] && rm -r build
      pip install --target build -r requirements.txt
      cp src/lambda_function.py build/
      deactivate
      find build -name "__pycache__" -type d | xargs rm -rv
      popd
    EOT

    interpreter = ["/bin/bash", "-c"]
  }
}

data "archive_file" "lambda_source" {
  type        = "zip"
  output_file = "${path.module}/lambda_function.zip"
  source_dir  = "${path.module}/build"

  depends_on  = [
    terraform_data.lambda_zip_packager
  ]
}

resource "aws_lambda_function" "my_python_function" {
  function_name = "my_python_function"
  architectures = [data.cpuinfo_info.cpu_info.isa == "amd64" ? "x86_64" : data.cpuinfo_info.cpu_info.isa]
  runtime       = "python3.10"
  handler       = "lambda_function.lambda_handler"

  filename         = data.archive_file.lambda_source.output_path
  source_code_hash = data.archive_file.lambda_source.output_base64sha256
}
