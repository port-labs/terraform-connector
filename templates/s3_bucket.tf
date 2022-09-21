resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name
  tags = merge(
    {
      Environment = "dev"
    },
    var.tags,
  )
}

resource "aws_s3_bucket_acl" "example_bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = var.bucket_acl
}

resource "port-labs_entity" "bucket" {
  properties {
    name = "bucket_name"
    value = aws_s3_bucket.bucket.bucket
  }
  properties {
    name = "bucket_acl"
    value = var.bucket_acl
  }
  properties {
    name = "tags"
    value = jsonencode(var.tags)
  }
  properties {
    name = "url"
    value = "https://s3.console.aws.amazon.com/s3/buckets/${aws_s3_bucket.bucket.bucket}"
  }
  run_id = var.run_id
  blueprint = var.blueprint
  identifier = var.entity_identifier
  title = "Bucket ${aws_s3_bucket.bucket.bucket}"
}

variable "bucket_name" {
  type = string
  description = "The name of the bucket"
}
variable "bucket_acl" {
  type = string
  default = "private"
  description = "The canned ACL of the bucket"
}
variable "tags" {
  type = map(string)
  description = "A map of tags to add to all resources"
}