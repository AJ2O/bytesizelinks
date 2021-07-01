# --- Launch Template

# permissions
resource "aws_iam_role" "bsl_webclient_role" {
  name_prefix = "bsl-webclient-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}
# allows management via SSM
resource "aws_iam_role_policy_attachment" "bsl_webclient_ssm_core_attach" {
  role       = aws_iam_role.bsl_webclient_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}
resource "aws_iam_role_policy_attachment" "bsl_webclient_ssm_maintenance_attach" {
  role       = aws_iam_role.bsl_webclient_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonSSMMaintenanceWindowRole"
}
# allows for grabbing CodeBuild artifacts
resource "aws_iam_role_policy_attachment" "bsl_webclient_s3_attach" {
  role       = aws_iam_role.bsl_webclient_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
}

# autoscaling 
/*
resource "aws_launch_template" "bsl_launch_template" {
  name = "bsl-launch-template"

  instance_type = "t2.micro"

  user_data = filebase64("ec2_userdata.sh")
}*/