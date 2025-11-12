# Cara Memulai - Accounting Web System

## Ringkasan Sistem

Sistem web untuk mengotomasi pengolahan data akuntansi dengan kemampuan:
- Upload file Excel dengan data transaksi
- Pemrosesan otomatis berdasarkan rules yang dikonfigurasi
- Export hasil pemrosesan ke Excel
- Manajemen master data (Accounts, Rules)

## Persiapan Awal

### Yang Dibutuhkan:
1. **Go 1.21 atau lebih tinggi** - https://golang.org/dl/
2. **MySQL 8.0 atau lebih tinggi**
3. **Redis 7.0 atau lebih tinggi**
4. **Text Editor** (VS Code, Notepad++, dll)

## Langkah 1: Setup Database

### Install MySQL (jika belum ada)

**Windows:**
- Download dari https://dev.mysql.com/downloads/installer/
- Install dan catat password root

**Linux:**
```bash
sudo apt update
sudo apt install mysql-server
sudo mysql_secure_installation
```

### Buat Database

```bash
# Login ke MySQL
mysql -u root -p

# Jalankan perintah berikut:
```

```sql
CREATE DATABASE accounting_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE accounting_db;
SOURCE database/schema.sql;
EXIT;
```

## Langkah 2: Setup Redis

### Install Redis

**Windows:**
- Download dari https://github.com/microsoftarchive/redis/releases
- Install dan jalankan sebagai service

**Linux:**
```bash
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

### Test Redis
```bash
redis-cli ping
# Harus return: PONG
```

## Langkah 3: Konfigurasi Aplikasi

### Copy File Environment
```bash
cp .env.example .env
```

### Edit File .env

Buka file `.env` dan sesuaikan dengan konfigurasi Anda:

```env
# Database (sesuaikan dengan setting MySQL Anda)
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=accounting_db
DB_USERNAME=root
DB_PASSWORD=your_mysql_password_here

# Redis (biasanya tidak perlu diubah)
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT Secret (untuk production, ganti dengan random string)
JWT_SECRET=your-secret-key-change-this

# Upload settings
UPLOAD_MAX_SIZE=104857600  # 100MB
UPLOAD_PATH=./storage/uploads

# Processing
BATCH_SIZE=5000
WORKER_CONCURRENCY=4
```

## Langkah 4: Install Dependencies

```bash
go mod download
go mod tidy
```

## Langkah 5: Buat Direktori Storage

```bash
mkdir -p storage/uploads
mkdir -p storage/exports
mkdir -p storage/logs
```

**Windows (Command Prompt):**
```cmd
mkdir storage\uploads
mkdir storage\exports
mkdir storage\logs
```

## Langkah 6: Jalankan Aplikasi

### Cara 1: Menggunakan Script (Recommended)

**Linux/Mac:**
```bash
chmod +x scripts/start.sh
./scripts/start.sh
```

**Windows (Git Bash):**
```bash
bash scripts/start.sh
```

### Cara 2: Manual (Buka 2 Terminal)

**Terminal 1 - Web Server:**
```bash
go run cmd/web/main.go
```

**Terminal 2 - Worker:**
```bash
go run cmd/worker/main.go
```

### Cara 3: Build dan Run

```bash
# Build
go build -o accounting-web.exe ./cmd/web     # Windows
go build -o accounting-worker.exe ./cmd/worker

# atau di Linux/Mac:
go build -o accounting-web ./cmd/web
go build -o accounting-worker ./cmd/worker

