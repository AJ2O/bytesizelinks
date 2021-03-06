# --- Data Sources ---
# ami id
data "aws_ssm_parameter" "amzn_linux_ami" {
  name = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
}

# --- Launch Template ---
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
# launch template
resource "aws_iam_instance_profile" "bsl_profile" {
  name = "bsl-webclient-profile"
  role = aws_iam_role.bsl_webclient_role.name
}
resource "aws_launch_template" "bsl_launch_template" {
  name_prefix = "bsl-launch-template"

  iam_instance_profile {
    arn = aws_iam_instance_profile.bsl_profile.arn
  }

  image_id = data.aws_ssm_parameter.amzn_linux_ami.value

  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.web_sg.id]

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "BSL Web Client"
    }
  }

  user_data = filebase64("./ec2_userdata.sh")
}

# --- Auto Scaling ---
# blue/green groups
resource "aws_autoscaling_group" "bsl_asg" {
  name             = "bsl-webclients-blue"
  max_size         = 3
  min_size         = 1
  desired_capacity = 1

  # networking
  vpc_zone_identifier = module.vpc.public_subnets
  target_group_arns   = [aws_lb_target_group.bsl_tg.arn]

  # health checks
  health_check_type         = "ELB"
  health_check_grace_period = 300

  launch_template {
    id      = aws_launch_template.bsl_launch_template.id
    version = "$Latest"
  }
}