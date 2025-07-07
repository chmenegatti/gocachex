# GoCacheX Web Server Examples

Este diretório contém exemplos práticos de como usar o GoCacheX em aplicações web server reais.

## Estrutura dos Exemplos

```
webserver/
├── basic/           # Exemplo básico com cache in-memory
│   ├── main.go
│   ├── go.mod
│   └── test.sh
├── hierarchical/    # Cache hierárquico (L1 + L2)
│   ├── main.go
│   ├── go.mod
│   └── test.sh
└── multibackend/   # Exemplo com múltiplos backends
    ├── main.go
    └── go.mod
```

## Exemplos Disponíveis

### 1. `basic/` - Web Server Básico
Demonstra uso básico do GoCacheX com cache in-memory em uma API REST simples.

### 2. `hierarchical/` - Cache Hierárquico
Demonstra cache hierárquico (L1/L2) usando Memory + Redis para máxima performance.

### 3. `multibackend/` - Múltiplos Backends
Demonstra diferentes estratégias de backend de cache.

## Características Demonstradas

- **Cache de Entidades**: Cache individual de usuários/produtos por ID
- **Cache de Listas**: Cache de lista completa de itens
- **Cache Hierárquico**: L1 (Memory) + L2 (Redis) para otimização
- **Invalidação de Cache**: Limpeza automática do cache quando dados são modificados
- **Cache Headers**: Headers HTTP indicando HIT/MISS do cache
- **Estatísticas de Cache**: Endpoint para visualizar métricas do cache
- **Health Checks**: Verificação de saúde do sistema de cache

## Como Executar

### Exemplo Básico (basic/)

1. Navegue até o diretório do exemplo:

```bash
cd examples/webserver/basic
```

2. Execute o servidor:

```bash
go run main.go
```

3. O servidor iniciará na porta 8080.

### Exemplo Hierárquico (hierarchical/)

1. Certifique-se de que o Redis está rodando (opcional):

```bash
docker run -d -p 6379:6379 redis:alpine
```

2. Execute o servidor hierárquico:

```bash
cd examples/webserver/hierarchical
go run main.go
```

3. O servidor iniciará na porta 8081.

## Endpoints da API

### Usuários

- `GET /api/v1/users` - Lista todos os usuários (com cache)
- `GET /api/v1/users/{id}` - Busca usuário por ID (com cache)
- `POST /api/v1/users` - Cria novo usuário
- `PUT /api/v1/users/{id}` - Atualiza usuário existente
- `DELETE /api/v1/users/{id}` - Remove usuário

### Cache Management

- `GET /api/v1/cache/stats` - Estatísticas do cache
- `POST /api/v1/cache/clear` - Limpa todo o cache

### Saúde

- `GET /health` - Health check do sistema

## Exemplos de Uso

### 1. Verificar saúde do sistema

```bash
curl http://localhost:8080/health
```

### 2. Listar todos os usuários

```bash
curl http://localhost:8080/api/v1/users
```

### 3. Buscar usuário específico

```bash
curl http://localhost:8080/api/v1/users/1
```

### 4. Criar novo usuário

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"David","email":"david@example.com"}'
```

### 5. Atualizar usuário

```bash
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'
```

### 6. Remover usuário

```bash
curl -X DELETE http://localhost:8080/api/v1/users/1
```

### 7. Ver estatísticas do cache

```bash
curl http://localhost:8080/api/v1/cache/stats
```

### 8. Limpar cache

```bash
curl -X POST http://localhost:8080/api/v1/cache/clear
```

## Observando o Comportamento do Cache

1. Faça uma requisição para buscar um usuário:

```bash
curl -i http://localhost:8080/api/v1/users/1
```

Você verá `X-Cache: MISS` no header.

2. Faça a mesma requisição novamente:

```bash
curl -i http://localhost:8080/api/v1/users/1
```

Agora você verá `X-Cache: HIT` no header.

3. Verifique os logs do servidor para ver as mensagens de HIT/MISS.

## Configuração do Cache

O exemplo usa cache in-memory por padrão, mas você pode facilmente mudar para Redis descomentando as linhas correspondentes no código:

```go
// Para usar Redis, descomente:
// Backend: "redis",
// Redis: config.RedisConfig{
//     Addresses: []string{"localhost:6379"},
//     Password:  "",
//     DB:        0,
//     PoolSize:  10,
// },
```

## Estratégias de Cache Demonstradas

### 1. Cache por Entidade

- Cada usuário é cacheado individualmente com chave `user:{id}`
- TTL de 5 minutos para usuários individuais

### 2. Cache de Lista

- Lista completa de usuários é cacheada com chave `users:all`
- TTL de 2 minutos (mais curto que entidades individuais)

### 3. Invalidação Inteligente

- Ao criar/atualizar/deletar usuário, o cache individual é atualizado
- O cache da lista é invalidado para manter consistência

### 4. Compressão

- Dados são comprimidos automaticamente usando gzip
- Serialização JSON para compatibilidade

## Métricas e Monitoramento

O endpoint `/api/v1/cache/stats` retorna estatísticas detalhadas:

- Número de hits/misses
- Total de operações (sets, deletes)
- Contagem de chaves
- Uso de memória
- Taxa de hit ratio

## Logs

O servidor registra todas as operações de cache:

- `Cache HIT for user X` - Dados encontrados no cache
- `Cache MISS for user X` - Dados não encontrados, buscados da fonte
- Tempo de resposta de cada requisição HTTP

Este exemplo serve como base para implementar caching em aplicações web reais usando o GoCacheX.