# Run
./accounting-web &
./accounting-worker &
```

## Langkah 7: Akses Aplikasi

Buka browser dan akses:
```
http://localhost:8080
```

**Login dengan:**
- Username: `admin`
- Password: `admin123`

## Tutorial Penggunaan

### 1. Login ke Sistem

![Login Page](http://localhost:8080/login)

Gunakan credentials default:
- Username: `admin`
- Password: `admin123`

### 2. Lihat Dashboard

Setelah login, Anda akan melihat:
- Statistics cards (Total uploads, Completed, Processing, Failed)
- Quick actions buttons
- Menu untuk master data

### 3. Upload Data Excel

#### Format Excel yang Dibutuhkan:

Buat file Excel dengan kolom berikut:

| Column | Type | Example | Keterangan |
|--------|------|---------|------------|
| Document Type | Text | Invoice | Tipe dokumen |
| Document Number | Text | INV001 | Nomor dokumen |
| Posting Date | Date | 2024-01-15 | Tanggal posting |
| Account | Text | 1010101 | Kode akun |
| Account Name | Text | Cash | Nama akun |
| Keterangan | Text | Biaya hotel + PPh 23 | Deskripsi |
| Debet | Number | 5000000 | Nilai debet |
| Credit | Number | 0 | Nilai kredit |
| Net | Number | 5000000 | Nilai net |

#### Contoh Data:

```
Document Type | Document Number | Posting Date | Account | Account Name | Keterangan | Debet | Credit | Net
Invoice | INV001 | 2024-01-15 | 1010101 | Cash | Biaya hotel + PPh 23 | 5000000 | 0 | 5000000
Invoice | INV002 | 2024-01-16 | 1010101 | Cash | Sewa kantor + Input Tax | 10000000 | 0 | 10000000
Payment | PAY001 | 2024-01-17 | 2010101 | AP | Bayar supplier + PPh 23 | 0 | 2000000 | -2000000
```

#### Langkah Upload:

1. Klik "Upload Data" di dashboard
2. Pilih file Excel Anda
3. Klik "Upload"
4. Tunggu proses parsing (sistem akan menampilkan preview 10 rows pertama)
5. Jika data sudah benar, klik "Process"
6. Tunggu hingga status berubah menjadi "Completed"
7. Klik "Export" untuk download hasil

### 4. Kelola Master Data

#### Accounts Management

1. Klik "Accounts" di menu
2. Anda akan melihat list semua accounts
3. Untuk menambah: Klik "Add New"
4. Untuk edit: Klik icon edit pada row
5. Untuk delete: Klik icon delete

**Field yang dibutuhkan:**
- Account Code: Kode akun (contoh: 1010101)
- Account Name: Nama akun (contoh: Cash on Hand)
- Account Type: Tipe akun (contoh: Current Asset)
- Nature: Nature akun (Asset/Liability/Equity/Revenue/Expense)
- Is Active: Status aktif

#### Rules Management

**Koreksi Rules:**
- Keyword: Kata kunci yang dicari di kolom Keterangan
- Value: Nilai koreksi yang akan diisi
- Priority: Priority matching (lebih tinggi = dicek lebih dulu)

Contoh:
- Keyword: "hotel" → Value: "Biaya Entertainment" → Priority: 100
- Keyword: "sewa" → Value: "Biaya Sewa" → Priority: 90

**Obyek Rules:**
- Sama seperti Koreksi Rules, tapi untuk obyek

Contoh:
- Keyword: "kantor" → Value: "Office" → Priority: 100
- Keyword: "mobil" → Value: "Vehicle" → Priority: 90

**Withholding Tax Rules:**
- Keyword: Kata kunci PPh di Keterangan
- Tax Type: Jenis PPh (wth_21/wth_23/wth_26/wth_4_2/wth_15)
- Tax Rate: Rate pajak (contoh: 0.02 untuk 2%)

Contoh:
- Keyword: "PPh 23" → Tax Type: wth_23 → Tax Rate: 0.02 (2%)
- Keyword: "PPh 21" → Tax Type: wth_21 → Tax Rate: 0.05 (5%)

**Tax Keywords:**
- Keyword: Kata kunci tax di Keterangan
- Tax Category: input_tax atau output_tax

Contoh:
- Keyword: "Input Tax" → Category: input_tax
- Keyword: "Output Tax" → Category: output_tax

## Cara Kerja Sistem

### Processing Flow:

1. **Upload**: File Excel diupload dan di-parse
2. **Validation**: Sistem validasi format dan struktur
3. **Insert**: Data dimasukkan ke database
4. **Queue**: Processing task dimasukkan ke queue (Redis)
5. **Processing**: Worker mengambil task dan memproses:
   - Analisa Nature Akun (lookup dari accounts)
   - Match Koreksi (dari koreksi_rules)
   - Match Obyek (dari obyek_rules)
   - Combine Koreksi-Obyek
   - Calculate Withholding Tax
   - Calculate Input Tax (PM DB)
   - Calculate Output Tax (PK CR)
6. **Update**: Hasil disimpan ke database
7. **Export**: User bisa download hasil dalam format Excel

### Contoh Processing:

**Input:**
```
Keterangan: "Biaya hotel Jakarta + PPh 23 + Input Tax"
Account: 1010101 (Nature: Asset)
Debet: 5,000,000
Credit: 0
```

**Processing:**
1. Nature: "Asset" (dari accounts table)
2. Koreksi: "Biaya Entertainment" (matched "hotel")
3. Obyek: "Hotel" (matched "hotel")
4. Koreksi-Obyek: "Biaya Entertainment - Hotel"
5. WHT 23: 0 (karena Credit = 0)
6. PM DB: 5,000,000 (matched "Input Tax" DAN Debet > 0)
7. PK CR: 0 (tidak ada Output Tax keyword)

**Output:**
```
analisa_nature_akun: "Asset"
koreksi: "Biaya Entertainment"
obyek: "Hotel"
analisa_koreksi_obyek: "Biaya Entertainment - Hotel"
pm_db: 5,000,000
wth_21_cr sampai wth_15_cr: 0
pk_cr: 0
```

## Troubleshooting

### Problem: Tidak bisa login

**Solusi:**
1. Pastikan database sudah di-setup dengan benar
2. Check table users ada dan berisi data default
3. Check credentials: username=admin, password=admin123

### Problem: Error "Database connection failed"

**Solusi:**
1. Check MySQL sudah running
2. Check username/password di file .env
3. Check database 'accounting_db' sudah dibuat
4. Test koneksi: `mysql -u root -p -e "USE accounting_db;"`

### Problem: Error "Redis connection failed"

**Solusi:**
1. Check Redis sudah running
2. Test dengan: `redis-cli ping`
3. Check REDIS_HOST dan REDIS_PORT di .env

### Problem: Upload gagal

**Solusi:**
1. Check ukuran file (max 100MB)
2. Check format file (.xlsx atau .xls)
3. Check struktur Excel (harus ada 9 kolom)
4. Check permission folder storage/uploads (harus writable)

### Problem: Processing tidak jalan

**Solusi:**
1. Pastikan Worker sudah running
2. Check Redis connection
3. Check logs di storage/logs/
4. Restart worker: kill process dan jalankan ulang

### Problem: Port 8080 sudah dipakai

**Solusi:**
1. Ubah APP_PORT di .env (contoh: 8081)
2. Atau kill process yang pakai port 8080:
   - Windows: `netstat -ano | findstr :8080` kemudian `taskkill /PID <PID> /F`
   - Linux: `lsof -ti:8080 | xargs kill -9`

## Tips Penggunaan

### 1. Persiapkan Rules Dulu
Sebelum upload data, pastikan sudah setup:
- Master accounts yang lengkap
- Koreksi rules untuk kata kunci yang sering muncul
- Obyek rules untuk kata kunci yang sering muncul
- Withholding tax rules sesuai kebutuhan
- Tax keywords (Input Tax, Output Tax, dll)

### 2. Test dengan Data Kecil
Upload file dengan 10-50 rows dulu untuk test:
- Apakah parsing berhasil?
- Apakah rules matching dengan benar?
- Apakah hasil sesuai ekspektasi?

### 3. Monitor Processing
Saat processing data besar:
- Check progress via API atau UI
- Monitor logs jika ada error
- Tunggu sampai status "Completed" sebelum export

### 4. Backup Data
Lakukan backup database secara berkala:
```bash
mysqldump -u root -p accounting_db > backup_$(date +%Y%m%d).sql
```

### 5. Optimasi Performance
Untuk data sangat besar (>500k rows):
- Naikkan BATCH_SIZE di .env (contoh: 10000)
- Naikkan WORKER_CONCURRENCY (contoh: 8)
- Pastikan server punya RAM cukup

## Perintah Berguna

```bash
# Build aplikasi
make build

# Run development mode
make dev

# Run tests
make test

# Setup database
make db-setup

# Clean build files
make clean

# Format code
make fmt
```

## Kontak Support

Jika ada masalah atau pertanyaan:
1. Check dokumentasi di README.md
2. Check PRD.md untuk business logic
3. Check logs di storage/logs/
4. Hubungi tim development

## Selamat Menggunakan!

Sistem sudah siap digunakan. Mulai dengan:
1. ✓ Login
2. ✓ Setup master data
3. ✓ Upload data test
4. ✓ Process dan verify hasil
5. ✓ Scale up untuk production data
