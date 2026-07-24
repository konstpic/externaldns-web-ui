.PHONY: dev-backend dev-frontend build-backend build-frontend docker-build helm-template

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

build-backend:
	cd backend && go build -o bin/server ./cmd/server

build-frontend:
	cd frontend && npm run build

docker-build:
	docker build -t externaldns-web-ui/backend:latest ./backend
	docker build -t externaldns-web-ui/frontend:latest -f frontend/Dockerfile .

helm-template:
	helm template externaldns-web-ui ./.helm/externaldns-web-ui \
		--set clusterName=sharxconnect \
		--set domainFilter=sharxconnect.app \
		--set ingress.host=externaldns.sharxconnect.app
