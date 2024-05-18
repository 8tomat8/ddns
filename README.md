# DDNS

A simple DDNS service for my local server, based on CloudFront API

## Configuration

### Local

```bash
DOMAIN=*.with-mask.domain.net,sub.domain.com
CLOUDFLARE_API_TOKEN=<CloudFlare API token>
```

### Docker compose

```yaml
version: "3.9"

name: gateway
services:
  ddns:
    container_name: ddns
    build:
      context: ddns
    environment:
      - DOMAIN=${DOMAIN:-sub.example.com}
      - CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN:-"a0we9fjaemRealSecretn0ae4jgf0aegicsn0"}

    restart: unless-stopped
```
