# S3 Support in Containerlab

Containerlab now supports using S3 URLs to retrieve topology files and startup configurations for network devices.

## Prerequisites

- AWS credentials configured in your environment (via AWS CLI, environment variables, or IAM roles)
- Access to the S3 bucket containing your files

## Usage Examples

### Topology Files from S3

Deploy a lab using a topology file stored in S3:

```bash
containerlab deploy -t s3://my-bucket/topologies/my-lab.clab.yml
```

### Startup Configurations from S3

In your topology file, you can reference startup configurations stored in S3:

```yaml
name: my-lab
topology:
  nodes:
    router1:
      kind: srl
      image: ghcr.io/nokia/srlinux:latest
      startup-config: s3://my-bucket/configs/router1.cli
    
    router2:
      kind: srl
      image: ghcr.io/nokia/srlinux:latest
      startup-config: s3://my-bucket/configs/router2.cli
```

## AWS Authentication

The S3 integration uses the standard AWS SDK credential chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (when running on EC2/ECS/Lambda)
4. AWS SSO

## S3 URL Format

S3 URLs must follow this format:
```
s3://bucket-name/path/to/file
```

Both the bucket name and file path are required.