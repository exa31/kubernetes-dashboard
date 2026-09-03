<#
.SYNOPSIS
    KubeNexus 1-Click Kubernetes Deployment Automation Script

.DESCRIPTION
    Simplifies building Docker images and deploying KubeNexus to Kubernetes using Helm.

.PARAMETER Action
    build   - Build Docker containers for backend, frontend, and migration
    deploy  - Install or upgrade KubeNexus via Helm
    status  - Show deployed pods, services, and ingress status
    logs    - Tail logs from backend pod
    all     - Execute build, deploy, and status in one command (default)

.PARAMETER Namespace
    Target Kubernetes namespace (default: kubenexus)

.EXAMPLE
    .\deploy.ps1 -Action all
    .\deploy.ps1 -Action deploy
    .\deploy.ps1 -Action status
#>

param (
    [ValidateSet("all", "build", "deploy", "status", "logs")]
    [string]$Action = "all",

    [string]$Namespace = "kubenexus",
    [string]$ReleaseName = "kubenexus"
)

$ErrorActionPreference = "Stop"

function Write-Header {
    param([string]$Text)
    Write-Host "`n========================================================" -ForegroundColor Cyan
    Write-Host "  $Text" -ForegroundColor Cyan
    Write-Host "========================================================`n" -ForegroundColor Cyan
}

function Invoke-HelmCommand {
    param([string[]]$HelmArgs)

    # Check if helm CLI exists locally
    $helmInstalled = Get-Command helm -ErrorAction SilentlyContinue
    if ($helmInstalled) {
        & helm @HelmArgs
    } else {
        Write-Host "Local helm CLI not detected. Using containerized Helm (alpine/helm)..." -ForegroundColor Yellow
        $currentDir = (Get-Location).Path -replace '\\', '/'
        docker run --rm `
            -v "${currentDir}/helm/kubenexus:/apps" `
            -v "$env:USERPROFILE/.kube:/root/.kube" `
            alpine/helm:3.17.0 @HelmArgs
    }
}

switch ($Action) {
    "build" {
        Write-Header "BUILDING DOCKER IMAGES (MULTI-STAGE)"

        Write-Host "1/3 Building Backend (Go 1.25 Alpine)..." -ForegroundColor Green
        docker build -t kubenexus-backend:latest ./be

        Write-Host "2/3 Building Frontend (Node 22 + Nginx)..." -ForegroundColor Green
        docker build -t kubenexus-frontend:latest ./fe

        Write-Host "3/3 Building Database Migration Tool..." -ForegroundColor Green
        docker build -t kubenexus-migrate:latest -f ./be/Dockerfile.migrate ./be

        Write-Host "`n[SUCCESS] All Docker images built successfully!" -ForegroundColor Green
    }

    "deploy" {
        Write-Header "DEPLOYING TO KUBERNETES VIA HELM"

        Write-Host "Ensuring namespace '$Namespace' exists..." -ForegroundColor Gray
        kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -

        Write-Host "Running Helm upgrade --install..." -ForegroundColor Green
        $chartPath = "./helm/kubenexus"
        Invoke-HelmCommand @("upgrade", "--install", $ReleaseName, $chartPath, "--namespace", $Namespace, "--create-namespace")

        Write-Host "`n[SUCCESS] Helm release '$ReleaseName' applied successfully!" -ForegroundColor Green
    }

    "status" {
        Write-Header "KUBENEXUS CLUSTER STATUS ($Namespace)"
        kubectl get all,ingress,pvc,secrets -n $Namespace
    }

    "logs" {
        Write-Header "BACKEND LOGS STREAM"
        kubectl logs -n $Namespace -l app.kubernetes.io/component=backend -f --tail=100
    }

    "all" {
        Write-Header "STARTING FULL 1-CLICK KUBENEXUS DEPLOYMENT"

        # Step 1: Build
        Write-Host "[1/3] Building images..." -ForegroundColor Yellow
        docker build -t kubenexus-backend:latest ./be
        docker build -t kubenexus-frontend:latest ./fe
        docker build -t kubenexus-migrate:latest -f ./be/Dockerfile.migrate ./be

        # Step 2: Deploy
        Write-Host "`n[2/3] Deploying via Helm..." -ForegroundColor Yellow
        kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -
        Invoke-HelmCommand @("upgrade", "--install", $ReleaseName, "./helm/kubenexus", "--namespace", $Namespace)

        # Step 3: Status
        Write-Host "`n[3/3] Checking cluster status..." -ForegroundColor Yellow
        Start-Sleep -Seconds 3
        kubectl get all,ingress -n $Namespace

        Write-Host "`n========================================================" -ForegroundColor Green
        Write-Host "  DEPLOIMENT SELESAI!" -ForegroundColor Green
        Write-Host "  Untuk mengakses dashboard melalui Ingress:" -ForegroundColor White
        Write-Host "  1. Pastikan ingress controller aktif di cluster Anda (misal: ingress-nginx)" -ForegroundColor Gray
        Write-Host "  2. Tambahkan domain ke /etc/hosts atau C:\Windows\System32\drivers\etc\hosts:" -ForegroundColor Gray
        Write-Host "     <INGRESS_IP_ATAU_127.0.0.1> kubenexus.local" -ForegroundColor Cyan
        Write-Host "  3. Buka browser: http://kubenexus.local" -ForegroundColor Cyan
        Write-Host "========================================================`n" -ForegroundColor Green
    }
}
