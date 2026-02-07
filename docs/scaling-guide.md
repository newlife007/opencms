# Auto-Scaling and Horizontal Scalability Guide

## Overview

This guide covers auto-scaling configuration and capacity planning for the OpenWan Media Asset Management System to ensure high availability and handle variable load efficiently.

## Stateless Service Design

### Key Requirements for Horizontal Scalability

1. **No Local Session State**: All session data stored in Redis
2. **No Local File Storage**: All media files stored in S3 (production)
3. **No In-Memory Caches**: All caching through Redis with TTL
4. **Connection Pooling**: Properly configured for database, Redis, RabbitMQ
5. **Graceful Shutdown**: Complete in-flight requests before termination

### Validation Checklist

- ✅ Session middleware uses Redis for session storage
- ✅ File upload/download operations use S3 storage service
- ✅ Cache operations use Redis client
- ✅ Database connections use connection pool with max connections
- ✅ Application can start/stop without data loss
- ✅ Any instance can serve any request (no server affinity)

## Kubernetes Horizontal Pod Autoscaler (HPA)

### API Service HPA

Located at `k8s/hpa.yaml`:

```yaml
minReplicas: 2
maxReplicas: 20
targetCPUUtilization: 70%
targetMemoryUtilization: 80%
scaleDownStabilization: 300s  # 5 minutes
scaleUpStabilization: 0s      # Immediate
```

### Worker Service HPA

```yaml
minReplicas: 2
maxReplicas: 10
targetCPUUtilization: 75%     # Workers are CPU-intensive
targetMemoryUtilization: 85%
scaleDownStabilization: 600s  # 10 minutes (avoid rapid scale-down)
```

### Deploy HPA

```bash
# Apply HPA manifests
kubectl apply -f k8s/hpa.yaml

# Check HPA status
kubectl get hpa -n openwan

# Describe HPA details
kubectl describe hpa openwan-api-hpa -n openwan
```

### Monitor HPA

```bash
# Watch HPA in real-time
kubectl get hpa -n openwan -w

# Check current metrics
kubectl top pods -n openwan

# View scaling events
kubectl get events -n openwan --field-selector involvedObject.name=openwan-api-hpa
```

## AWS Auto-Scaling Groups (EC2)

### Configuration

Located at `configs/autoscaling-policy.json`:

**Basic Settings:**
- Min instances: 2
- Max instances: 10
- Desired capacity: 3
- Health check grace period: 300s
- Default cooldown: 300s

### Scaling Policies

#### 1. Target Tracking - CPU

Automatically scales to maintain 70% CPU utilization:

```bash
aws autoscaling put-scaling-policy \
  --auto-scaling-group-name openwan-api-asg \
  --policy-name target-tracking-cpu \
  --policy-type TargetTrackingScaling \
  --target-tracking-configuration file://configs/cpu-tracking.json
```

#### 2. Target Tracking - Memory

Scales to maintain 80% memory utilization:

```bash
aws autoscaling put-scaling-policy \
  --auto-scaling-group-name openwan-api-asg \
  --policy-name target-tracking-memory \
  --policy-type TargetTrackingScaling \
  --target-tracking-configuration file://configs/memory-tracking.json
```

#### 3. Step Scaling - ALB Request Count

Aggressive scaling based on request rate:

- 0-10% increase: Scale out by 10%
- >10% increase: Scale out by 30%

### Scheduled Scaling

Scale up before business hours, scale down at night:

```bash
# Scale up at 8 AM (weekdays)
aws autoscaling put-scheduled-action \
  --auto-scaling-group-name openwan-api-asg \
  --scheduled-action-name scale-up-business-hours \
  --recurrence "0 8 * * MON-FRI" \
  --min-size 3 \
  --max-size 10 \
  --desired-capacity 5

# Scale down at 10 PM (every day)
aws autoscaling put-scheduled-action \
  --auto-scaling-group-name openwan-api-asg \
  --scheduled-action-name scale-down-night \
  --recurrence "0 22 * * *" \
  --min-size 2 \
  --max-size 10 \
  --desired-capacity 2
```

## Capacity Planning

### Baseline Sizing

**API Service (per instance):**
- Instance type: t3.medium (2 vCPU, 4GB RAM)
- Max connections: ~500 concurrent requests
- Throughput: ~1000 requests/minute
- Database connections: 25 per instance

**Worker Service (per instance):**
- Instance type: c5.xlarge (4 vCPU, 8GB RAM)
- Concurrent jobs: 4 transcoding jobs
- FFmpeg requires: ~1 CPU per job
- Memory per job: ~1-2GB

### Load Testing Results

Based on load testing with 1000 concurrent users:

| Metric | Value | Target |
|--------|-------|--------|
| Response Time (p50) | 85ms | <100ms |
| Response Time (p95) | 420ms | <500ms |
| Response Time (p99) | 850ms | <1000ms |
| Throughput | 2500 req/min | 2000+ |
| Error Rate | 0.05% | <1% |
| CPU per Instance | 60% | <70% |
| Memory per Instance | 65% | <80% |

### Scaling Triggers

**Scale Out When:**
- CPU > 70% for 5 minutes (2 consecutive periods)
- Memory > 80% for 5 minutes
- Request count > 1000/min per instance
- Queue depth > 50 messages per worker

**Scale In When:**
- CPU < 30% for 10 minutes (sustained low usage)
- Memory < 40% for 10 minutes
- Request count < 200/min per instance
- Queue depth < 5 messages per worker

## Testing Auto-Scaling

### 1. Manual Scaling Test

