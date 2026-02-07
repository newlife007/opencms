# AWS Application Load Balancer Setup Guide

## Overview

This guide provides instructions for setting up AWS Application Load Balancer (ALB) for the OpenWan Media Asset Management System to achieve high availability and horizontal scalability.

## Architecture

```
Internet
    ↓
Application Load Balancer (Multi-AZ)
    ↓
Target Group (Health Checks)
    ↓
Backend API Instances (Auto-Scaling Group)
    ├─ Instance 1 (AZ-a)
    ├─ Instance 2 (AZ-b)
    └─ Instance 3 (AZ-c)
```

## Prerequisites

- AWS Account with appropriate permissions
- VPC with public subnets in at least 2 availability zones
- ACM certificate for HTTPS (or use ALB-generated certificate)
- Backend API instances running in private subnets

## Deployment Options

### Option 1: CloudFormation Template

Use the provided CloudFormation template:

```bash
aws cloudformation create-stack \
  --stack-name openwan-alb-prod \
  --template-body file://configs/alb-config.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=VpcId,ParameterValue=vpc-xxxxx \
    ParameterKey=SubnetIds,ParameterValue=\"subnet-xxx,subnet-yyy\" \
    ParameterKey=CertificateArn,ParameterValue=arn:aws:acm:region:account:certificate/xxx \
  --capabilities CAPABILITY_IAM
```

### Option 2: AWS Console

1. Navigate to EC2 → Load Balancers
2. Click "Create Load Balancer"
3. Select "Application Load Balancer"
4. Configure as described below

### Option 3: Terraform

See `terraform/alb.tf` for infrastructure as code

## Configuration Details

### 1. Load Balancer Settings

- **Name**: `production-openwan-alb`
- **Scheme**: Internet-facing
- **IP address type**: IPv4
- **Network mapping**: 
  - Select VPC
  - Select at least 2 subnets in different availability zones
  - Use public subnets for internet-facing ALB

### 2. Security Groups

Create security group for ALB:

```bash
# Inbound Rules
- HTTP (80) from 0.0.0.0/0
- HTTPS (443) from 0.0.0.0/0

# Outbound Rules
- All traffic to backend security group on port 8080
```

### 3. Listeners

#### HTTPS Listener (Port 443)

- **Protocol**: HTTPS
- **Port**: 443
- **Default action**: Forward to API target group
- **SSL Certificate**: Select ACM certificate
- **Security Policy**: ELBSecurityPolicy-TLS-1-2-2017-01 (or newer)

#### HTTP Listener (Port 80)

- **Protocol**: HTTP
- **Port**: 80
- **Default action**: Redirect to HTTPS (301)

### 4. Target Group Configuration

#### API Target Group

**Basic Configuration:**
- **Name**: `production-openwan-api-tg`
- **Target type**: IP targets (for ECS/Fargate) or Instances (for EC2)
- **Protocol**: HTTP
- **Port**: 8080
- **VPC**: Select your VPC

**Health Check Settings:**
```yaml
Path: /health
Protocol: HTTP
Port: traffic-port
Interval: 30 seconds
Timeout: 5 seconds
Healthy threshold: 2 consecutive successes
Unhealthy threshold: 2 consecutive failures
Success codes: 200
```

**Advanced Settings:**
```yaml
Deregistration delay: 300 seconds  # Connection draining
Slow start duration: 0 seconds     # No slow start needed for stateless services
Load balancing algorithm: Round robin
Stickiness: Disabled               # Stateless design
```

### 5. Routing Rules

Configure listener rules for routing:

**Rule 1: API Routes**
- **Priority**: 10
- **Condition**: Path pattern `/api/*`, `/health`, `/ready`, `/alive`
- **Action**: Forward to API target group

**Rule 2: Frontend Routes**
- **Priority**: 20
- **Condition**: Path pattern `/*`
- **Action**: Forward to frontend target group (if separate) or serve from S3/CloudFront

## Health Check Endpoints

The backend must implement these endpoints:

### /health - Comprehensive Health Check

Returns 200 OK when all dependencies are healthy:

```json
{
  "status": "healthy",
  "service": "openwan-api",
  "version": "1.0.0",
  "timestamp": "2024-02-01T12:00:00Z",
  "uptime": "3600 seconds",
  "checks": {
    "database": {"status": "healthy", "open_connections": 5},
    "redis": {"status": "healthy"},
    "queue": {"status": "healthy"},
    "storage": {"status": "healthy"},
    "ffmpeg": {"status": "healthy"}
  }
}
```

Returns 503 Service Unavailable when unhealthy:

```json
{
  "status": "unhealthy",
  "checks": {
    "database": {"status": "unhealthy", "error": "connection refused"}
  }
}
```

### /ready - Readiness Probe

Returns 200 only when service is ready to accept traffic:
- All connections established
- Database accessible
- Redis accessible

### /alive - Liveness Probe

Lightweight check that returns 200 if process is running.

## Auto-Scaling Integration

ALB integrates with Auto-Scaling groups:

1. **Register Targets**: New instances automatically register with target group
2. **Health Checks**: Unhealthy instances are deregistered automatically
3. **Connection Draining**: Connections drain for 300s before instance termination
4. **Zero-Downtime Deployments**: Rolling updates with health checks

## Monitoring and Alarms

### CloudWatch Metrics

Monitor these ALB metrics:

