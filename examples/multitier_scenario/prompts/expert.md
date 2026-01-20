Generate Kubernetes YAML for namespace `ecommerce`:

**namespace.yaml:**
- Namespace: ecommerce, labels app=ecommerce
- ResourceQuota: 20 pods, 10 CPU, 20Gi memory

**frontend.yaml:**
- Deployment: frontend, 3 replicas, image ecommerce/frontend:latest, port 8080, 200m/256Mi
- Service: frontend, LoadBalancer, port 80->8080
- Labels: app=ecommerce, tier=frontend

**backend.yaml:**
- Deployment: backend, 2 replicas, image ecommerce/backend:latest, port 8080, 300m/512Mi
- Service: backend, ClusterIP, port 8080->8080
- Labels: app=ecommerce, tier=backend
- Env from ConfigMap: DATABASE_HOST, DATABASE_PORT, DATABASE_NAME
- Env from Secret: DATABASE_USER, DATABASE_PASSWORD

**config.yaml:**
- ConfigMap: app-config, database.host=postgres, database.port=5432, database.name=ecommerce
- Secret: db-credentials, username=postgres, password=changeme

**network.yaml:**
- NetworkPolicy: backend-policy, allow ingress from tier=frontend to tier=backend
- HPA: frontend-hpa, target frontend, 3-10 replicas, 70% CPU
