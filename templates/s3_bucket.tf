resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name
  tags = merge(
    {
      Environment = "dev"
    },
    var.tags,
  )
}
resource "port-labs_entity" "bucket" {
  properties {
    name = "bucket"
    value = aws_s3_bucket.bucket.bucket
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
variable "tags" {
  type = map(string)
  description = "A map of tags to add to all resources"
}