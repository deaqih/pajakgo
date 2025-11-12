# Quick Start Guide

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: Download from https://golang.org/dl/
- **MySQL 8.0+**: Database server
- **Redis 7.0+**: In-memory cache and message broker
- **Git**: Version control (optional)

## Quick Installation (Development)

### 1. Setup Environment

```bash
# Clone or navigate to project directory
cd accounting-web

# Copy environment file
cp .env.example .env

# Edit .env with your database credentials
# nano .env
```

### 2. Setup Database

```bash
# Login to MySQL
mysql -u root -p

# Run the following commands in MySQL:
```

```sql
CREATE DATABASE accounting_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE accounting_db;
SOURCE database/schema.sql;
EXIT;
```

### 3. Install Go Dependencies

```bash
go mod download
go mod tidy
```

### 4. Create Storage Directories

```bash
mkdir -p storage/uploads
mkdir -p storage/exports
mkdir -p storage/logs
```

### 5. Run the Application

**Option A: Using Start Script (Recommended)**
```bash
chmod +x scripts/start.sh
./scripts/start.sh
```

**Option B: Manual Start**

Terminal 1 (Web Server):
```bash
go run cmd/web/main.go
```

Terminal 2 (Worker):
```bash
go run cmd/worker/main.go
```

### 6. Access the Application

Open your browser and navigate to:
```
http://localhost:8080
```

**Default Login:**
- Username: `admin`
- Password: `admin123`

## Quick Test

### 1. Login to the System
Use the default credentials above.

### 2. Upload Test Data

Create a test Excel file with these columns:

| Document Type | Document Number | Posting Date | Account | Account Name | Keterangan | Debet | Credit | Net |
|--------------|----------------|--------------|---------|--------------|------------|-------|--------|-----|
| Invoice | INV001 | 2024-01-15 | 1010101 | Cash | Biaya hotel + PPh 23 | 5000000 | 0 | 5000000 |
| Payment | PAY001 | 2024-01-16 | 1010101 | Cash | Bayar sewa kantor + Input Tax | 10000000 | 0 | 10000000 |

### 3. Upload and Process
1. Go to "Upload Data"
2. Select your Excel file
3. Click "Upload"
4. Review the preview
5. Click "Process"
6. Wait for processing to complete
7. Click "Export" to download results

## Troubleshooting

### Issue: Database Connection Failed

**Solution:**
```bash
# Check MySQL is running
sudo systemctl status mysql

# Check database exists
mysql -u root -p -e "SHOW DATABASES;"

# Verify credentials in .env file
```

### Issue: Redis Connection Failed

**Solution:**
```bash
# Check Redis is running
sudo systemctl status redis

# Test Redis connection
redis-cli ping
# Should return: PONG
```

### Issue: Port 8080 Already in Use

**Solution:**
```bash
# Change port in .env file
APP_PORT=8081

# Or kill the process using port 8080
# On Linux/Mac:
lsof -ti:8080 | xargs kill -9

# On Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Issue: Upload Directory Permission Denied

**Solution:**
```bash
# Fix permissions
chmod -R 777 storage/
```

## Production Deployment

For production deployment on Linux servers:

```bash
# Make install script executable
chmod +x scripts/install.sh

# Run installation (requires sudo)
sudo ./scripts/install.sh
```

This will:
- Install system dependencies
- Setup MySQL database
- Configure Nginx reverse proxy
- Setup Supervisor for process management
- Start all services automatically

## Nginx Configuration (Optional)

If you want to use Nginx as a reverse proxy:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    client_max_body_size 100M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## API Testing

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Get Accounts (Protected)
```bash
curl -X GET http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Next Steps

1. **Customize Rules**: Add your own koreksi, obyek, and tax rules
2. **Add Accounts**: Import your account master data
3. **Test Processing**: Upload sample data and verify results
4. **Configure Security**: Change default passwords and JWT secret
5. **Setup Backups**: Configure automated database backups

## Support

For issues or questions:
1. Check the main README.md
2. Review the PRD.md for business logic
3. Check logs in `storage/logs/`

## Useful Commands

```bash
# Build for production
go build -o accounting-web ./cmd/web
go build -o accounting-worker ./cmd/worker

# Run tests
go test ./...

# Check code formatting
go fmt ./...

# View logs (if using Supervisor)
sudo supervisorctl tail -f accounting-web stdout
sudo supervisorctl tail -f accounting-worker stdout

# Restart services (if using Supervisor)
sudo supervisorctl restart accounting-web
sudo supervisorctl restart accounting-worker:*
```

Enjoy using Accounting Web!
