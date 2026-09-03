# KubeNexus (v2.0)

> **Cloud-Native Kubernetes Orchestration & Observability Control Plane**

[![GitHub Actions CI](https://github.com/your-username/dashboard-kubernetes/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/your-username/dashboard-kubernetes/actions)
[![Helm Chart](https://img.shields.io/badge/Helm%20Chart-v0.1.0-blue.svg)](./helm/kubenexus)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.5-4FC08D.svg)](https://vuejs.org)
[![TailwindCSS](https://img.shields.io/badge/Tailwind-4.0-38B2AC.svg)](https://tailwindcss.com)

KubeNexus is a centralized, cloud-native Kubernetes dashboard engineered for DevOps, SREs, and platform engineers. It provides interactive workload orchestration, container web terminal streaming, real-time cluster telemetry, dynamic manifest editor, and enterprise Role-Based Access Control (RBAC).

---

## 📁 Monorepo Structure

```
dashboard-kubernetes/
├── be/                         # Go 1.25 Fiber Backend API (client-go in-cluster manager)
│   ├── cmd/                    # Entrypoints (server, migrate)
│   ├── internal/module/        # K8s, Auth, User RBAC domain modules
│   ├── migrations/             # PostgreSQL database migrations
│   └── Dockerfile              # Multi-stage production container
├── fe/                         # Vue 3 + Vite + PrimeVue + TailwindCSS Frontend SPA
│   ├── src/                    # Components, Stores, Pages, Composables
│   ├── nginx.conf              # Production Nginx reverse proxy with WebSocket support
│   └── Dockerfile              # Multi-stage production container
├── helm/
│   └── kubenexus/              # Official Helm Chart (Package for Kubernetes)
│       ├── Chart.yaml          # Chart metadata
│       ├── values.yaml         # Centralized configuration
│       ├── README.md           # Detailed Helm chart documentation & parameters table
│       └── templates/          # K8s Deployments, Services, RBAC, Ingress, Jobs
├── .github/workflows/          # GitHub Actions CI/CD Pipeline
│   └── ci-cd.yml               # Automated test, build & push to GHCR, package chart
├── deploy.ps1                  # 1-Click deployment automation script for Windows
├── deploy.sh                   # 1-Click deployment automation script for Linux/WSL
└── DEPLOYMENT_GUIDE.md         # Comprehensive deployment & ingress guide
```

---

## 🚀 Quick Start: Deploying to Kubernetes via Helm

The recommended way to deploy KubeNexus is using the official **Helm Chart**.

### 1. Install KubeNexus via Helm
Run the following command from the root of the repository:

```bash
helm upgrade --install kubenexus ./helm/kubenexus \
  --namespace kubenexus \
  --create-namespace
```

### 2. Retrieve Initial Admin Credentials
KubeNexus automatically generates a cryptographically secure, random 16-character administrator password:

```bash
# Get Admin Login Email (Default: admin@kubenexus.local)
kubectl get secret --namespace kubenexus kubenexus-secret \
  -o jsonpath="{.data.ADMIN_EMAIL}" | base64 -d && echo

# Get Auto-Generated Admin Password
kubectl get secret --namespace kubenexus kubenexus-secret \
  -o jsonpath="{.data.ADMIN_PASSWORD}" | base64 -d && echo
```

### 3. Access the Dashboard via Ingress
1. Add the domain to your `hosts` file (`C:\Windows\System32\drivers\etc\hosts` or `/etc/hosts`):
   ```text
   127.0.0.1 kubenexus.local
   ```
2. Open your browser at: **`http://kubenexus.local`**

*(For full parameter details, refer to the [Helm Chart Documentation](./helm/kubenexus/README.md).)*

---

## 🛠️ 1-Click Automation Scripts

If you want a fast, interactive deployment experience:

- **Windows PowerShell**:
  ```powershell
  .\deploy.ps1 -Action all
  ```
- **Linux / macOS / WSL**:
  ```bash
  chmod +x deploy.sh
  ./deploy.sh all
  ```

*These scripts automatically handle Docker builds, Helm release installation, and print your login credentials upon completion.*

---

## 🔄 Cloud CI/CD Pipeline (GitHub Actions)

This monorepo comes with an automated GitHub Actions pipeline in [`.github/workflows/ci-cd.yml`](./.github/workflows/ci-cd.yml):

- **Continuous Integration**: Runs Go 1.25 tests, Vue 3 type checking, and Helm lint on every Pull Request and commit.
- **Continuous Delivery**: Automatically builds multi-stage Docker images and pushes them to **GitHub Container Registry (`ghcr.io`)**.
- **Chart Packaging**: Automatically packages the Helm chart into a `.tgz` artifact.

You can point your cluster directly to the cloud images in `values.yaml` or via `--set`:
```bash
helm upgrade --install kubenexus ./helm/kubenexus -n kubenexus --create-namespace \
  --set backend.image.repository=ghcr.io/<your-username>/dashboard-kubernetes/backend \
  --set frontend.image.repository=ghcr.io/<your-username>/dashboard-kubernetes/frontend \
  --set migration.image.repository=ghcr.io/<your-username>/dashboard-kubernetes/migrate
```

---

## 💻 Local Development

### Backend (Go 1.25)
```bash
cd be
go mod download
go run cmd/server/main.go
```

### Frontend (Vue 3 + Vite)
```bash
cd fe
npm install
npm run dev
```

---

## 📖 Documentation Links

- [Helm Chart Reference & Parameters](./helm/kubenexus/README.md)
- [Comprehensive Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [RBAC & User Management Architecture](./USER_MANAGEMENT_IMPLEMENTATION.md)

---

## 📄 License

Distributed under the MIT License.
