# Job Example

This example demonstrates Kubernetes Job patterns for batch processing using the wetwire pattern.

## Jobs

| Job | Type | Description |
|-----|------|-------------|
| `database-migration` | Simple | One-time execution with retry and timeout |
| `batch-processor` | Parallel | Process 10 items with 3 concurrent pods |
| `report-generator` | Indexed | Each pod gets unique index (0-4) |
| `data-import` | Init Container | Download data before processing |

## Patterns Demonstrated

### Simple Job
- `backoffLimit`: Retry on failure
- `activeDeadlineSeconds`: Time limit
- `ttlSecondsAfterFinished`: Auto-cleanup

### Parallel Job
- `completions`: Total successful runs needed
- `parallelism`: Max concurrent pods
- Work queue pattern

### Indexed Job
- `completionMode: Indexed`: Each pod gets `JOB_COMPLETION_INDEX`
- Useful for sharded processing
- Each index processes different data

### Init Container Pattern
- Download/prepare data before main container
- Shared volume between init and main containers

## Usage

```bash
# Build YAML manifests
wetwire-k8s build ./examples/job

# Run a specific job
wetwire-k8s build ./examples/job | kubectl apply -f -

# Watch job progress
kubectl get jobs -w

# View job logs
kubectl logs job/database-migration
```

## Best Practices

1. **Always set resource limits** - Prevent runaway jobs
2. **Use `backoffLimit`** - Control retry behavior
3. **Set `activeDeadlineSeconds`** - Prevent hung jobs
4. **Use `ttlSecondsAfterFinished`** - Auto-cleanup completed jobs
5. **Use `restartPolicy: Never`** - Let Job controller handle retries