```yaml
- TargetResponseTime: Average response time from targets
  Alarm: > 1000ms for 2 consecutive periods
  
- UnHealthyHostCount: Number of unhealthy targets
  Alarm: > 0 for 2 consecutive periods
  
- HTTPCode_Target_5XX_Count: 5xx errors from targets
  Alarm: > 10 per minute
  
- HTTPCode_ELB_5XX_Count: 5xx errors from ALB
  Alarm: > 5 per minute
  
- RequestCount: Total requests
  Monitor: Track baseline and spikes
  
- RejectedConnectionCount: Rejected connections
  Alarm: > 0
```

### Create CloudWatch Alarms

```bash
# Unhealthy target alarm
aws cloudwatch put-metric-alarm \
  --alarm-name openwan-alb-unhealthy-targets \
  --alarm-description "Alert when targets are unhealthy" \
  --metric-name UnHealthyHostCount \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 60 \
  --evaluation-periods 2 \
  --threshold 1 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=TargetGroup,Value=targetgroup/production-openwan-api-tg/xxx \
  --alarm-actions arn:aws:sns:region:account:openwan-alerts
```

## Access Logs

Enable ALB access logs for debugging and analytics:

1. Create S3 bucket: `production-openwan-alb-logs`
2. Configure bucket policy to allow ALB to write logs
3. Enable access logs on ALB with bucket name

```bash
aws elbv2 modify-load-balancer-attributes \
  --load-balancer-arn arn:aws:elasticloadbalancing:... \
  --attributes \
    Key=access_logs.s3.enabled,Value=true \
    Key=access_logs.s3.bucket,Value=production-openwan-alb-logs \
    Key=access_logs.s3.prefix,Value=alb-logs
```

## Connection Draining

Configure deregistration delay for graceful shutdown:

- **Timeout**: 300 seconds
- **Behavior**: ALB stops sending new requests to deregistering target
- **In-flight requests**: Complete within timeout period
- **Backend**: Implement graceful shutdown to honor deregistration

## Sticky Sessions (Optional)

For stateful applications (not recommended for OpenWan stateless design):

```yaml
Stickiness: Enabled
Stickiness type: Load balancer generated cookie
Cookie name: AWSALB
Duration: 86400 seconds (24 hours)
```

For OpenWan, keep stickiness **disabled** as all state is in Redis.

## Cross-Zone Load Balancing

ALB has cross-zone load balancing enabled by default:
- Distributes traffic evenly across all targets in all AZs
- No additional configuration needed
- No extra charges

## SSL/TLS Configuration

### Use ACM Certificate

1. Request certificate in ACM for your domain
2. Validate domain ownership (DNS or email)
3. Attach certificate to HTTPS listener

### Security Policy

Use modern TLS policy:
- **Recommended**: `ELBSecurityPolicy-TLS13-1-2-2021-06`
- **Minimum**: `ELBSecurityPolicy-TLS-1-2-2017-01`
- Supports TLS 1.2 and 1.3
- Strong cipher suites

## Testing

### Test Health Checks

```bash
# From any instance, test health endpoint
curl http://backend-instance:8080/health

# Expected: 200 OK with JSON response
```

### Test ALB

```bash
# Test HTTPS
curl -I https://your-domain.com/health

# Test API endpoint
curl https://your-domain.com/api/v1/files

# Test redirect from HTTP to HTTPS
curl -I http://your-domain.com
# Expected: 301 redirect to HTTPS
```

### Verify Target Registration

```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn arn:aws:elasticloadbalancing:...

# Expected output:
# TargetHealth: healthy
# Reason: N/A
```

## Troubleshooting

### Targets Show Unhealthy

1. **Check health check path**: Verify `/health` returns 200
2. **Check security groups**: Ensure ALB can reach targets on port 8080
3. **Check target health**: SSH to instance and test `curl localhost:8080/health`
4. **Check logs**: Review target application logs for errors

### High Response Time

1. **Check target metrics**: Are targets CPU/memory constrained?
2. **Scale out**: Add more targets to handle load
3. **Check database**: Database might be bottleneck
4. **Enable caching**: Use Redis caching to reduce database load

### 5xx Errors

1. **Check target logs**: Review application error logs
2. **Check database**: Ensure database is accessible
3. **Check dependencies**: Verify Redis, RabbitMQ, S3 are healthy
4. **Increase timeout**: If requests timeout, increase health check timeout

## Best Practices

1. **Multi-AZ**: Deploy ALB across at least 2 availability zones
2. **Health Checks**: Use meaningful health checks that verify dependencies
3. **Monitoring**: Set up CloudWatch alarms for all critical metrics
4. **Access Logs**: Enable for troubleshooting and analytics
5. **HTTPS Only**: Redirect HTTP to HTTPS for security
6. **Security Groups**: Use least privilege principle
7. **Deletion Protection**: Enable in production to prevent accidental deletion
8. **Connection Draining**: Set appropriate timeout for graceful shutdown
9. **Target Group**: Use stateless design, disable sticky sessions
10. **Regular Testing**: Test failover scenarios regularly

## Cost Optimization

- Use ALB (not CLB) for better pricing and features
- Enable access logs only in production (storage costs)
- Review metrics and right-size target instances
- Use target utilization for auto-scaling decisions

## Security Considerations

1. **WAF Integration**: Consider AWS WAF for protection against web exploits
2. **Security Groups**: Restrict ALB to only necessary ports
3. **TLS Policy**: Use latest TLS policy for encryption
4. **Private Targets**: Place backend instances in private subnets
5. **Logging**: Enable access logs for security auditing

## References

- [ALB Documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/)
- [Health Check Configuration](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
- [Connection Draining](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-target-groups.html#deregistration-delay)
