Create Kubernetes resources for a multi-tier e-commerce application:

**Namespace:**
- Name: ecommerce
- ResourceQuota: 20 pods, 10 CPU, 20Gi memory

**Frontend:**
- 3 replicas, LoadBalancer service on port 80
- Image: ecommerce/frontend:latest, port 8080
- Resources: 200m CPU, 256Mi memory
- HPA: scale 3-10 pods based on 70% CPU

**Backend:**
- 2 replicas, ClusterIP service on port 8080
- Image: ecommerce/backend:latest, port 8080
- Resources: 300m CPU, 512Mi memory
- Environment: DATABASE_HOST, DATABASE_PORT, DATABASE_NAME from ConfigMap
- Environment: DATABASE_USER, DATABASE_PASSWORD from Secret

**ConfigMap:**
- database.host: postgres
- database.port: 5432
- database.name: ecommerce

**Secret:**
- username: postgres
- password: changeme

**NetworkPolicy:**
- Backend only accepts traffic from frontend
