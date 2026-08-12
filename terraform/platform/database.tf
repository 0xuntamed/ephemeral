resource "aws_db_subnet_group" "platform" {
  name       = var.project_name
  subnet_ids = aws_subnet.public[*].id
}

resource "aws_db_instance" "platform" {
  identifier                   = var.project_name
  engine                       = "postgres"
  instance_class               = var.database_instance_class
  allocated_storage            = 20
  max_allocated_storage        = 100
  storage_encrypted            = true
  db_name                      = var.database_name
  username                     = var.database_username
  manage_master_user_password  = true
  publicly_accessible          = false
  multi_az                     = false
  backup_retention_period      = 1
  deletion_protection          = false
  skip_final_snapshot          = true
  db_subnet_group_name         = aws_db_subnet_group.platform.name
  vpc_security_group_ids       = [aws_security_group.database.id]
  performance_insights_enabled = false
}
