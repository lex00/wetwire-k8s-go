Namespace: `ecommerce`. Create these files:

**expected/namespace.go:**
- Namespace: ecommerce, labels app=ecommerce, tier=production
- ResourceQuota: 20 pods, 10 CPU, 20Gi memory

**expected/frontend.go:**
- Deployment: frontend, 3 replicas, image ecommerce/frontend:latest:8080, 200m/256Mi
- Service: frontend, LoadBalancer, port 80->8080
- Labels: app=ecommerce, tier=frontend

**expected/backend.go:**
- Deployment: backend, 2 replicas, image ecommerce/backend:latest:8080, 300m/512Mi
- Service: backend, ClusterIP, port 8080->8080
- Labels: app=ecommerce, tier=backend
- Env from ConfigMap: DATABASE_HOST (database.host), DATABASE_PORT (database.port), DATABASE_NAME (database.name)
- Env from Secret: DATABASE_USER (username), DATABASE_PASSWORD (password)

**expected/config.go:**
- ConfigMap: app-config, database.host=postgres, database.port=5432, database.name=ecommerce
- Secret: db-credentials, username=postgres, password=changeme

**expected/network.go:**
- NetworkPolicy: backend-policy, allow ingress from tier=frontend to tier=backend
- HPA: frontend-hpa, target frontend deployment, 3-10 replicas, 70% CPU
