# Cloud Storage Support in Containerlab

Containerlab supports using cloud storage URLs to retrieve topology files and startup configurations for network devices. This includes support for:

- **AWS S3** (`s3://`)
- **Google Cloud Storage** (`gs://`)
- **Azure Blob Storage** (`azblob://`)

## Usage Examples

### Topology Files from Cloud Storage

Deploy a lab using a topology file stored in cloud storage:

```bash
# AWS S3
containerlab deploy -t s3://my-bucket/topologies/my-lab.clab.yml

# Google Cloud Storage
containerlab deploy -t gs://my-bucket/topologies/my-lab.clab.yml

# Azure Blob Storage
containerlab deploy -t azblob://my-container/topologies/my-lab.clab.yml
```

### Startup Configurations from Cloud Storage

In your topology file, you can reference startup configurations stored in cloud storage:

```yaml
name: my-lab
topology:
  nodes:
    router1:
      kind: srl
      image: ghcr.io/nokia/srlinux:latest
      # Can use S3, GCS, or Azure Blob URLs
      startup-config: s3://my-bucket/configs/router1.cli
    
    router2:
      kind: srl
      image: ghcr.io/nokia/srlinux:latest
      startup-config: gs://my-bucket/configs/router2.cli
    
    router3:
      kind: srl
      image: ghcr.io/nokia/srlinux:latest
      startup-config: azblob://my-container/configs/router3.cli
```

## AWS S3

### Prerequisites

- AWS credentials configured in your environment
- Access to the S3 bucket containing your files

### Authentication

The S3 integration uses the standard AWS SDK credential chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (when running on EC2/ECS/Lambda)
4. AWS SSO

### URL Format

```
s3://bucket-name/path/to/file
```

### Specifying AWS Region

You can optionally specify the AWS region directly in the URL using a query parameter:
```
s3://bucket-name/path/to/file?region=us-west-2
```

If no region is specified in the URL, the S3 integration will use:
1. The default region from your AWS config file (`~/.aws/config`)
2. The `AWS_DEFAULT_REGION` environment variable
3. The region from instance metadata (when running on EC2)

Examples:
- `s3://my-bucket/configs/router.cli` - uses default region
- `s3://my-bucket/configs/router.cli?region=eu-west-1` - uses eu-west-1 region
- `s3://my-bucket/configs/router.cli?region=auto` - auto-detect region

## Google Cloud Storage (GCS)

### Prerequisites

- Google Cloud credentials configured in your environment
- Access to the GCS bucket containing your files

### Authentication

The GCS integration uses Application Default Credentials (ADC):

1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to a service account key file
2. gcloud auth application-default login
3. Service account attached to the GCE instance
4. Google Cloud Shell or Cloud Run service identity

### URL Format

```
gs://bucket-name/path/to/file
```

### Setting up Authentication

```bash
# Option 1: Using service account key
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"

# Option 2: Using gcloud CLI
gcloud auth application-default login
```

## Azure Blob Storage

### Prerequisites

- Azure credentials configured in your environment
- Access to the Azure storage container containing your files

### Authentication

The Azure Blob integration supports multiple authentication methods:

1. Storage account key via environment variables:
   - `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY`
2. Azure AD authentication:
   - `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, and `AZURE_CLIENT_SECRET`
3. Managed Identity (when running on Azure)
4. Azure CLI authentication

### URL Format

```
azblob://container-name/path/to/file
```

### Setting up Authentication

```bash
# Option 1: Using storage account key
export AZURE_STORAGE_ACCOUNT="mystorageaccount"
export AZURE_STORAGE_KEY="myaccesskey"

# Option 2: Using Azure AD service principal
export AZURE_CLIENT_ID="client-id"
export AZURE_TENANT_ID="tenant-id"
export AZURE_CLIENT_SECRET="client-secret"

# Option 3: Using Azure CLI
az login
```

## Troubleshooting

### Authentication Issues

If you encounter authentication errors:

1. **AWS S3**: Run `aws s3 ls s3://your-bucket/` to verify credentials
2. **GCS**: Run `gsutil ls gs://your-bucket/` to verify credentials
3. **Azure**: Run `az storage container list --account-name your-account` to verify credentials

### Network Connectivity

Ensure your containerlab host has network access to the cloud storage endpoints:
- S3: `*.s3.amazonaws.com` or `*.s3.[region].amazonaws.com`
- GCS: `storage.googleapis.com`
- Azure: `*.blob.core.windows.net`