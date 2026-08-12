locals {
  name = substr(var.environment_id, 0, 32)
  tags = merge(var.tags, {
    ManagedBy     = "ephemeral-platform"
    EnvironmentID = var.environment_id
    Ephemeral     = "true"
  })

  database_secrets = var.database_password_secret_arn == "" ? [] : [
    {
      name      = "DB_PASSWORD"
      valueFrom = "${var.database_password_secret_arn}:password::"
    }
  ]
}

resource "aws_cloudwatch_log_group" "preview" {
  name              = "/ephemeral/${var.environment_id}"
  retention_in_days = 7
  tags              = local.tags
}

resource "aws_lb_target_group" "preview" {
  name        = local.name
  port        = var.container_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = var.vpc_id

  deregistration_delay = 15

  health_check {
    enabled             = true
    path                = var.health_check_path
    healthy_threshold   = 2
    unhealthy_threshold = 3
    interval            = 15
    timeout             = 5
    matcher             = "200-399"
  }

  tags = local.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_listener_rule" "preview" {
  listener_arn = var.https_listener_arn
  priority     = var.listener_rule_priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.preview.arn
  }

  condition {
    host_header {
      values = [var.host_name]
    }
  }

  tags = local.tags
}

resource "aws_ecs_task_definition" "preview" {
  family                   = local.name
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = tostring(var.cpu)
  memory                   = tostring(var.memory)
  execution_role_arn       = var.task_execution_role_arn
  task_role_arn            = var.task_role_arn

  container_definitions = jsonencode([
    {
      name      = "application"
      image     = var.image
      essential = true
      portMappings = [{
        containerPort = var.container_port
        hostPort      = var.container_port
        protocol      = "tcp"
      }]
      environment = [
        { name = "PORT", value = tostring(var.container_port) },
        { name = "ENVIRONMENT_ID", value = var.environment_id },
        { name = "DB_HOST", value = var.database_host },
        { name = "DB_NAME", value = var.database_name },
        { name = "DB_USER", value = var.database_user },
        { name = "DB_SCHEMA", value = var.database_schema }
      ]
      secrets = local.database_secrets
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.preview.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "app"
        }
      }
    }
  ])

  tags = local.tags
}

resource "aws_ecs_service" "preview" {
  name            = var.environment_id
  cluster         = var.ecs_cluster_arn
  task_definition = aws_ecs_task_definition.preview.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  enable_ecs_managed_tags           = true
  propagate_tags                    = "SERVICE"
  health_check_grace_period_seconds = 30

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [var.service_security_group_id]
    assign_public_ip = var.assign_public_ip
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.preview.arn
    container_name   = "application"
    container_port   = var.container_port
  }

  depends_on = [aws_lb_listener_rule.preview]
  tags       = local.tags
}
