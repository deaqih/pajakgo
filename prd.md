# PRD - Sistem Automasi Pengolahan Data Akuntansi

**Version:** 1.1 | **Date:** November 12, 2025 | **Status:** Draft

---

## 1. Executive Summary

Sistem web berbasis Golang untuk mengotomasi pengolahan data akuntansi skala besar (100k-1M+ records). User mengupload file Excel, sistem memproses data dengan rule-based logic, dan mengekspor hasil dalam format Excel.

**Objectives:**
- Automasi pengisian field analisis data akuntansi
- Mengurangi waktu pemrosesan dan human error
- Support batch processing untuk data besar
- System tracking dan monitoring

**Success Metrics:**
- Processing time: < 5 menit untuk 100k records
- Error rate: < 0.1%
- System uptime: > 99.5%

---

## 2. Tech Stack

| Component | Technology |
|-----------|-----------|
| **Backend** | Go 1.21+, Fiber Framework |
| **Frontend** | Go Templates, Tailwind CSS |
| **Database** | MySQL 8.0+ |
| **Cache/Queue** | Redis + Asynq |
| **Excel** | excelize library |
| **Deployment** | Native (non-Docker), Nginx, Supervisor |

---

## 3. Database Schema (Simplified)

### Core Tables:

**users** - Authentication & user management
```sql
id, username, email, password_hash, role (admin/user), is_active
```

**accounts** - Master data akun
```sql
id, account_code, account_name, account_type, nature, is_active
```

**upload_sessions** - Tracking upload files
```sql
id, session_code, user_id, filename, total_rows, processed_rows, 
failed_rows, status (uploaded/processing/completed/failed)
```

**transaction_data** - Data transaksi
```sql
-- Input Fields:
document_type, document_number, posting_date, account, account_name, 
keterangan, debet, credit, net

-- Output Fields (Auto-filled):
analisa_nature_akun, analisa_koreksi_obyek, koreksi, obyek,
um_pajak_db, pm_db, wth_21_cr, wth_23_cr, wth_26_cr, wth_4_2_cr, 
wth_15_cr, pk_cr, analisa_tambahan
```

### Rule Tables:

**koreksi_rules** - Mapping keyword → kategori koreksi
```sql
keyword, value, priority, is_active
```

**obyek_rules** - Mapping keyword → kategori obyek
```sql
keyword, value, priority, is_active
```

**withholding_tax_rules** - WHT calculation rules
```sql
keyword, tax_type (wth_21/23/26/4_2/15), tax_rate, priority, is_active
```

**tax_keywords** - Keywords untuk input/output tax
```sql
keyword, tax_category (input_tax/output_tax), priority, is_active
```

---

## 4. Core Features

### 4.1 Authentication
- JWT-based login/logout
- Role-based access (Admin/User)
- Session management (24h access, 7d refresh token)

### 4.2 Dashboard
- Upload statistics (total, by status, by period)
- Record processing metrics
- Recent activity list
- System health status

### 4.3 Master Data Management
**Accounts:**
- CRUD operations
- Import/Export Excel
- Search, filter, pagination

**Rule Tables** (Koreksi, Obyek, WHT, Tax Keywords):
- CRUD operations
- Import/Export Excel
- Priority management
- Active/inactive toggle

### 4.4 Data Upload & Processing

**Upload Flow:**
1. User upload Excel file (max 100MB)
2. System validates format & structure
3. Generate unique session code
4. Stream parse & batch insert to DB
5. Return preview & session info

**Processing Flow:**
1. Click "Process" button
2. Job queued in Redis (Asynq)
3. Worker processes in batches (5000 rows/batch)
4. Real-time progress tracking
5. Update results to database
6. Notification on completion

**Export:**
- Export processed data to Excel
- All input + output fields
- Summary sheet included

---

## 5. Processing Rules Logic

### Sequential Processing Steps:

#### **STEP 1: Analisa Nature Akun**
```
Lookup account in accounts table → get nature field
Output: analisa_nature_akun
```

#### **STEP 2: Koreksi**
```
Match keterangan with koreksi_rules.keyword (by priority)
First match → koreksi_rules.value
Output: koreksi
```

#### **STEP 3: Obyek**
```
Match keterangan with obyek_rules.keyword (by priority)
First match → obyek_rules.value
Output: obyek
```

#### **STEP 4: Analisa Koreksi - Obyek**
```
Combine: "{koreksi} - {obyek}"
Output: analisa_koreksi_obyek
```

#### **STEP 5: Withholding Tax (21, 23, 26, 4.2, 15)**
```
IF keterangan contains WHT keyword AND credit > 0:
  Calculate: credit × tax_rate
  Assign to corresponding wth_XX_cr field
```

Example: "PPh 23" → wth_23_cr = credit × 0.02

#### **STEP 6: PM DB (Input Tax)**
```
IF keterangan contains input_tax keyword AND debet > 0:
  pm_db = debet
ELSE:
  pm_db = 0
```

#### **STEP 7: PK CR (Output Tax)**
```
IF keterangan contains output_tax keyword AND credit > 0:
  pk_cr = credit
ELSE:
  pk_cr = 0
```

