resource "aws_sns_topic" "sns_topic" {
  name = var.topic_name
  fifo_topic = var.fifo_topic
}
resource "port-labs_entity" "sns_topic" {
  properties {
    name = "sns_topic"
    value = aws_sns_topic.sns_topic.name
  }
  properties {
    name = "url"
    value = "https://eu-west-1.console.aws.amazon.com/sns/v3/home?region=eu-west-1#/topic/${aws_sns_topic.sns_topic.arn}"
  }
  run_id = var.run_id
  blueprint = var.blueprint
  identifier = var.entity_identifier
  title = "SNS Topic ${aws_sns_topic.sns_topic.name}"
}

variable "topic_name" {
  type = string
  description = "The name of the bucket"
}
variable "fifo_topic" {
  type = bool
  description = "Whether to create a FIFO topic"
}