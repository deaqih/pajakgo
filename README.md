# Accounting Web - Data Processing Automation System

Sistem web berbasis Golang untuk mengotomasi pengolahan data akuntansi skala besar dengan rule-based logic.

## Tech Stack

- **Backend**: Go 1.21+, Fiber Framework
- **Frontend**: Go Templates, Tailwind CSS
- **Database**: MySQL 8.0+
- **Cache/Queue**: Redis + Asynq
- **Excel Processing**: excelize library

## Features

- JWT-based authentication
- Excel file upload and parsing
- Rule-based transaction processing
- Batch processing with Asynq workers
- Real-time progress tracking
- Excel export of processed data
- Master data management (Accounts, Rules)

## Project Structure

```
accounting-web/
├── cmd/
│   ├── web/         # Web server entry point
│   └── worker/      # Background worker entry point
├── internal/
│   ├── config/      # Configuration management
│   ├── database/    # Database connections
│   ├── handler/     # HTTP handlers
│   ├── middleware/  # HTTP middleware
│   ├── models/      # Data models
│   ├── repository/  # Database repositories
│   ├── router/      # Route definitions
│   ├── service/     # Business logic
│   ├── utils/       # Utilities
│   └── worker/      # Worker tasks
├── views/           # HTML templates
├── public/          # Static files
├── storage/         # File storage
│   ├── uploads/     # Uploaded files
│   ├── exports/     # Exported files
│   └── logs/        # Log files
└── database/        # Database schema

```

## Installation

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Redis 7.0 or higher

### Setup Steps

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd accounting-web
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup database**
   ```bash
   mysql -u root -p < database/schema.sql
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

5. **Create storage directories**
   ```bash
   mkdir -p storage/uploads storage/exports storage/logs
   ```

6. **Build applications**
   ```bash
   go build -o accounting-web ./cmd/web
   go build -o accounting-worker ./cmd/worker
   ```

## Running the Application

### Development Mode

**Terminal 1 - Web Server:**
```bash
go run cmd/web/main.go
```

**Terminal 2 - Worker:**
```bash
go run cmd/worker/main.go
```

### Production Mode

**Using Systemd:**

Create `/etc/systemd/system/accounting-web.service`:
```ini
[Unit]
Description=Accounting Web Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/accounting-app
ExecStart=/opt/accounting-app/accounting-web
Restart=always

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/accounting-worker.service`:
```ini
[Unit]
Description=Accounting Worker
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/accounting-app
ExecStart=/opt/accounting-app/accounting-worker
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start services:
```bash
sudo systemctl enable accounting-web accounting-worker
sudo systemctl start accounting-web accounting-worker
```

## Configuration

Key environment variables in `.env`:

```env
# Application
APP_PORT=8080
APP_ENV=production

# Database
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=accounting_db
DB_USERNAME=root
DB_PASSWORD=your_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key

# Upload
UPLOAD_MAX_SIZE=104857600  # 100MB
UPLOAD_PATH=./storage/uploads

# Processing
BATCH_SIZE=5000
WORKER_CONCURRENCY=4
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### Dashboard
- `GET /api/v1/dashboard/stats` - Dashboard statistics

### Accounts
- `GET /api/v1/accounts` - List accounts
- `POST /api/v1/accounts` - Create account
- `PUT /api/v1/accounts/:id` - Update account
- `DELETE /api/v1/accounts/:id` - Delete account

### Rules (Koreksi, Obyek, WHT, Tax Keywords)
- `GET /api/v1/koreksi-rules` - List koreksi rules
- `POST /api/v1/koreksi-rules` - Create koreksi rule
- `PUT /api/v1/koreksi-rules/:id` - Update koreksi rule
- `DELETE /api/v1/koreksi-rules/:id` - Delete koreksi rule

### Uploads & Processing
- `POST /api/v1/uploads` - Upload Excel file
- `GET /api/v1/uploads` - List upload sessions
- `GET /api/v1/uploads/:id` - Get session detail
- `GET /api/v1/uploads/:id/transactions` - Get transactions
- `POST /api/v1/uploads/:id/process` - Start processing
- `GET /api/v1/uploads/:id/export` - Export processed data

## Processing Rules

The system implements these processing steps:

1. **Analisa Nature Akun** - Lookup account nature
2. **Koreksi** - Match keterangan with koreksi rules
3. **Obyek** - Match keterangan with obyek rules
4. **Analisa Koreksi-Obyek** - Combine koreksi and obyek
5. **Withholding Tax** - Calculate WHT based on rules
6. **PM DB (Input Tax)** - Calculate input tax
7. **PK CR (Output Tax)** - Calculate output tax

## Default Credentials

- Username: `admin`
- Password: `admin123`

## Performance

Target metrics:
- Upload (100MB): < 30 seconds
- Parse (100k rows): < 2 minutes
- Process (100k rows): < 5 minutes
- Export (100k rows): < 3 minutes

## Development

### Running Tests
```bash
go test ./...
```

### Code Formatting
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

## Troubleshooting

### Database Connection Issues
- Verify MySQL is running: `sudo systemctl status mysql`
- Check database credentials in `.env`
- Ensure database exists: `mysql -u root -p -e "SHOW DATABASES;"`

### Redis Connection Issues
- Verify Redis is running: `sudo systemctl status redis`
- Check Redis configuration: `redis-cli ping`

### Upload Issues
- Check storage directory permissions: `chmod -R 755 storage/`
- Verify `UPLOAD_MAX_SIZE` in `.env`
- Check available disk space

## License

Proprietary - All rights reserved

## Support

For questions or issues, contact the development team.
