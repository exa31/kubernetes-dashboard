# KubeNexus Helm Chart

[![Helm Version](https://img.shields.io/badge/Helm-v3.12%2B-blue.svg)](https://helm.sh)
[![Kubernetes Version](https://img.shields.io/badge/Kubernetes-1.24%2B-brightgreen.svg)](https://kubernetes.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Official Helm Chart to deploy **KubeNexus** — an Enterprise Cloud-Native Kubernetes Orchestration & Observability Control Plane.

---

## Architecture Overview

This Helm chart deploys the complete KubeNexus microservice ecosystem into your Kubernetes cluster:

- **Backend API (`kubenexus-backend`)**: High-performance Go Fiber service interacting with Kubernetes API via `client-go` (in-cluster RBAC) and streaming real-time pod logs/web terminal.
- **Frontend SPA (`kubenexus-frontend`)**: Vue 3 + TailwindCSS + PrimeVue dashboard served via production-grade Nginx.
- **Database & Migration (`kubenexus-migrate` & PostgreSQL)**: PostgreSQL database with automated Helm lifecycle pre-install / pre-upgrade migration jobs.
- **Caching & Real-time State (`redis`)**: Redis service for distributed JWT token blacklisting and SSE/WebSocket pub-sub.
- **Ingress Controller Routing**: Single unified Ingress mapping `/` to Frontend, `/api` to Backend, with automated WebSocket upgrade headers.

---

## Prerequisites

Before deploying KubeNexus, ensure you have:

1. **Kubernetes Cluster**: Version `1.24` or higher.
2. **Helm**: Version `3.12.0` or higher.
3. **Ingress Controller**: Installed in the cluster (e.g., `ingress-nginx`, Traefik, or cloud provider ingress).
4. **Storage Class** *(Optional)*: Default StorageClass configured if enabling persistent storage for internal PostgreSQL.

---

## Quick Start Installation

### 1. Deploy KubeNexus with Default Settings
Run the following command to deploy KubeNexus in a dedicated `kubenexus` namespace:

```bash
helm install kubenexus ./helm/kubenexus \
  --namespace kubenexus \
  --create-namespace
```

### 2. Retrieve the Auto-Generated Admin Credentials
By default, KubeNexus automatically generates a secure, random **16-character administrator password** and stores it in the Kubernetes Secret.

Run the following commands to get your credentials:

```bash
# Get Admin Login Email
kubectl get secret --namespace kubenexus kubenexus-secret \
  -o jsonpath="{.data.ADMIN_EMAIL}" | base64 -d && echo

# Get Admin Random Password
kubectl get secret --namespace kubenexus kubenexus-secret \
  -o jsonpath="{.data.ADMIN_PASSWORD}" | base64 -d && echo
```

---

## Accessing the Dashboard

### 1. Custom Ingress with Let's Encrypt (Recommended for Production)
By default, **Ingress is disabled (`ingress.enabled: false`)** so you can define your own Ingress resource with your custom domain and SSL/TLS certificates (e.g. via **Cert-Manager** & **Let's Encrypt**).

A complete template is available at `helm/kubenexus/examples/ingress-letsencrypt.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubenexus-ingress
  namespace: kubenexus
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - dashboard.yourdomain.com
      secretName: kubenexus-tls-cert
  rules:
    - host: dashboard.yourdomain.com
      http:
        paths:
          # Only 1 path needed! The Frontend Nginx internally proxies /api and WebSockets to the Backend
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kubenexus-frontend
                port:
                  number: 80
```

Apply the file to your cluster:
```bash
kubectl apply -f helm/kubenexus/examples/ingress-letsencrypt.yaml
```

### 2. Via Port-Forwarding (For Local Testing)
To test the dashboard immediately without an Ingress:

```bash
kubectl port-forward svc/kubenexus-frontend 8080:80 -n kubenexus
```
Then visit: `http://localhost:8080`

### 3. Or Enable Chart-Managed Ingress
If you prefer the Helm chart to manage the Ingress automatically, set `ingress.enabled: true`:

```bash
helm upgrade --install kubenexus ./helm/kubenexus -n kubenexus \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=kubenexus.yourdomain.com
```

---

## Upgrading KubeNexus

To update KubeNexus with a new version, image tag, or configuration values:

```bash
helm upgrade kubenexus ./helm/kubenexus \
  --namespace kubenexus \
  --values my-custom-values.yaml
```

> **Zero-Worry Security Guarantee:**
> Thanks to Helm's built-in `lookup` function, your auto-generated passwords (`ADMIN_PASSWORD`, `JWT_SECRET`, `DB_PASSWORD`) **will persist across upgrades** and will not be overwritten.

---

## Uninstalling the Chart

To remove the KubeNexus release and all associated resources:

```bash
helm uninstall kubenexus --namespace kubenexus
```

To delete the namespace and persistent volume claims:
```bash
kubectl delete namespace kubenexus
```

---

## Configuration Parameters

The following table lists the configurable parameters of the KubeNexus chart and their default values.

### Global & Administration
| Parameter | Description | Default |
|---|---|---|
| `global.nameOverride` | Override chart name | `""` |
| `global.fullnameOverride` | Override full release name | `""` |
| `admin.email` | Initial administrator login email | `"admin@kubenexus.local"` |
| `admin.password` | Administrator password. If empty (`""`), automatically generated (16 chars) | `""` (Auto-generated) |
| `jwt.secret` | JWT signing key. If empty (`""`), automatically generated (32 chars) | `""` (Auto-generated) |
| `jwt.accessDuration` | JWT Access token expiration | `"15m"` |
| `jwt.refreshDuration` | JWT Refresh token expiration | `"168h"` |

### Backend Service (`backend`)
| Parameter | Description | Default |
|---|---|---|
| `backend.replicaCount` | Number of backend pod replicas | `1` |
| `backend.image.repository` | Backend Docker image repository | `"kubenexus-backend"` |
| `backend.image.tag` | Backend image tag | `"latest"` |
| `backend.image.pullPolicy` | Image pull policy | `"IfNotPresent"` |
| `backend.service.type` | Kubernetes Service type | `"ClusterIP"` |
| `backend.service.port` | Backend internal service port | `3001` |
| `backend.resources.requests.cpu` | CPU request | `"100m"` |
| `backend.resources.requests.memory` | Memory request | `"128Mi"` |
| `backend.resources.limits.cpu` | CPU limit | `"500m"` |
| `backend.resources.limits.memory` | Memory limit | `"512Mi"` |

### Frontend Service (`frontend`)
| Parameter | Description | Default |
|---|---|---|
| `frontend.replicaCount` | Number of frontend pod replicas | `1` |
| `frontend.image.repository` | Frontend Docker image repository | `"kubenexus-frontend"` |
| `frontend.image.tag` | Frontend image tag | `"latest"` |
| `frontend.image.pullPolicy` | Image pull policy | `"IfNotPresent"` |
| `frontend.service.type` | Kubernetes Service type | `"ClusterIP"` |
| `frontend.service.port` | Frontend internal service port | `80` |
| `frontend.resources.requests.cpu` | CPU request | `"50m"` |
| `frontend.resources.requests.memory` | Memory request | `"64Mi"` |
| `frontend.resources.limits.cpu` | CPU limit | `"200m"` |
| `frontend.resources.limits.memory` | Memory limit | `"256Mi"` |

### Ingress (`ingress`)
| Parameter | Description | Default |
|---|---|---|
| `ingress.enabled` | Enable Ingress resource | `true` |
| `ingress.className` | Ingress class name (e.g. `nginx`, `traefik`) | `"nginx"` |
| `ingress.annotations` | Ingress annotations (WebSocket, timeouts, proxy limits) | `{...}` |
| `ingress.hosts[0].host` | Hostname to route traffic | `"kubenexus.local"` |
| `ingress.tls` | TLS configuration secrets and hosts | `[]` |

### Database & Storage
| Parameter | Description | Default |
|---|---|---|
| `postgresql.enabled` | Deploy built-in lightweight PostgreSQL | `true` |
| `postgresql.image.repository` | PostgreSQL image | `"postgres"` |
| `postgresql.image.tag` | PostgreSQL image tag | `"17-alpine"` |
| `postgresql.persistence.enabled` | Enable PVC for database data | `true` |
| `postgresql.persistence.size` | Storage volume size | `"2Gi"` |
| `database.host` | External Postgres host (if `postgresql.enabled: false`) | `""` |
| `database.port` | PostgreSQL port | `5432` |
| `database.name` | Database name | `"kubenexus_db"` |
| `database.user` | Database user | `"postgres"` |
| `database.password` | Database password. If empty (`""`), auto-generated | `""` (Auto-generated) |
| `redisInternal.enabled` | Deploy built-in lightweight Redis | `true` |
| `redis.host` | External Redis host (if `redisInternal.enabled: false`) | `""` |
| `redis.port` | Redis port | `6379` |

### RBAC (`rbac`)
| Parameter | Description | Default |
|---|---|---|
| `rbac.create` | Create ServiceAccount for Backend client-go | `true` |
| `rbac.clusterRole.create` | Create ClusterRole & Binding for cluster management | `true` |

---

## Production Deployment Example with Custom Values

Create a file named `production-values.yaml`:

```yaml
ingress:
  className: nginx
  hosts:
    - host: kubenexus.mycompany.com
      paths:
        - path: /api
          pathType: Prefix
          service: backend
        - path: /
          pathType: Prefix
          service: frontend
  tls:
    - secretName: kubenexus-tls-cert
      hosts:
        - kubenexus.mycompany.com

backend:
  replicaCount: 2
  image:
    repository: ghcr.io/myorg/kubenexus-backend
    tag: v2.0.0

frontend:
  replicaCount: 2
  image:
    repository: ghcr.io/myorg/kubenexus-frontend
    tag: v2.0.0

migration:
  image:
    repository: ghcr.io/myorg/kubenexus-migrate
    tag: v2.0.0

# Using External Cloud RDS PostgreSQL
postgresql:
  enabled: false

database:
  host: "postgres-cluster.production.rds.amazonaws.com"
  port: 5432
  name: "kubenexus_prod"
  user: "kubenexus_admin"
  password: "MySecureExternalPassword!"
  sslmode: "require"
```

Deploy with your custom values:
```bash
helm upgrade --install kubenexus ./helm/kubenexus \
  --namespace kubenexus \
  --create-namespace \
  -f production-values.yaml
```
