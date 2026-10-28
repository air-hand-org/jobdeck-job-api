# jobdeck-job-api

## Tables

| Name | Columns | Comment | Type |
| ---- | ------- | ------- | ---- |
| [job.actor](job.actor.md) | 4 |  | VIEW |
| [job.actor_projection](job.actor_projection.md) | 5 |  | BASE TABLE |
| [job.actor_role](job.actor_role.md) | 1 |  | BASE TABLE |
| [job.actor_role_assignment](job.actor_role_assignment.md) | 2 |  | BASE TABLE |
| [job.execution](job.execution.md) | 12 |  | BASE TABLE |
| [job.execution_claim_detail](job.execution_claim_detail.md) | 9 |  | VIEW |
| [job.execution_detail](job.execution_detail.md) | 19 |  | VIEW |
| [job.execution_event](job.execution_event.md) | 4 |  | BASE TABLE |
| [job.execution_event_status](job.execution_event_status.md) | 2 |  | BASE TABLE |
| [job.execution_trigger](job.execution_trigger.md) | 1 |  | BASE TABLE |
| [job.job](job.job.md) | 6 |  | BASE TABLE |
| [job.job_current_revision](job.job_current_revision.md) | 2 |  | BASE TABLE |
| [job.job_detail](job.job_detail.md) | 12 |  | VIEW |
| [job.job_retry_backoff_type](job.job_retry_backoff_type.md) | 1 |  | BASE TABLE |
| [job.job_revision](job.job_revision.md) | 10 |  | BASE TABLE |
| [job.job_type](job.job_type.md) | 2 |  | BASE TABLE |
| [job.outbox_execution_completed](job.outbox_execution_completed.md) | 4 |  | BASE TABLE |
| [job.outbox_execution_requested](job.outbox_execution_requested.md) | 4 |  | BASE TABLE |
| [job.outbox_execution_status](job.outbox_execution_status.md) | 1 |  | BASE TABLE |

## Stored procedures and functions

| Name | ReturnType | Arguments | Type |
| ---- | ------- | ------- | ---- |
| job.update_updated_at_column | trigger |  | FUNCTION |

## Relations

