.PHONY: help build deploy delete status logs clean validate lint

help: ## Show this help message
	@echo "E-Commerce Kubernetes Setup"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Kubernetes manifests from Go code
	wetwire-k8s build -o manifests.yaml
	@echo "✓ Manifests generated: manifests.yaml"

validate: ## Validate the Go code and Kubernetes schemas
	wetwire-k8s validate
	@echo "✓ Validation passed"

lint: ## Check code for issues
	wetwire-k8s lint
	@echo "✓ Lint checks passed"

lint-fix: ## Auto-fix code issues
	wetwire-k8s lint --fix
	@echo "✓ Code auto-fixed"

deploy: build ## Build and deploy to Kubernetes
	kubectl apply -f manifests.yaml
	@echo "✓ Deployed to Kubernetes"
	@echo ""
	@echo "Check status with: make status"

delete: ## Delete all resources from Kubernetes
	kubectl delete -f manifests.yaml
	@echo "✓ Deleted from Kubernetes"

status: ## Show status of all resources
	@echo "=== Namespace ==="
	kubectl get namespace ecommerce
	@echo ""
	@echo "=== Deployments ==="
	kubectl get deployments -n ecommerce
	@echo ""
	@echo "=== StatefulSets ==="
	kubectl get statefulsets -n ecommerce
	@echo ""
	@echo "=== Pods ==="
	kubectl get pods -n ecommerce
	@echo ""
	@echo "=== Services ==="
	kubectl get services -n ecommerce
	@echo ""
	@echo "=== HPA ==="
	kubectl get hpa -n ecommerce
	@echo ""
	@echo "=== PVC ==="
	kubectl get pvc -n ecommerce

logs-frontend: ## Show frontend logs
	kubectl logs -n ecommerce -l tier=frontend --tail=50 -f

logs-backend: ## Show backend logs
	kubectl logs -n ecommerce -l tier=backend --tail=50 -f

logs-db: ## Show database logs
	kubectl logs -n ecommerce -l tier=database --tail=50 -f

url: ## Get the frontend URL
	@echo "Frontend LoadBalancer:"
	@kubectl get service frontend-service -n ecommerce -o jsonpath='{.status.loadBalancer.ingress[0].ip}' && echo "" || echo "Waiting for LoadBalancer IP..."
	@echo ""
	@echo "Run 'kubectl get service frontend-service -n ecommerce -w' to watch for the IP"

clean: ## Remove generated files
	rm -f manifests.yaml
	@echo "✓ Cleaned generated files"

graph: ## Generate dependency graph
	wetwire-k8s graph -o graph.md
	@echo "✓ Graph generated: graph.md"

shell-frontend: ## Open shell in frontend pod
	kubectl exec -it -n ecommerce $$(kubectl get pod -n ecommerce -l tier=frontend -o jsonpath='{.items[0].metadata.name}') -- sh

shell-backend: ## Open shell in backend pod
	kubectl exec -it -n ecommerce $$(kubectl get pod -n ecommerce -l tier=backend -o jsonpath='{.items[0].metadata.name}') -- sh

shell-db: ## Open psql shell in database
	kubectl exec -it -n ecommerce postgres-0 -- psql -U ecommerce_user -d ecommerce_db
