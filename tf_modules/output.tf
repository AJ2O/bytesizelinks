# --- General ---
output "codestar_connection" {
  value = format("https://console.aws.amazon.com/codesuite/settings/%s/%s/connections/%s",
    data.aws_caller_identity.current.account_id,
    data.aws_region.current.name,
  split("/", aws_codestarconnections_connection.github.arn)[1])
  description = "Accept this CodeStar connection for GitHub authentication."
}
output "pipeline_webclient" {
  value = format("https://console.aws.amazon.com/codesuite/codepipeline/pipelines/%s/view?region=%s",
    aws_codepipeline.web_client.name,
  data.aws_region.current.name)
  description = "The link to the pipeline for the web client."
}
output "app_link" {
    value = format("https://%s", var.domain_name)
    description = "The link to the home page of the app."
}