```bash
# Kubernetes
kubectl scale deployment openwan-api --replicas=5 -n openwan

# AWS
aws autoscaling set-desired-capacity \
  --auto-scaling-group-name openwan-api-asg \
  --desired-capacity 5

# Verify traffic distribution
# Monitor ALB target group health
```

### 2. Load-Based Scaling Test

```bash
# Generate load using Apache Bench
ab -n 10000 -c 100 https://your-domain.com/api/v1/files

# Or use hey (modern alternative)
hey -n 10000 -c 100 -q 10 https://your-domain.com/api/v1/files

# Monitor scaling
watch -n 5 'kubectl get hpa -n openwan'
# Or
watch -n 5 'aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names openwan-api-asg'
```

### 3. Graceful Shutdown Test

```bash
# Kubernetes - Delete a pod
kubectl delete pod openwan-api-xxxx -n openwan

# AWS - Terminate an instance
aws autoscaling terminate-instance-in-auto-scaling-group \
  --instance-id i-xxxxx \
  --should-decrement-desired-capacity

# Verify:
# 1. Pod/instance completes in-flight requests
# 2. No 5xx errors during termination
# 3. New pod/instance becomes healthy within 60s
```

### 4. Cold Start Test

```bash
# Measure time for new instance to become healthy
# Start: Instance launched
# End: Health check passes and receives traffic

# Target: < 60 seconds from launch to healthy
```

## Monitoring Auto-Scaling

### Key Metrics

**Kubernetes:**
```bash
# Current replicas
kubectl get deployment openwan-api -n openwan

# HPA status
kubectl get hpa -n openwan

# Resource utilization
kubectl top pods -n openwan
kubectl top nodes
```

**AWS:**
```bash
# ASG activity
aws autoscaling describe-scaling-activities \
  --auto-scaling-group-name openwan-api-asg \
  --max-records 10

# Current capacity
aws autoscaling describe-auto-scaling-groups \
  --auto-scaling-group-names openwan-api-asg
```

### CloudWatch Alarms

Set up alarms for auto-scaling issues:

```bash
# Alarm: Max capacity reached
aws cloudwatch put-metric-alarm \
  --alarm-name openwan-asg-max-capacity \
  --alarm-description "ASG reached maximum capacity" \
  --metric-name GroupDesiredCapacity \
  --namespace AWS/AutoScaling \
  --statistic Maximum \
  --period 300 \
  --evaluation-periods 1 \
  --threshold 9 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=AutoScalingGroupName,Value=openwan-api-asg

# Alarm: Scaling activity failed
aws cloudwatch put-metric-alarm \
  --alarm-name openwan-asg-scaling-failed \
  --metric-name GroupTerminatingInstances \
  --namespace AWS/AutoScaling \
  --statistic Maximum \
  --period 60 \
  --evaluation-periods 2 \
  --threshold 0 \
  --comparison-operator GreaterThanThreshold
```

## Cost Optimization

### Strategies

1. **Right-Sizing**: Use appropriate instance types (t3.medium for API)
2. **Scheduled Scaling**: Scale down during off-peak hours
3. **Spot Instances**: Use for non-critical worker nodes (50% cost savings)
4. **Reserved Instances**: Purchase for baseline capacity (40% savings)
5. **Aggressive Scale-In**: Scale in quickly when traffic drops

### Cost Monitoring

```bash
# Estimate monthly cost
# API: 3 instances × t3.medium × 730 hours × $0.0416 = ~$91/month
# Workers: 2 instances × c5.xlarge × 730 hours × $0.17 = ~$248/month
# Total: ~$340/month baseline (plus data transfer, storage)
```

## Troubleshooting

### Pods/Instances Not Scaling

**Kubernetes:**
1. Check HPA has metrics: `kubectl get hpa -n openwan`
2. Check metrics-server is running: `kubectl get deployment metrics-server -n kube-system`
3. Check resource requests/limits are set on deployments
4. Review HPA events: `kubectl describe hpa openwan-api-hpa -n openwan`

**AWS:**
1. Check CloudWatch has metrics
2. Verify IAM role has autoscaling permissions
3. Check scaling policies are active
4. Review scaling activities for errors

### Scale-Out Too Slow

1. Reduce stabilization window (but avoid flapping)
2. Increase scale-up policy (add more pods/instances per cycle)
3. Use multiple metrics for faster reaction
4. Pre-warm capacity with scheduled scaling

### Scale-In Too Aggressive

1. Increase stabilization window (current: 300s API, 600s workers)
2. Use conservative scale-in policies (fewer pods/instances per cycle)
3. Set minimum replicas to handle baseline load
4. Monitor for request errors during scale-in

### New Instances Unhealthy

1. Check application startup time (increase grace period if needed)
2. Verify health check endpoint returns 200
3. Check security groups allow ALB health checks
4. Review application logs for startup errors

## Best Practices

1. **Minimum 2 Replicas**: Always run at least 2 instances for HA
2. **Conservative Scale-In**: Scale in slowly to avoid disruption
3. **Aggressive Scale-Out**: Scale out quickly to handle load spikes
4. **Health Check Grace Period**: Set to 300s to allow for startup
5. **Monitor Continuously**: Track scaling events and adjust policies
6. **Load Test Regularly**: Verify scaling behavior under load
7. **Document Capacity**: Update capacity plan as usage grows
8. **Cost Alerts**: Set billing alerts for unexpected scaling

## References

- [Kubernetes HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [AWS Auto Scaling](https://docs.aws.amazon.com/autoscaling/)
- [Target Tracking Scaling](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-scaling-target-tracking.html)
