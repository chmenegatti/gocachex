# GoCacheX Web Server Examples - Estrutura Final

## ✅ Problema Resolvido

O erro de "main redeclared" foi **resolvido** organizando os exemplos em diretórios separados:

```text
webserver/
├── basic/           # Exemplo básico (porta 8080)
│   ├── main.go      # Web server com cache in-memory
│   ├── go.mod       # Dependências
│   ├── test.sh      # Script de teste
│   └── README.md    # Documentação
├── hierarchical/    # Cache hierárquico (porta 8081)
│   ├── main.go      # L1 (Memory) + L2 (Redis)
│   ├── go.mod       # Dependências
│   ├── test.sh      # Script de teste
│   └── README.md    # Documentação
├── multibackend/    # Múltiplos backends
│   ├── main.go      # Comparação de backends
│   ├── go.mod       # Dependências
│   └── README.md    # Documentação
├── README.md        # Documentação geral
└── FILES.md         # Este arquivo
```

## 🚀 Como Usar Agora

### Exemplo Básico

```bash
cd examples/webserver/basic
go run main.go
# Servidor na porta 8080
./test.sh  # Testar
```

### Exemplo Hierárquico

```bash
cd examples/webserver/hierarchical
go run main.go
# Servidor na porta 8081
./test.sh  # Testar
```

### Exemplo Multi-Backend

```bash
cd examples/webserver/multibackend
go run main.go
```

## ✅ Status dos Exemplos

- [x] **basic/** - ✅ Compilando e funcionando
- [x] **hierarchical/** - ✅ Compilando e funcionando
- [x] **multibackend/** - ✅ Compilando (em desenvolvimento)

## 🎯 Características por Exemplo

### Basic (Porta 8080)

- Cache in-memory com LRU
- API REST completa (usuários)
- Headers de cache (HIT/MISS)
- Estatísticas e health checks

### Hierarchical (Porta 8081)

- Cache L1 (Memory) + L2 (Redis)
- Promoção automática de dados
- Fallback gracioso sem Redis
- API REST (produtos)

### Multi-Backend

- Comparação de backends
- Diferentes configurações
- Análise de performance

## 🔧 Benefícios da Nova Estrutura

1. **Sem Conflitos**: Cada exemplo tem sua própria função main
2. **Modulares**: Dependências independentes por exemplo
3. **Testáveis**: Scripts de teste dedicados
4. **Documentados**: README individual por exemplo
5. **Escaláveis**: Fácil adicionar novos exemplos

Os exemplos estão agora **totalmente funcionais** e **livres de erros de compilação**!

## Estrutura dos Exemplos

### Web Server Básico (`main.go`)

```text
Features:
✅ Cache in-memory com LRU
✅ API REST completa (CRUD)
✅ Headers de cache (HIT/MISS)
✅ Estatísticas de cache
✅ Health checks
✅ Invalidação automática
✅ Compressão gzip
✅ Serialização JSON

Endpoints:
- GET /api/v1/users         - Lista usuários
- GET /api/v1/users/{id}    - Busca usuário
- POST /api/v1/users        - Cria usuário
- PUT /api/v1/users/{id}    - Atualiza usuário
- DELETE /api/v1/users/{id} - Remove usuário
- GET /api/v1/cache/stats   - Estatísticas
- POST /api/v1/cache/clear  - Limpa cache
- GET /health               - Health check
```

### Cache Hierárquico (`hierarchical_cache.go`)

```text
Features:
✅ Cache L1 (Memory) + L2 (Redis)
✅ Promoção automática L2→L1
✅ Invalidação em todos os níveis
✅ Fallback gracioso (Redis opcional)
✅ Diferentes TTLs por nível
✅ Estatísticas separadas L1/L2

Endpoints:
- GET /api/v1/products      - Lista produtos
- GET /api/v1/products/{id} - Busca produto
- PUT /api/v1/products/{id} - Atualiza produto
- GET /api/v1/cache/stats   - Estatísticas L1/L2
```

## Como Usar

### 1. Executar Exemplo Básico

```bash
go run main.go
# Servidor na porta 8080
./test.sh  # Executar testes
```

### 2. Executar Exemplo Hierárquico

```bash
# Opcional: Redis
docker run -d -p 6379:6379 redis:alpine

go run hierarchical_cache.go
# Servidor na porta 8081
./test_hierarchical.sh  # Executar testes
```

## Tecnologias Demonstradas

### GoCacheX Features

- [x] Multiple backends (Memory, Redis)
- [x] Hierarchical caching (L1/L2)
- [x] Compression (gzip)
- [x] Serialization (JSON)
- [x] Cache invalidation
- [x] Statistics & monitoring
- [x] Health checks
- [x] Atomic operations
- [x] Batch operations
- [x] TTL management

### Web Server Features

- [x] RESTful API design
- [x] JSON serialization
- [x] Error handling
- [x] HTTP headers
- [x] Request logging
- [x] URL routing
- [x] CRUD operations

## Benefícios Demonstrados

### Performance

- **Cache Hits**: Respostas instantâneas do cache
- **Hierarchical**: Máxima velocidade com L1, capacidade com L2
- **Compression**: Redução do uso de memória/rede

### Reliability

- **Health Checks**: Monitoramento da saúde do cache
- **Graceful Fallback**: Funciona mesmo sem Redis
- **Error Handling**: Tratamento robusto de erros

### Observability

- **Cache Headers**: Visibilidade de hits/misses
- **Statistics**: Métricas detalhadas de performance
- **Logging**: Rastreamento de operações

## Próximos Passos

Para usar em produção, considere:

- [ ] Adicionar autenticação/autorização
- [ ] Implementar rate limiting
- [ ] Configurar métricas Prometheus
- [ ] Adicionar tracing OpenTelemetry
- [ ] Usar connection pooling
- [ ] Implementar circuit breakers
- [ ] Adicionar testes unitários
- [ ] Configurar CI/CD

Estes exemplos servem como base sólida para aplicações web que precisam de cache de alta performance!
