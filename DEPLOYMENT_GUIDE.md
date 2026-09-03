# Panduan Deployment KubeNexus ke Kubernetes Menggunakan Helm & Ingress

Panduan ini dirancang agar Anda dapat mendeploy seluruh ekosistem **KubeNexus** (Backend Go Fiber, Frontend Vue 3 + Nginx, PostgreSQL, Redis, dan RBAC) ke cluster Kubernetes dengan **sangat mudah (anti-ribet)**.

---

## 1. Arsitektur Paket Helm & Komponen

```
helm/kubenexus/
├── Chart.yaml                  # Metadata Helm Chart (v0.1.0, App v2.0.0)
├── values.yaml                 # Konfigurasi terpusat (Replica, Ingress, Database, Secrets)
└── templates/
    ├── rbac.yaml               # ServiceAccount & ClusterRole (Wajib agar backend bisa akses K8s API)
    ├── configmap.yaml          # Environment variables aplikasi
    ├── secret.yaml             # Kredensial DB & JWT Secret
    ├── backend-deployment.yaml # Go Fiber API Pod
    ├── backend-service.yaml    # Service Port 3001
    ├── frontend-deployment.yaml# Vue 3 SPA + Nginx Pod
    ├── frontend-service.yaml   # Service Port 80
    ├── ingress.yaml            # Ingress unified routing (/ ke frontend, /api ke backend + WebSocket)
    ├── migration-job.yaml      # Automated DB Migration Hook saat install/upgrade
    ├── postgres.yaml           # Built-in lightweight PostgreSQL (bisa dimatikan jika pakai RDS external)
    └── redis.yaml              # Built-in lightweight Redis (bisa dimatikan jika pakai Redis external)
```

---

## 2. Cara Deploy Paling Cepat (1-Click Automation)

