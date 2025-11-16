# Supabase Local Development

This directory contains the Docker Compose configuration for running Supabase locally.

## Quick Start

1. **Copy the environment file:**
   ```bash
   cd supabase
   cp .env.example .env
   ```

2. **Update the `.env` file** with your values (at minimum, change the passwords and secrets)
   
   **Note:** The `docker-compose.yml` uses `env_file: .env` to explicitly load environment variables from the `.env` file.

3. **Start Supabase:**
   ```bash
   docker-compose up -d
   ```

4. **Check logs:**
   ```bash
   docker-compose logs -f
   ```

5. **Stop Supabase:**
   ```bash
   docker-compose down
   ```

## Access Points

Once running, you can access:

- **Studio (Admin UI)**: http://localhost:3011
- **API Gateway**: http://localhost:3018
- **REST API**: http://localhost:3018/rest/v1/
- **Auth API**: http://localhost:3018/auth/v1/
- **Realtime**: ws://localhost:3018/realtime/v1/
- **Storage API**: http://localhost:3018/storage/v1/
- **Database**: localhost:54322 (default port)

## Environment Files

### Development (`.env.example`)

The `.env.example` file contains default values suitable for local development. These include:
- Demo API keys (safe for local use only)
- Default passwords (change these!)
- Localhost URLs

**⚠️ Never use these values in production!**

### Production (`.env.production.example`)

The `.env.production.example` file contains placeholders for production configuration. Before deploying:

1. Copy `.env.production.example` to `.env.production`
2. Generate secure, random values for all secrets
3. Update all URLs to your production domains
4. Configure S3 for storage (if using cloud storage)

**To use production config:**
```bash
docker-compose --env-file .env.production up -d
```

## Generating Secure Keys

### JWT Secret
```bash
openssl rand -base64 64
```

### Postgres Password
```bash
openssl rand -base64 32
```

### API Keys (ANON_KEY and SERVICE_ROLE_KEY)

Use Supabase CLI:
```bash
supabase gen keys
```

Or generate JWT tokens manually with appropriate claims.

## Services

The Docker Compose setup includes:

- **supabase-db**: PostgreSQL database
- **supabase-studio**: Admin UI
- **supabase-rest**: PostgREST API
- **supabase-auth**: GoTrue authentication
- **supabase-realtime**: Realtime subscriptions
- **supabase-storage**: File storage
- **supabase-imgproxy**: Image transformation
- **supabase-meta**: Postgres metadata API
- **supabase-kong**: API Gateway

## Volumes

Data is persisted in Docker volumes:
- `supabase-db-data`: Database data
- `supabase-storage-data`: Storage files

To remove all data:
```bash
docker-compose down -v
```

## Troubleshooting

### Port conflicts
If ports are already in use, update the port mappings in `.env`:
```env
DB_PORT=54322
STUDIO_PORT=3012
KONG_HTTP_PORT=3019
# etc.
```

### Database connection issues
Ensure the database is healthy before other services start:
```bash
docker-compose ps
docker-compose logs supabase-db
```

### Reset everything
```bash
docker-compose down -v
docker-compose up -d
```