#### **STEP 8 & 9: Pending**
- `um_pajak_db` - Logic TBD
- `analisa_tambahan` - Logic TBD

---

## 6. Processing Example

**Input:**
```
Keterangan: "Biaya entertaint hotel + PPh 23 + Input Tax"
Account: 1010101 (nature: "Asset")
Debet: 5,000,000
Credit: 0
```

**Processing:**
1. Analisa Nature: "Asset" (from accounts)
2. Koreksi: "Biaya Entertainment" (matched "entertaint")
3. Obyek: "Hotel" (matched "hotel")
4. Analisa Koreksi-Obyek: "Biaya Entertainment - Hotel"
5. WHT: wth_23_cr = 0 (credit = 0, condition not met)
6. PM DB: 5,000,000 (matched "Input Tax" AND debet > 0)
7. PK CR: 0 (no output tax keyword)

**Output:**
```
analisa_nature_akun: "Asset"
analisa_koreksi_obyek: "Biaya Entertainment - Hotel"
koreksi: "Biaya Entertainment"
obyek: "Hotel"
pm_db: 5,000,000
wth_21_cr to wth_15_cr: 0
pk_cr: 0
```

---

## 7. API Endpoints (Key Routes)

### Authentication
```
POST /api/v1/auth/login
POST /api/v1/auth/logout
```

### Dashboard
```
GET /api/v1/dashboard/stats
```

### Master Data
```
GET|POST|PUT|DELETE /api/v1/accounts
GET|POST /api/v1/accounts/import|export
GET|POST|PUT|DELETE /api/v1/koreksi-rules
GET|POST|PUT|DELETE /api/v1/obyek-rules
GET|POST|PUT|DELETE /api/v1/withholding-tax-rules
GET|POST|PUT|DELETE /api/v1/tax-keywords
```

### Upload & Processing
```
POST /api/v1/uploads (multipart/form-data)
GET /api/v1/uploads
GET /api/v1/uploads/:id
GET /api/v1/uploads/:id/transactions
POST /api/v1/uploads/:id/process
GET /api/v1/uploads/:id/export
GET /api/v1/jobs/:job_id/progress
```

---

## 8. Deployment (Native)

### Requirements:
- Ubuntu 22.04 LTS / CentOS 8
- CPU: 4 cores | RAM: 16 GB | Storage: 100 GB SSD
- Go 1.21+, MySQL 8.0, Redis 7.0, Nginx

### Setup:
```bash
# Build
go build -o accounting-web ./cmd/web
go build -o accounting-worker ./cmd/worker

# Deploy to /opt/accounting-app
# Configure Nginx as reverse proxy
# Use Supervisor or Systemd for process management
# Setup 4 worker instances for parallel processing
```

### Monitoring:
- Health check endpoint: `/health`
- Automated monitoring script (cron every 5 min)
- Daily database backup (2 AM)
- Weekly file backup (Sunday 3 AM)

---

## 9. Performance Targets

| Metric | Target |
|--------|--------|
| Upload (100MB) | < 30 seconds |
| Parse (100k rows) | < 2 minutes |
| Process (100k rows) | < 5 minutes |
| Export (100k rows) | < 3 minutes |

**Optimization:**
- Batch processing (5000 rows/batch)
- Streaming Excel read/write
- Database indexing & connection pooling
- Redis caching for master data
- Concurrent processing with goroutine pool

---

## 10. Security

- HTTPS/TLS 1.3 enforced
- Bcrypt password hashing (cost 12)
- JWT authentication with refresh tokens
- SQL injection prevention (parameterized queries)
- XSS & CSRF protection
- File type & size validation
- Rate limiting for login attempts
- Comprehensive audit logging

---

## 11. Menu Structure

```
├── Dashboard
│   └── Statistics & Summary
│
├── Master Data
│   ├── Accounts
│   ├── Koreksi Rules
│   ├── Obyek Rules
│   ├── Withholding Tax Rules
│   └── Tax Keywords
│
├── Transaction Data
│   ├── Upload Data
│   ├── Upload Sessions List
│   └── Session Detail
│       ├── Transaction List
│       ├── Process Button
│       └── Export Data
│
└── Settings
    ├── User Profile
    └── Change Password
```


## 13. Open Questions

1. **UM Pajak DB** - Apa logic yang diinginkan?
2. **Analisa Tambahan** - Apa rule/logic yang diperlukan?
3. **Multiple WHT** - Apakah satu transaksi bisa punya multiple WHT (PPh 21 + 23)?
4. **Priority Handling** - Jika multiple keywords match, selain priority apakah ada rule lain?
5. **Data Retention** - Berapa lama data upload disimpan?
6. **Notification** - Perlu email notification setelah processing selesai?

---

## 14. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Large file timeout | Streaming processing, increased timeout |
| Memory overflow | Batch processing, optimize memory |
| Processing rule complexity | Iterative development, thorough testing |
| DB performance | Proper indexing, partitioning |
| Concurrent conflicts | Job queue, proper locking |

---

**Document prepared by Product Team**
**For questions: Contact product owner**