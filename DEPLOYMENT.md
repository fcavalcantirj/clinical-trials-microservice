# Deployment Guide

This guide covers deploying the Clinical Trials Microservice using Docker, Render, and Easypanel.

## Prerequisites

- Docker installed (for local testing)
- Render account (for cloud deployment) - [render.com](https://render.com) (free tier available)
- Easypanel account (for cloud deployment) - [easypanel.io](https://easypanel.io)
- Git repository (GitHub, GitLab, or Bitbucket) for auto-deployment

## Local Docker Deployment

### Build and Run with Docker

```bash
# Build the image
docker build -t clinical-trials-service .

# Run the container
docker run -p 8080:8080 clinical-trials-service
```

### Using Docker Compose

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Easypanel Deployment

Easypanel is a modern hosting platform that makes deploying Docker containers simple.

### Method 1: Deploy from Git Repository (Recommended)

1. **Push your code to a Git repository** (GitHub, GitLab, etc.)

2. **Create a new project in Easypanel**
   - Click "New Project"
   - Select your Git repository
   - Choose "Dockerfile" as the build method

3. **Configure the service:**
   - **Port**: `8080` (the service will automatically use the PORT environment variable)
   - **Health Check Path**: `/health`
   - **Health Check Port**: `8080`

4. **Environment Variables** (Optional):
   ```
   PORT=8080
   CACHE_ENABLED=true
   CACHE_TTL=6h
   ```

5. **Deploy!**

### Method 2: Deploy from Docker Image

If you've pushed the image to a container registry (Docker Hub, GHCR, etc.):

1. **Build and push the image:**
   ```bash
   # Tag the image
   docker build -t yourusername/clinical-trials-service:latest .

   # Push to registry
   docker push yourusername/clinical-trials-service:latest
   ```

2. **In Easypanel:**
   - Create a new project
   - Select "Docker Image"
   - Enter your image name: `yourusername/clinical-trials-service:latest`
   - Configure port: `8080`
   - Add environment variables if needed

### Easypanel Configuration

**Service Settings:**
- **Container Port**: `8080`
- **Health Check**: Enabled
  - **Path**: `/health`
  - **Port**: `8080`
  - **Interval**: `30s`

**Environment Variables:**
```bash
PORT=8080                    # Port to listen on (default: 8080)
CACHE_ENABLED=true           # Enable caching (default: true)
CACHE_TTL=6h                 # Cache TTL (default: 6h)
```

**Resource Limits** (Recommended):
- **CPU**: 0.5 - 1 CPU
- **Memory**: 256MB - 512MB
- **Disk**: 1GB (minimal, service is stateless)

### Post-Deployment

After deployment, test your service:

```bash
# Health check
curl https://your-domain.com/health

# Search trials
curl "https://your-domain.com/api/v1/trials/search?page_size=5"
```

## Render Deployment

Render is a modern cloud platform that makes deploying Docker containers simple with a generous free tier. The service can spin down after inactivity on the free tier but wakes up automatically on requests.

### Method 1: Deploy using render.yaml (Recommended)

The project includes a `render.yaml` configuration file that makes deployment automatic.

1. **Push your code to a Git repository** (GitHub, GitLab, or Bitbucket)

2. **Sign up for Render** at [render.com](https://render.com) (free tier available)

3. **Create a new Web Service:**
   - Go to your Render Dashboard
   - Click "New +" → "Blueprint"
   - Connect your Git repository
   - Render will automatically detect `render.yaml` and configure the service
   - Click "Apply" to deploy

4. **That's it!** Render will:
   - Build your Docker image from the Dockerfile
   - Deploy the service
   - Provide a public URL (e.g., `https://clinical-trials-microservice.onrender.com`)

### Method 2: Deploy via Render Dashboard (Manual Configuration)

If you prefer to configure manually or want to customize settings:

1. **Push your code to a Git repository**

2. **Create a new Web Service in Render:**
   - Go to Render Dashboard
   - Click "New +" → "Web Service"
   - Connect your Git repository
   - Select the repository and branch

3. **Configure the service:**
   - **Name**: `clinical-trials-microservice` (or your preferred name)
   - **Environment**: `Docker`
   - **Dockerfile Path**: `./Dockerfile`
   - **Docker Context**: `.` (root directory)
   - **Build Command**: (leave empty - Dockerfile handles this)
   - **Start Command**: (leave empty - uses CMD from Dockerfile)

4. **Environment Variables:**
   Add the following environment variables:
   ```
   PORT=8080
   CACHE_ENABLED=true
   CACHE_TTL=6h
   ```

5. **Health Check:**
   - **Health Check Path**: `/health`
   - Render will automatically check this endpoint

6. **Plan:**
   - **Free**: 512MB RAM, spins down after 15 minutes of inactivity (free)
   - **Starter**: $7/month - Always on, 512MB RAM
   - **Standard**: $25/month - Always on, 1GB RAM (recommended for production)

7. **Click "Create Web Service"** to deploy

### Render Configuration Details

**Service Settings:**
- **Container Port**: `8080` (automatically detected)
- **Health Check Path**: `/health`
- **Auto-Deploy**: Enabled by default (deploys on every push to main branch)
- **Region**: Oregon (US) or select closest to your users

**Environment Variables:**
```bash
PORT=8080                    # Port to listen on (required by Render)
CACHE_ENABLED=true           # Enable caching (default: true)
CACHE_TTL=6h                 # Cache TTL duration (default: 6h)
```

**Free Tier Limitations:**
- Service spins down after 15 minutes of inactivity
- First request after spin-down may take 30-60 seconds (cold start)
- 512MB RAM limit
- Free SSL certificate included
- Public URL provided (e.g., `*.onrender.com`)

**Upgrading from Free Tier:**
- **Starter Plan ($7/month)**: Service stays awake, better for production
- **Standard Plan ($25/month)**: More RAM, better performance, recommended for high traffic

### Post-Deployment (Render)

After deployment, Render will provide a public URL. Test your service:

```bash
# Health check
curl https://your-service-name.onrender.com/health

# Search trials
curl "https://your-service-name.onrender.com/api/v1/trials/search?page_size=5"
```

### Render-Specific Features

**Auto-Deploy:**
- Automatically deploys on every push to your main/master branch
- Can be configured in service settings
- Manual deploys available from dashboard

**Logs:**
- View real-time logs in the Render dashboard
- Logs are retained for 7 days (free tier) or 30 days (paid plans)
- Can download logs as needed

**Metrics:**
- CPU and memory usage available in dashboard
- Request metrics and response times
- Monitor service health and performance

### Render Troubleshooting

**Service won't start:**
- Check logs in Render dashboard (under "Logs" tab)
- Verify Dockerfile builds successfully locally first
- Ensure PORT environment variable is set to 8080

**Health check failing:**
- Verify service is responding: `curl https://your-service.onrender.com/health`
- Check that health check path is set to `/health`
- Review logs for startup errors

**Cold start delays (Free tier):**
- First request after 15 minutes of inactivity may be slow
- This is normal for free tier - consider upgrading to Starter plan for always-on service
- Subsequent requests are fast

**Build failures:**
- Check that Dockerfile is in the repository root
- Verify go.mod and dependencies are correct
- Review build logs in Render dashboard

**Out of memory:**
- Free tier has 512MB limit
- Consider upgrading to Starter ($7/month) or Standard ($25/month) for more resources
- Reduce cache TTL to lower memory usage

**Service spinning down too often:**
- Free tier spins down after 15 minutes of inactivity
- Upgrade to Starter plan ($7/month) for always-on service
- Consider using an uptime monitor to ping your service periodically (though this may violate free tier terms)

## Production Considerations

### Environment Variables

The service supports the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port to listen on |
| `CACHE_ENABLED` | `true` | Enable/disable caching (via command flag) |
| `CACHE_TTL` | `6h` | Cache TTL duration (via command flag) |

Note: `CACHE_ENABLED` and `CACHE_TTL` are currently only configurable via command flags. For Docker, you can override the CMD:

```dockerfile
CMD ["./trials-service", "-cache=true", "-cache-ttl=6h"]
```

### Health Checks

The service includes a health check endpoint at `/health` that returns:
```json
{
  "status": "healthy"
}
```

Both Render and Easypanel will automatically use this for container health monitoring.

### Scaling

The service is stateless and can be horizontally scaled:

- **Multiple instances**: Run multiple containers behind a load balancer
- **Cache**: Each instance has its own in-memory cache
- **Rate limiting**: Rate limiting is per-instance (2-second delay)

### Monitoring

Consider adding:
- Application logs (stdout/stderr are automatically captured)
- Metrics endpoint (can be added)
- Uptime monitoring (external service)

### Security

- The service runs as a non-root user (UID 1000)
- Only necessary ports are exposed
- CORS is enabled (configure for your domain in production if needed)

## Troubleshooting

### Container won't start

Check logs:
```bash
docker logs clinical-trials-service
# or in Easypanel: View logs in the dashboard
```

### Health check failing

Verify the service is responding:
```bash
curl http://localhost:8080/health
```

### Port conflicts

If port 8080 is in use, change it:
```bash
# Docker
docker run -p 3000:8080 -e PORT=8080 clinical-trials-service

# Easypanel: Update container port in settings
```

### Out of memory

Increase memory limits in Easypanel or reduce cache TTL.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: docker build -t yourusername/clinical-trials-service:${{ github.sha }} .
      
      - name: Push to registry
        run: |
          docker push yourusername/clinical-trials-service:${{ github.sha }}
          docker push yourusername/clinical-trials-service:latest
```

Then configure Easypanel to pull from your registry automatically.

## Cost Optimization

- **Memory**: Start with 256MB, increase if needed
- **CPU**: 0.5 CPU is usually sufficient for moderate traffic
- **Caching**: Keep caching enabled to reduce external API calls
- **Scaling**: Use horizontal scaling rather than over-provisioning

## Support

For deployment issues:
1. Check container logs in Render/Easypanel dashboard
2. Verify health check endpoint
3. Test locally with Docker first
4. Review the main README.md for API usage

**Platform-Specific Support:**
- **Render**: Check logs in Render dashboard, review [Render Docs](https://render.com/docs)
- **Easypanel**: Check logs in Easypanel dashboard, review [Easypanel Docs](https://easypanel.io/docs)
