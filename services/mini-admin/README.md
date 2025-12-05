# Mini Admin - Game Management Service

A simple service for managing and uploading game files to AWS S3 with a React frontend.

## Features

- **Backend (Go):**
  - RESTful API for game CRUD operations
  - Direct S3 file upload support
  - Optional S3 bucket versioning
  - PostgreSQL database for game metadata
  - SQLC for type-safe SQL queries

- **Frontend (React + Vite):**
  - Upload Game tab with drag & drop
  - Games List with edit and delete
  - Simple and clean UI
  - AWS Amplify deployment ready

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 18+
- PostgreSQL 15+
- AWS Account with S3 access
- Docker (optional, for local development)

### Backend Setup

1. **Install dependencies:**
   ```bash
   cd services/mini-admin
   go mod download
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run database migrations:**
   ```bash
   export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/mini_admin?sslmode=disable"
   make migrate-up
   ```

4. **Generate SQLC code:**
   ```bash
   make sqlc
   ```

5. **Run the service:**
   ```bash
   make run
   ```

The API will be available at `http://localhost:8088`

### Frontend Setup

1. **Install dependencies:**
   ```bash
   cd frontends/mini-admin
   npm install
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env to point to your backend
   ```

3. **Run development server:**
   ```bash
   npm run dev
   ```

The frontend will be available at `http://localhost:3001`

### Docker Development

Run both backend and PostgreSQL with Docker Compose:

```bash
cd services/mini-admin
docker-compose -f docker-compose.dev.yaml up
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/games` | List all games |
| GET | `/v1/games/:id` | Get a single game |
| POST | `/v1/games` | Create a new game |
| PUT | `/v1/games/:id` | Update a game |
| DELETE | `/v1/games/:id` | Delete a game (soft delete) |
| POST | `/v1/games/:id/files` | Upload files for a game |
| GET | `/v1/games/:id/url` | Get public S3 URL |
| GET | `/healthz` | Health check |

## Deployment

### Backend (ECS/Fargate)

1. **Build and push Docker image:**
   ```bash
   docker build -t mini-admin:latest -f cmd/mini-admin/Dockerfile .
   docker tag mini-admin:latest <your-ecr-repo>:latest
   docker push <your-ecr-repo>:latest
   ```

2. **Deploy with Terraform:**
   ```hcl
   module "mini_admin" {
     source = "../../modules/ecs-service"
     
     service_name    = "mini-admin"
     container_image = "<your-ecr-repo>:latest"
     # ... other configuration
   }
   ```

### Frontend (AWS Amplify)

1. **Connect your Git repository to AWS Amplify**

2. **Configure build settings:** (uses `amplify.yml`)

3. **Set environment variables:**
   - `VITE_API_URL`: Your backend API URL

4. **Deploy:** Amplify will auto-deploy on git push

### S3 Bucket (Terraform)

```hcl
module "games_bucket" {
  source = "../../modules/s3-static"

  bucket_name           = "mini-admin-games-prod"
  environment           = "prod"
  enable_versioning     = true
  enable_static_website = true
  
  tags = {
    Project = "mini-admin"
  }
}
```

## Configuration

### Backend Environment Variables

- `APP_ENV`: Application environment (development/production)
- `SERVICE_PORT`: HTTP server port (default: 8088)
- `POSTGRES_DSN`: PostgreSQL connection string
- `AWS_REGION`: AWS region for S3
- `AWS_S3_BUCKET`: S3 bucket name
- `AWS_ACCESS_KEY_ID`: AWS access key (optional, uses IAM role in ECS)
- `AWS_SECRET_ACCESS_KEY`: AWS secret key (optional)

### Frontend Environment Variables

- `VITE_API_URL`: Backend API URL

## Development

### Generate SQLC Code

After modifying SQL queries:

```bash
make sqlc
```

### Database Migrations

Create a new migration:

```bash
# Manually create migration file in migrations/
# Format: 000XXX_description.sql
```

Apply migrations:

```bash
export DATABASE_URL="your_database_url"
make migrate-up
```

### Testing

```bash
# Backend
cd services/mini-admin
make test

# Frontend
cd frontends/mini-admin
npm test
```

## Architecture

```
mini-admin/
├── services/mini-admin/        # Go backend
│   ├── cmd/mini-admin/         # Entry point
│   ├── internal/
│   │   ├── config/             # Configuration
│   │   ├── http/               # HTTP handlers & server
│   │   ├── repository/         # Database layer
│   │   └── storage/            # S3 client
│   └── migrations/             # Database migrations
│
├── frontends/mini-admin/       # React frontend
│   ├── src/
│   │   ├── api/                # API client
│   │   ├── features/games/     # Game components
│   │   └── App.tsx             # Main app
│   └── amplify.yml             # Amplify build spec
│
└── platform/terraform/
    └── modules/s3-static/      # S3 Terraform module
```

## License

MIT

