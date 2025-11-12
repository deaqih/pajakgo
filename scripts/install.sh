#!/bin/bash
# Installation script for Accounting Web

set -e

echo "==================================="
echo "Accounting Web - Installation"
echo "==================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
   echo "Please run as root (use sudo)"
   exit 1
fi

# Variables
APP_DIR="/opt/accounting-app"
APP_USER="www-data"

# Install system dependencies
echo "Installing system dependencies..."
apt-get update
apt-get install -y mysql-server redis-server nginx supervisor

# Create application directory
echo "Creating application directory..."
mkdir -p $APP_DIR
mkdir -p $APP_DIR/storage/uploads
mkdir -p $APP_DIR/storage/exports
mkdir -p $APP_DIR/storage/logs

# Copy files
echo "Copying application files..."
cp -r ./* $APP_DIR/
cd $APP_DIR

# Set permissions
echo "Setting permissions..."
chown -R $APP_USER:$APP_USER $APP_DIR
chmod -R 755 $APP_DIR
chmod -R 777 $APP_DIR/storage

# Setup database
echo "Setting up database..."
mysql -u root <<EOF
CREATE DATABASE IF NOT EXISTS accounting_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'accounting_user'@'localhost' IDENTIFIED BY 'accounting_pass123';
GRANT ALL PRIVILEGES ON accounting_db.* TO 'accounting_user'@'localhost';
FLUSH PRIVILEGES;
EOF

# Import database schema
mysql -u root accounting_db < database/schema.sql

# Create .env file
echo "Creating .env file..."
cat > $APP_DIR/.env <<EOF
APP_NAME=Accounting Web
APP_ENV=production
APP_PORT=8080
APP_URL=http://localhost

DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=accounting_db
DB_USERNAME=accounting_user
DB_PASSWORD=accounting_pass123
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=$(openssl rand -base64 32)
JWT_ACCESS_EXPIRE=24h
JWT_REFRESH_EXPIRE=168h

UPLOAD_MAX_SIZE=104857600
UPLOAD_PATH=$APP_DIR/storage/uploads

BATCH_SIZE=5000
WORKER_CONCURRENCY=4

ASYNQ_REDIS_ADDR=localhost:6379
ASYNQ_REDIS_PASSWORD=
ASYNQ_REDIS_DB=0
EOF

# Build applications
echo "Building applications..."
go build -o $APP_DIR/accounting-web ./cmd/web
go build -o $APP_DIR/accounting-worker ./cmd/worker

# Configure Supervisor
echo "Configuring Supervisor..."
cat > /etc/supervisor/conf.d/accounting-web.conf <<EOF
[program:accounting-web]
command=$APP_DIR/accounting-web
directory=$APP_DIR
autostart=true
autorestart=true
user=$APP_USER
redirect_stderr=true
stdout_logfile=$APP_DIR/storage/logs/web.log

[program:accounting-worker]
command=$APP_DIR/accounting-worker
directory=$APP_DIR
autostart=true
autorestart=true
user=$APP_USER
redirect_stderr=true
stdout_logfile=$APP_DIR/storage/logs/worker.log
numprocs=4
process_name=%(program_name)s_%(process_num)02d
EOF

# Configure Nginx
echo "Configuring Nginx..."
cat > /etc/nginx/sites-available/accounting-web <<EOF
server {
    listen 80;
    server_name localhost;

    client_max_body_size 100M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

ln -sf /etc/nginx/sites-available/accounting-web /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Restart services
echo "Restarting services..."
supervisorctl reread
supervisorctl update
supervisorctl start accounting-web:*
supervisorctl start accounting-worker:*

nginx -t
systemctl restart nginx

echo ""
echo "==================================="
echo "Installation completed!"
echo "==================================="
echo ""
echo "Application URL: http://localhost"
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: admin123"
echo ""
echo "Logs location: $APP_DIR/storage/logs"
echo ""
echo "To check status:"
echo "  sudo supervisorctl status"
echo ""
