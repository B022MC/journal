#!/usr/bin/env bash
# =============================================================
# deploy.sh — One-shot full deployment of Journal to K3s cluster
# Usage: SSH into the master node and run this script
# =============================================================
set -euo pipefail

echo "🚀 Deploying Journal to K3s..."

NAMESPACE="journal"
K8S_DIR="/tmp/journal-k8s"

# Create a k8s directory on the server if needed
mkdir -p "$K8S_DIR"

echo "📦 Step 1: Apply namespace and secrets..."
k3s kubectl apply -f "$K8S_DIR/namespace.yaml"
k3s kubectl apply -f "$K8S_DIR/secrets.yaml"

echo "🗄️ Step 2: Deploy infrastructure (MySQL, Redis, etcd, Jaeger)..."
k3s kubectl apply -f "$K8S_DIR/mysql.yaml"
k3s kubectl apply -f "$K8S_DIR/redis.yaml"
k3s kubectl apply -f "$K8S_DIR/etcd.yaml"
k3s kubectl apply -f "$K8S_DIR/jaeger.yaml"

echo "⏳ Waiting for MySQL to be ready..."
k3s kubectl -n "$NAMESPACE" wait --for=condition=ready pod -l app=mysql --timeout=180s

echo "🏗️ Step 3: Initialize MySQL schema..."
k3s kubectl apply -f "$K8S_DIR/mysql-init.yaml"
k3s kubectl -n "$NAMESPACE" wait --for=condition=complete job/mysql-init-schema --timeout=120s

echo "⏳ Waiting for etcd to be ready..."
k3s kubectl -n "$NAMESPACE" wait --for=condition=ready pod -l app=etcd --timeout=60s

echo "🔧 Step 4: Deploy backend services..."
k3s kubectl apply -f "$K8S_DIR/user-rpc.yaml"
k3s kubectl apply -f "$K8S_DIR/paper-rpc.yaml"
k3s kubectl apply -f "$K8S_DIR/rating-rpc.yaml"
k3s kubectl apply -f "$K8S_DIR/news-rpc.yaml"
k3s kubectl apply -f "$K8S_DIR/api-deployment.yaml"

echo "🌐 Step 5: Deploy frontend..."
k3s kubectl apply -f "$K8S_DIR/frontend.yaml"

echo "⏳ Waiting for all pods to be ready..."
k3s kubectl -n "$NAMESPACE" wait --for=condition=ready pod --all --timeout=300s

echo ""
echo "✅ Deployment complete!"
echo ""
k3s kubectl -n "$NAMESPACE" get pods
echo ""
k3s kubectl -n "$NAMESPACE" get svc
