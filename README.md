# ExternalDNS Web UI

Современный web-интерфейс для [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) — просмотр DNS записей, источников (Service/Ingress/DNSEndpoint), статуса controller и логов.

## Возможности

- **Dashboard** — сводка по записям, источникам и статусу ExternalDNS
- **DNS записи** — таблица с поиском по hostname, namespace, target
- **Источники** — Services, Ingresses и DNSEndpoint CRD с annotations
- **Controller** — provider, policy, txt-owner-id, replicas, image
- **Логи** — tail логов pod ExternalDNS

## Аутентификация (Authentik OIDC)

Все API endpoints `/api/v1/*` защищены JWT. Вход через Authentik OIDC:

1. Пользователь нажимает «Войти через Authentik»
2. Backend редиректит на IdP → callback `/api/auth/callback`
3. Backend выдаёт JWT и редиректит на `/auth/callback?access_token=...`
4. SPA сохраняет токены и загружает профиль через `/api/auth/me`

### Переменные окружения (backend)

| Variable | Описание |
|----------|----------|
| `JWT_SECRET` | Секрет для подписи JWT (**обязателен**) |
| `OIDC_ISSUER_URL` | Authentik issuer, напр. `https://auth.example.com/application/o/externaldns-web-ui/` |
| `OIDC_CLIENT_ID` | OAuth2 client ID |
| `OIDC_CLIENT_SECRET` | OAuth2 client secret |
| `FRONTEND_URL` | Public URL UI, напр. `https://externaldns.example.com` |
| `AUTHENTIK_PUBLIC_URL` | Публичный URL Authentik (если issuer внутренний) |
| `ADMIN_ROLES` | Группы/роли с доступом к `/admin/settings` |
| `AUTH_REQUIRED` | `false` только для локальной разработки |

### Локальная разработка без OIDC

```bash
AUTH_REQUIRED=false JWT_SECRET=dev-secret go run ./cmd/server
```

## Архитектура

```
Browser → nginx (frontend) → /api/* → Go backend → Kubernetes API
                              /*    → React SPA
```

Backend использует in-cluster ServiceAccount (или kubeconfig локально) для чтения Services, Ingresses, DNSEndpoint CRD и статуса deployment ExternalDNS.

## Локальная разработка

### Backend

```bash
cd backend
go run ./cmd/server
# LISTEN_ADDR=:8080 KUBECONFIG=~/.kube/config
```

### Frontend

```bash
cd frontend
npm install
npm run dev
# http://localhost:5173 — прокси /api → :8080
```

### Docker Compose

```bash
docker compose up --build
# UI: http://localhost:8088
```

## Сборка образов

```bash
docker build -t externaldns-web-ui/backend:latest ./backend
docker build -t externaldns-web-ui/frontend:latest -f frontend/Dockerfile .
```

## Helm (Kubernetes)

```bash
helm upgrade --install externaldns-web-ui ./.helm/externaldns-web-ui \
  --namespace externaldns-web-ui --create-namespace \
  --set clusterName=sharxconnect \
  --set domainFilter=sharxconnect.app \
  --set ingress.host=externaldns.sharxconnect.app \
  --set global.imageRegistry=harbor.sharxconnect.app \
  --set backend.image.tag=latest \
  --set frontend.image.tag=latest
```

### Values

| Key | Описание |
|-----|----------|
| `clusterName` | Имя кластера для UI |
| `domainFilter` | Domain filter ExternalDNS |
| `externalDns.namespace` | Namespace ExternalDNS (default: `external-dns`) |
| `ingress.host` | Hostname Ingress |
| `global.imageRegistry` | Harbor/registry prefix |

## ArgoCD (konstpic-k8s)

Манифест Application добавлен в `deployments/apps/apps/externaldns-web-ui.yaml`.

После push образов в Harbor и настройки `EXTERNALDNS_WEB_UI_REPO` в `config.env`:

```bash
./scripts/render.sh
kubectl apply -f argocd/apps/root-app.yaml
```

ExternalDNS автоматически создаст DNS запись для Ingress по annotation `external-dns.alpha.kubernetes.io/hostname`.

## Admin (только для admin-ролей)

- **Обзор** — статистика, ресурсы без DNS annotations
- **Управление DNS** — create / **edit** / delete annotations и DNSEndpoint CRD
- **Audit log** — журнал всех admin-операций
- **Настройки** — OIDC и конфигурация приложения

Admin API: `/api/v1/admin/*`

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | `/annotate` | Создать annotation |
| PUT | `/annotate` | Обновить hostname/TTL/internal |
| GET | `/annotate?kind&namespace&name` | Текущие значения |
| DELETE | `/annotate` | Удалить annotations |
| POST | `/dnsendpoints` | Создать CRD |
| PUT | `/dnsendpoints/{ns}/{name}` | Обновить CRD |
| GET | `/dnsendpoints/{ns}/{name}` | Прочитать CRD |
| DELETE | `/dnsendpoints/{ns}/{name}` | Удалить CRD |


Chart создаёт ClusterRole с read-only доступом:

- `services`, `ingresses` (cluster-wide)
- `dnsendpoints.externaldns.k8s.io`
- `deployments`, `pods`, `pods/log` (для ExternalDNS namespace через controller status)

## License

Apache-2.0
