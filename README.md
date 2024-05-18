# DDNS

A simple DDNS service for my local server, based on CloudFront API

## Configuration

> **!!! IMPORTANT !!!**
>
> You must create the initial records in the CloudFront manually

### .env file for local or Docker setup

```bash
DOMAIN=*.with-mask.domain.net,sub.domain.com
CLOUDFLARE_API_TOKEN=<CloudFlare API token>
```

### Docker compose

```yaml
version: "3.9"

services:
  ddns:
    container_name: ddns
    build:
      context: .
    env_file:
      - .env

    restart: unless-stopped
```

### Include into your exising docker setup

From `v2.20.3` you can use `include` to easily plug the DDNS into your running setup. [docs](https://docs.docker.com/compose/multiple-compose-files/include/)

```yaml
version: "3.9"

include:
  - path: <path to this folder>/docker-compose.yml

services:
```
