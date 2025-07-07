# Basic Web Server Example

Este exemplo demonstra o uso básico do GoCacheX em uma aplicação web server simples.

## Como Executar

```bash
go run main.go
```

O servidor iniciará na porta 8080.

## Testar

Execute o script de teste:

```bash
./test.sh
```

## Funcionalidades

- Cache in-memory com LRU
- API REST completa (CRUD)
- Headers de cache (HIT/MISS)
- Estatísticas de cache
- Health checks
- Invalidação automática

## Endpoints

- `GET /api/v1/users` - Lista usuários
- `GET /api/v1/users/{id}` - Busca usuário
- `POST /api/v1/users` - Cria usuário
- `PUT /api/v1/users/{id}` - Atualiza usuário
- `DELETE /api/v1/users/{id}` - Remove usuário
- `GET /api/v1/cache/stats` - Estatísticas
- `POST /api/v1/cache/clear` - Limpa cache
- `GET /health` - Health check
