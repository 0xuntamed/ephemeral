# Security model

## Trust boundaries

- Treat repository names, branches, image URIs, owners, and webhook data as untrusted input.
- The controller validates lifecycle identifiers before using them in filesystem paths, Terraform workspaces, database identifiers, or hostnames.
- Terraform is executed directly with an argument array; no user field is interpolated into a shell command.
- PostgreSQL schema names are derived by the controller, limited to 63 characters, checked against a strict identifier pattern, and quoted.

## Authentication and authorization

`EPHEMERAL_API_KEY` enables bearer authentication for API and metrics endpoints. With `PROVIDER=terraform` the API refuses to start without it, so a missing secret fails closed instead of exposing endpoints that create and destroy infrastructure; the standalone worker serves no HTTP and does not need it. It is suitable for the demo but not a complete multi-tenant authorization system. Production should terminate OIDC at an identity-aware proxy or add JWT validation and repository/owner authorization in the API.

Preview endpoints are intentionally unauthenticated in the local simulator. Put production previews behind your access proxy or ALB OIDC when the application is not public.

## AWS IAM

Use separate roles:

- Controller role: only the ECS, ELB, CloudWatch, Secrets Manager, Route53, S3-state, and lock operations represented by the Terraform modules.
- Task execution role: image pull, log delivery, and the one shared database secret.
- Preview workload role: application-specific data-plane permissions only; never infrastructure provisioning permissions.
- Jenkins/GitHub role: image publication and the ability to call the control plane, preferably via OIDC rather than stored AWS access keys.

## Secrets and preview data

- Never copy raw production data into previews. Use generated fixtures or an anonymization pipeline and a sanitized snapshot/template.
- Do not put database passwords in Terraform variables, container definitions, Jenkins artifacts, or API responses. The reference module injects the managed RDS password from Secrets Manager.
- Set `EPHEMERAL_API_KEY` and Jenkins credentials through their secret stores, not `.env` committed to Git.
- Terraform state can contain sensitive resource metadata; keep the S3 bucket encrypted, blocked from public access, versioned, and restricted to the controller role.

## Network controls

The reference workload security group accepts port 8080 only from the ALB security group. RDS accepts 5432 only from the workload group and is not publicly accessible. TLS terminates at the shared HTTPS listener using a wildcard ACM certificate.
