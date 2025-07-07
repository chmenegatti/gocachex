# Hierarchical Cache Example

Este exemplo demonstra cache hierárquico usando L1 (Memory) + L2 (Redis).

## Pré-requisitos

Redis (opcional):

```bash
docker run -d -p 6379:6379 redis:alpine
```

## Como Executar

```bash
go run main.go
```

O servidor iniciará na porta 8081.

## Testar

Execute o script de teste:

```bash
./test.sh
```

## Funcionalidades

- Cache L1 (Memory) - rápido, limitado
- Cache L2 (Redis) - distribuído, maior capacidade
- Promoção automática L2 → L1
- Invalidação coordenada
- Fallback gracioso (funciona sem Redis)
- TTLs diferentes por nível

## Endpoints

- `GET /api/v1/products` - Lista produtos
- `GET /api/v1/products/{id}` - Busca produto
- `PUT /api/v1/products/{id}` - Atualiza produto
- `GET /api/v1/cache/stats` - Estatísticas L1/L2

## Como Funciona

1. **Busca**: L1 → L2 → Database
2. **Cache Hit L1**: Resposta instantânea
3. **Cache Hit L2**: Dados promovidos para L1
4. **Cache Miss**: Dados salvos em L1 e L2
5. **Invalidação**: Remove de ambos os níveis
