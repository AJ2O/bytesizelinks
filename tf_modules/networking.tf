# --- Data Sources ---
data "aws_availability_zones" "available" {
  state = "available"
}
data "aws_route53_zone" "dns" {
  name = var.domain_name
}

# --- VPC ---
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "my-vpc"
  cidr = "10.0.0.0/16"

  azs = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}
resource "aws_security_group" "web_sg" {
  name_prefix = "bsl-web-sg"
  description = "Allow access from the public Internet."
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

# --- DNS ---
# route53
resource "aws_route53_record" "cert_record" {
  for_each = {
    for dvo in aws_acm_certificate.web_cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.dns.zone_id
}
resource "aws_route53_record" "web_client" {
  zone_id = data.aws_route53_zone.dns.zone_id
  name    = data.aws_route53_zone.dns.name
  type    = "A"
  alias {
    name                   = aws_lb.bsl_alb.dns_name
    zone_id                = aws_lb.bsl_alb.zone_id
    evaluate_target_health = true
  }
}

# acm
resource "aws_acm_certificate" "web_cert" {
  domain_name       = var.domain_name
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_acm_certificate_validation" "web_cert_validation" {
  certificate_arn         = aws_acm_certificate.web_cert.arn
  validation_record_fqdns = [for record in aws_route53_record.cert_record : record.fqdn]
}

# --- Load Balancing ---
# alb
resource "aws_lb" "bsl_alb" {
  name               = "bsl-webclients-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.web_sg.id]
  subnets            = module.vpc.public_subnets
}
resource "aws_lb_target_group" "bsl_alb_tg" {
  name        = "bsl-web-alb-tg"
  target_type = "instance"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = module.vpc.vpc_id

  health_check {
    protocol = "HTTP"
    path     = "/"
    matcher  = "200-299"

    interval            = 10
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }
}
resource "aws_lb_listener" "bsl_alb_listener" {
  load_balancer_arn = aws_lb.bsl_alb.arn
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  port              = "443"
  protocol          = "HTTPS"
  certificate_arn   = aws_acm_certificate.web_cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.bsl_alb_tg.arn
  }
}
resource "aws_lb_listener" "bsl_alb_listener_http" {
  load_balancer_arn = aws_lb.bsl_alb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}