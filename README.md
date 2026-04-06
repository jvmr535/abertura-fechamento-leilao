# Leilão com Fechamento Automático

Sistema de leilões em Go com fechamento automático via Goroutines. Quando um leilão é criado, uma goroutine é disparada para monitorar o tempo e fechar o leilão automaticamente após a duração configurada.

## Como Funciona

Ao criar um leilão via `POST /auction`, o sistema:
1. Insere o leilão no MongoDB com status `Active`
2. Dispara uma **goroutine** em background que aguarda o tempo configurado em `AUCTION_DURATION`
3. Após o tempo expirar, a goroutine atualiza o status do leilão para `Completed` no banco de dados

A implementação está concentrada em `internal/infra/database/auction/create_auction.go`.

## Variáveis de Ambiente

Configuráveis no arquivo `cmd/auction/.env`:

| Variável | Descrição | Padrão |
|---|---|---|
| `AUCTION_DURATION` | Duração do leilão (ex: `5m`, `1h`, `30s`) | `5m` |
| `AUCTION_INTERVAL` | Intervalo para verificação de leilão no bid | `20s` |
| `BATCH_INSERT_INTERVAL` | Intervalo de inserção em lote de bids | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de bids | `4` |
| `MONGODB_URL` | URL de conexão ao MongoDB | - |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

## Rodar o Projeto

### Com Docker Compose

```bash
docker-compose up --build
```

O serviço estará disponível em `http://localhost:8080`.

### Endpoints

- `POST /auction` — Criar leilão
- `GET /auction` — Listar leilões
- `GET /auction/:auctionId` — Buscar leilão por ID
- `GET /auction/winner/:auctionId` — Buscar lance vencedor
- `POST /bid` — Criar lance
- `GET /bid/:auctionId` — Listar lances de um leilão
- `GET /user/:userId` — Buscar usuário

### Exemplo de Criação de Leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics",
    "description": "Brand new iPhone 15 Pro Max 256GB",
    "condition": 1
  }'
```

## Rodar os Testes

Os testes de integração requerem um MongoDB rodando. Com o docker-compose ativo:

```bash
# Subir apenas o MongoDB
docker-compose up -d mongodb

# Rodar o teste de fechamento automático
MONGODB_URL="mongodb://admin:admin@localhost:27017/auctions?authSource=admin" \
  go test ./internal/infra/database/auction/ -v -run TestAuctionAutoClose -timeout 30s
```

O teste cria um leilão com `AUCTION_DURATION=3s`, verifica que o status inicial é `Active`, aguarda 4 segundos e confirma que o status mudou para `Completed` automaticamente.
