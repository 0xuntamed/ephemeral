data "aws_iam_policy_document" "ecs_tasks_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "task_execution" {
  name               = "${var.project_name}-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}

resource "aws_iam_role_policy_attachment" "task_execution" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "database_secret" {
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_db_instance.platform.master_user_secret[0].secret_arn]
  }
}

resource "aws_iam_role_policy" "database_secret" {
  name   = "database-secret"
  role   = aws_iam_role.task_execution.id
  policy = data.aws_iam_policy_document.database_secret.json
}

resource "aws_iam_role" "task" {
  name               = "${var.project_name}-preview-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks_assume_role.json
}