Kami telah menyiapkan script otomatisasi [deploy.ps1](file:///d:/Project/dashboard-kubernetes/deploy.ps1) (untuk Windows PowerShell) dan [deploy.sh](file:///d:/Project/dashboard-kubernetes/deploy.sh) (untuk Linux/WSL/macOS).

### Jalankan Semua (Build Images + Deploy Helm + Cek Status):
```powershell
.\deploy.ps1 -Action all
```
*Script ini otomatis:*
1. Membuild Docker container multi-stage backend (Go 1.25 Alpine), frontend (Node 22 + Nginx), dan migration tool.
2. Membuat namespace `kubenexus`.
3. Menjalankan `helm upgrade --install` (otomatis menggunakan container `alpine/helm` jika Helm CLI belum terpasang di komputer Anda).
4. Menampilkan status Pods, Services, dan Ingress.

### Perintah Bertahap:
```powershell
# Hanya build image Docker
.\deploy.ps1 -Action build

# Hanya deploy / upgrade konfigurasi via Helm
.\deploy.ps1 -Action deploy

# Cek status Pods, Ingress, & Services
.\deploy.ps1 -Action status

# Streaming log backend secara realtime
.\deploy.ps1 -Action logs
```

---

## 3. Cara Deploy Manual Menggunakan Helm CLI

Jika Anda terbiasa menggunakan terminal dengan Helm CLI terpasang:

### Langkah 1: Build Docker Images
```bash
docker build -t kubenexus-backend:latest ./be
docker build -t kubenexus-frontend:latest ./fe
docker build -t kubenexus-migrate:latest -f ./be/Dockerfile.migrate ./be
```

> **Catatan untuk Minikube / Kind:**
> - Jika menggunakan **Minikube**: Jalankan `eval $(minikube docker-env)` sebelum build agar image langsung tersedia di cluster.
> - Jika menggunakan **Kind**: Jalankan `kind load docker-image kubenexus-backend:latest kubenexus-frontend:latest kubenexus-migrate:latest`.

### Langkah 2: Install / Upgrade Helm Chart
```bash
helm upgrade --install kubenexus ./helm/kubenexus -n kubenexus --create-namespace
```

---

## 4. Konfigurasi Ingress & Cara Akses Dashboard

Paket Helm ini sudah menyertakan resource **Ingress** bawaan yang mengarahkan:
- Path `/` ──> Frontend (Vue 3 / Nginx)
- Path `/api` ──> Backend (Go Fiber)
- Path `/api/v1/k8s/ws/` ──> WebSocket Container Terminal (dilengkapi anotasi upgrade)

### Mengakses Melalui Domain Lokal (`kubenexus.local`):
1. Pastikan Ingress Controller aktif di cluster Anda (misal: NGINX Ingress Controller).
   - Di Minikube: `minikube addons enable ingress`
   - Di MicroK8s: `microk8s enable ingress`
2. Tambahkan domain ke file `hosts` Anda:
   - **Windows**: `C:\Windows\System32\drivers\etc\hosts`
   - **Linux / macOS**: `/etc/hosts`
   ```text
   127.0.0.1 kubenexus.local
   ```
   *(Ganti `127.0.0.1` dengan IP Ingress Controller jika menggunakan VM remote/cloud)*.
3. Buka browser: **`http://kubenexus.local`**

### Alternatif Akses Tanpa Ingress (Port-Forward):
Jika cluster Anda belum memiliki Ingress Controller:
```bash
# Forward port Frontend ke 8080
kubectl port-forward svc/kubenexus-frontend 8080:80 -n kubenexus

# Buka di browser:
# http://localhost:8080
```

---

## 5. Kustomisasi Parameter di `values.yaml`

Semua konfigurasi dapat diubah melalui file [values.yaml](file:///d:/Project/dashboard-kubernetes/helm/kubenexus/values.yaml) atau parameter `--set`:

| Parameter | Default | Keterangan |
|---|---|---|
| `backend.replicaCount` | `1` | Jumlah replika Pod Backend |
| `frontend.replicaCount` | `1` | Jumlah replika Pod Frontend |
| `ingress.enabled` | `true` | Aktifkan / matikan Ingress |
| `ingress.hosts[0].host` | `kubenexus.local` | Domain akses dashboard |
| `postgresql.enabled` | `true` | Gunakan Postgres internal (set `false` jika pakai RDS luar) |
| `database.host` | `""` | Host Postgres eksternal jika `postgresql.enabled=false` |
| `database.password` | `""` | Kosongkan untuk **auto-generate password aman (24 karakter)** |
| `redisInternal.enabled`| `true` | Gunakan Redis internal (set `false` jika pakai Redis luar) |
| `jwt.secret` | `""` | Kosongkan untuk **auto-generate signing key aman (32 karakter)** |

---

## 6. Fitur Auto-Generated Secrets (Otomatis & Aman)

Anda **tidak perlu lagi mengisi password atau secret JWT secara manual** di `values.yaml`.
Jika dibiarkan kosong (`""`), Helm akan otomatis:
1. Menghasilkan string acak berkekuatan kriptografis tinggi:
   - `JWT_SECRET`: 32 karakter alfanumerik acak
   - `DB_PASSWORD`: 24 karakter alfanumerik acak
2. **Persistence across Upgrades**: Melalui fungsi `lookup`, Helm mendeteksi Secret yang sudah ada di cluster sehingga secret **tidak akan berubah** saat Anda menjalankan `helm upgrade`.

---

## 7. Cloud CI/CD Pipeline (Build Tanpa Komputer Lokal)

Jika Anda tidak ingin membebani laptop/komputer lokal untuk membuild Docker image, Anda cukup melakukan `git push` ke GitHub atau GitLab!

### Menggunakan GitHub Actions (`.github/workflows/ci-cd.yml`):
1. Push repository monorepo Anda ke GitHub:
   ```bash
   git remote add origin https://github.com/<username>/<repo-name>.git
   git push -u origin master
   ```
2. GitHub Actions akan otomatis:
   - Menjalankan unit test Backend (Go 1.25).
   - Menjalankan build & type-check Frontend (Vue 3 + Vite).
   - Melakukan linting Helm chart.
   - **Membuild & mem-publish 3 image Docker ke GitHub Container Registry (`ghcr.io`)** secara gratis dan otomatis.
3. Anda tinggal mendeploy ke cluster Kubernetes Anda menggunakan image dari cloud:
   ```bash
   helm upgrade --install kubenexus ./helm/kubenexus -n kubenexus --create-namespace \
     --set backend.image.repository=ghcr.io/<username>/<repo-name>/backend \
     --set frontend.image.repository=ghcr.io/<username>/<repo-name>/frontend \
     --set migration.image.repository=ghcr.io/<username>/<repo-name>/migrate
   ```