```mermaid
erDiagram

"job.actor_role_assignment" }o--|| "job.actor_role" : "FOREIGN KEY (role) REFERENCES job.actor_role(role) ON UPDATE CASCADE"
"job.actor_role_assignment" |o--|| "job.actor_projection" : "FOREIGN KEY (actor_id) REFERENCES job.actor_projection(actor_id)"
"job.execution" }o--|| "job.job" : "FOREIGN KEY (job_id) REFERENCES job.job(job_id) ON DELETE CASCADE"
"job.execution" }o--|| "job.execution_event_status" : "FOREIGN KEY (status) REFERENCES job.execution_event_status(status) ON UPDATE CASCADE"
"job.execution" }o--|| "job.execution_trigger" : "FOREIGN KEY (trigger_type) REFERENCES job.execution_trigger(trigger_type) ON UPDATE CASCADE"
"job.execution" }o--|| "job.job_revision" : "FOREIGN KEY (job_id, job_revision_id) REFERENCES job.job_revision(job_id, revision_id)"
"job.execution_event" }o--|| "job.execution_event_status" : "FOREIGN KEY (status) REFERENCES job.execution_event_status(status) ON UPDATE CASCADE"
"job.execution_event" }o--|| "job.execution" : "FOREIGN KEY (execution_id) REFERENCES job.execution(execution_id) ON DELETE CASCADE"
"job.job_current_revision" |o--|| "job.job" : "FOREIGN KEY (job_id) REFERENCES job.job(job_id) ON DELETE CASCADE"
"job.job_current_revision" }o--|| "job.job_revision" : "FOREIGN KEY (job_id, revision_id) REFERENCES job.job_revision(job_id, revision_id)"
"job.job_revision" }o--|| "job.job_type" : "FOREIGN KEY (type) REFERENCES job.job_type(type) ON UPDATE CASCADE"
"job.job_revision" }o--|| "job.job_retry_backoff_type" : "FOREIGN KEY (retry_backoff_type) REFERENCES job.job_retry_backoff_type(type) ON UPDATE CASCADE"
"job.job_revision" }o--|| "job.job" : "FOREIGN KEY (job_id) REFERENCES job.job(job_id) ON DELETE CASCADE"
"job.outbox_execution_requested" }o--|| "job.outbox_execution_status" : "FOREIGN KEY (status) REFERENCES job.outbox_execution_status(status)"

"job.actor" {
  boolean active
  uuid actor_id
  varchar_20_ role
  varchar_30_ type
}
"job.actor_projection" {
  boolean active
  uuid actor_id
  timestamp_with_time_zone created_at
  varchar_30_ type
  timestamp_with_time_zone updated_at
}
"job.actor_role" {
  varchar_20_ role
}
"job.actor_role_assignment" {
  uuid actor_id FK
  varchar_20_ role FK
}
"job.execution" {
  integer attempt
  timestamp_with_time_zone created_at
  uuid execution_id
  uuid job_id FK
  uuid job_revision_id FK
  timestamp_with_time_zone lease_expires_at
  uuid lease_token
  timestamp_with_time_zone next_run_at
  jsonb result
  varchar_30_ status FK
  varchar_30_ trigger_type FK
  timestamp_with_time_zone updated_at
}
"job.execution_claim_detail" {
  integer attempt
  timestamp_with_time_zone created_at
  uuid execution_id
  jsonb job_config
  uuid job_id
  uuid job_revision_id
  varchar_30_ job_type
  timestamp_with_time_zone lease_expires_at
  uuid lease_token
}
"job.execution_detail" {
  integer attempt
  timestamp_with_time_zone created_at
  varchar_30_ current_status
  uuid execution_id
  jsonb job_config
  uuid job_id
  uuid job_owner_id
  varchar_50_ job_retry_backoff_type
  bigint job_retry_initial_delay_ms
  integer job_retry_max_attempts
  bigint job_retry_max_delay_ms
  uuid job_revision_id
  varchar_30_ job_type
  timestamp_with_time_zone lease_expires_at
  uuid lease_token
  timestamp_with_time_zone next_run_at
  jsonb result
  varchar_30_ trigger_type
  timestamp_with_time_zone updated_at
}
"job.execution_event" {
  timestamp_with_time_zone created_at
  uuid execution_event_id
  uuid execution_id FK
  varchar_30_ status FK
}
"job.execution_event_status" {
  boolean active
  varchar_30_ status
}
"job.execution_trigger" {
  varchar_30_ trigger_type
}
"job.job" {
  timestamp_with_time_zone created_at
  boolean enabled
  uuid job_id
  varchar_255_ name
  uuid owner_id
  timestamp_with_time_zone updated_at
}
"job.job_current_revision" {
  uuid job_id FK
  uuid revision_id FK
}
"job.job_detail" {
  jsonb config
  timestamp_with_time_zone created_at
  boolean enabled
  uuid job_id
  varchar_255_ name
  uuid owner_id
  varchar_50_ retry_backoff_type
  bigint retry_initial_delay_ms
  integer retry_max_attempts
  bigint retry_max_delay_ms
  varchar_30_ type
  timestamp_with_time_zone updated_at
}
"job.job_retry_backoff_type" {
  varchar_50_ type
}
"job.job_revision" {
  jsonb config
  timestamp_with_time_zone created_at
  uuid job_id FK
  varchar_50_ retry_backoff_type FK
  bigint retry_initial_delay_ms
  integer retry_max_attempts
  bigint retry_max_delay_ms
  uuid revision_id
  integer revision_number
  varchar_30_ type FK
}
"job.job_type" {
  boolean active
  varchar_30_ type
}
"job.outbox_execution_completed" {
  timestamp_with_time_zone completed_at
  uuid execution_id
  timestamp_with_time_zone published_at
  timestamp_with_time_zone requested_at
}
"job.outbox_execution_requested" {
  timestamp_with_time_zone created_at
  uuid execution_id
  timestamp_with_time_zone published_at
  varchar_20_ status FK
}
"job.outbox_execution_status" {
  varchar_20_ status
}
```

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
