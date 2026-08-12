# Failure modes

| Failure | Detection | Recovery |
|---|---|---|
| Duplicate webhook | Unique idempotency key | Return existing environment |
| CI dies after create | Durable desired state | Worker continues independently |
| Terraform apply fails | Provider error | Mark `FAILED`, preserve diagnostics, allow retry/cleanup |
| Partial AWS resources | Terraform workspace state | Destroy from the same workspace |
| Preview never becomes healthy | Health polling hits operation deadline | Mark `FAILED`; operator can inspect and clean up |
| PR-close webhook is missed | `expires_at` scan | TTL reaper schedules destroy |
| Worker dies mid-operation | Stale `updated_at` lease | Mark `ORPHANED`, reconcile toward destroyed |
| Terraform destroy fails | Provider error | `DESTROY_FAILED` with bounded retry backoff |
| Database schema cleanup fails | SQL error | Treat destroy as failed and retry the complete cleanup |
| Two workers see the same job | Atomic row lock and status check | Only one status claim succeeds |
| ALB priority hash collision | Terraform rejects the conflicting rule | Surface failure; production deployments should use a transactional priority allocator |
| State lock contention | S3 native lockfile | Terraform waits/fails without corrupting state |
| Registry unavailable | `/readyz` fails | Load balancer removes the API instance; state remains durable |

## Deliberate V1 trade-offs

- The worker scans the registry instead of using SQS/Redis. PostgreSQL row locking provides correctness at this scale; a durable queue is a straightforward V2 optimization.
- The reference VPC avoids NAT gateway cost by using public task IPs with security-group ingress restricted to the ALB. Production can use private subnets and VPC endpoints.
- Listener priorities are deterministic hashes. The collision is safe—it fails rather than routing to the wrong target—but a production control plane should allocate priorities transactionally.
- The local file store is single-process only. The standalone worker refuses to start without PostgreSQL.
