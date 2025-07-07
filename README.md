# GoCacheX 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/chmenegatti/gocachex)](https://goreportcard.com/report/github.com/chmenegatti/gocachex)
[![GoDoc](https://godoc.org/github.com/chmenegatti/gocachex?status.svg)](https://godoc.org/github.com/chmenegatti/gocachex)
[![GitHub release](https://img.shields.io/github/release/chmenegatti/gocachex.svg)](https://github.com/chmenegatti/gocachex/releases)
[![GitHub stars](https://img.shields.io/github/stars/chmenegatti/gocachex.svg)](https://github.com/chmenegatti/gocachex/stargazers)
[![Coverage Status](https://coveralls.io/repos/github/chmenegatti/gocachex/badge.svg?branch=main)](https://coveralls.io/github/chmenegatti/gocachex?branch=main)

**GoCacheX** é uma biblioteca de cache distribuído plug-and-play em Go, projetada para ser simples, eficiente e altamente escalável. Suporta múltiplos backends, recursos avançados de cache e monitoramento completo.

## ✨ Características

### 🔌 Múltiplos Backends

- **Redis**: Suporte completo com clustering e persistência
- **Memcached**: Implementação otimizada para alta performance
- **In-Memory**: Cache local de alta velocidade
- **Plugin System**: Arquitetura extensível para backends customizados

### 🚀 Recursos Avançados

- **Sharding Automático**: Distribuição inteligente de dados entre nós
- **Políticas de Invalidação**: TTL, LRU, LFU configuráveis
- **Cache Hierárquico**: Suporte a L1/L2 cache
- **Compressão**: Compressão opcional de dados com múltiplos algoritmos
- **Serialização**: Serialização/deserialização automática de objetos Go

### 🔄 Sincronização Distribuída

- **gRPC Integration**: Comunicação eficiente entre nós
- **Invalidação Distribuída**: Protocolo de invalidação consistente
- **Operações Atômicas**: Suporte a operações distribuídas ACID

### 📊 Monitoramento & Observabilidade

- **Métricas Prometheus**: Exportação nativa de métricas
- **OpenTelemetry**: Rastreamento distribuído completo
- **Logging Estruturado**: Logs estruturados com múltiplos níveis
- **Health Checks**: Verificações de saúde dos backends

## 🚀 Instalação

```bash
go get github.com/chmenegatti/gocachex
```

## 💡 Uso Rápido

### Cache Básico

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/chmenegatti/gocachex"
)

func main() {
    // Criar cache in-memory
    cache := gocachex.New(gocachex.Config{
        Backend: "memory",
    })

    ctx := context.Background()

    // Set
    cache.Set(ctx, "key", "value", time.Minute)

    // Get
    value, err := cache.Get(ctx, "key")
    if err != nil {
        panic(err)
    }
    fmt.Println(value) // Output: value
}
```

### Cache Redis

```go
cache := gocachex.New(gocachex.Config{
    Backend: "redis",
    Redis: gocachex.RedisConfig{
        Addresses: []string{"localhost:6379"},
        Password:  "",
        DB:        0,
    },
})
```

### Cache Distribuído com gRPC

```go
cache := gocachex.New(gocachex.Config{
    Backend:     "redis",
    Distributed: true,
    GRPC: gocachex.GRPCConfig{
        Port:  50051,
        Peers: []string{"node1:50051", "node2:50051"},
    },
    Redis: gocachex.RedisConfig{
        Addresses: []string{"localhost:6379"},
    },
})
```

### Cache Hierárquico (L1/L2)

```go
cache := gocachex.New(gocachex.Config{
    Hierarchical: true,
    L1: gocachex.CacheConfig{
        Backend: "memory",
        Size:    "100MB",
    },
    L2: gocachex.CacheConfig{
        Backend: "redis",
        Redis: gocachex.RedisConfig{
            Addresses: []string{"localhost:6379"},
        },
    },
})
```

## 📖 Documentação

### Interface Principal

```go
type Cache interface {
    // Operações básicas
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)

    // Operações em lote
    GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
    SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    DeleteMulti(ctx context.Context, keys []string) error

    // Operações atômicas
    Increment(ctx context.Context, key string, delta int64) (int64, error)
    Decrement(ctx context.Context, key string, delta int64) (int64, error)

    // Gerenciamento
    Clear(ctx context.Context) error
    Stats(ctx context.Context) (*Stats, error)
    Health(ctx context.Context) error
}
```

### Configuração

```go
type Config struct {
    Backend      string        `json:"backend"`       // "memory", "redis", "memcached"
    Compression  bool          `json:"compression"`   // Habilitar compressão
    Serializer   string        `json:"serializer"`    // "json", "gob", "msgpack"
    Distributed  bool          `json:"distributed"`   // Cache distribuído
    Hierarchical bool          `json:"hierarchical"`  // Cache hierárquico
    
    // Configurações específicas
    Memory    MemoryConfig    `json:"memory,omitempty"`
    Redis     RedisConfig     `json:"redis,omitempty"`
    Memcached MemcachedConfig `json:"memcached,omitempty"`
    GRPC      GRPCConfig      `json:"grpc,omitempty"`
    
    // Cache hierárquico
    L1 CacheConfig `json:"l1,omitempty"`
    L2 CacheConfig `json:"l2,omitempty"`
    
    // Monitoramento
    Prometheus PrometheusConfig `json:"prometheus,omitempty"`
    Tracing    TracingConfig    `json:"tracing,omitempty"`
}
```

## 🏗️ Arquitetura

```bash
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │   Application   │    │   Application   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                        GoCacheX                                 │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Memory    │  │    Redis    │  │      Memcached          │  │
│  │   Backend   │  │   Backend   │  │       Backend           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Compression │  │ Serializer  │  │    Sharding             │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Monitoring  │  │   Tracing   │  │      Metrics            │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 📊 Métricas

GoCacheX exporta métricas detalhadas para Prometheus:

- `gocachex_operations_total`: Total de operações por tipo
- `gocachex_operation_duration_seconds`: Duração das operações
- `gocachex_cache_hits_total`: Total de cache hits
- `gocachex_cache_misses_total`: Total de cache misses
- `gocachex_cache_size_bytes`: Tamanho do cache em bytes
- `gocachex_active_connections`: Conexões ativas por backend

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Testes com cobertura
make test-coverage

# Testes de integração
make test-integration

# Benchmarks
make benchmark
```

## 🔧 Desenvolvimento

```bash
# Clonar o repositório
git clone https://github.com/chmenegatti/gocachex.git
cd gocachex

# Instalar dependências
go mod download

# Executar testes
make test

# Executar linting
make lint

# Build
make build
```

## 📚 Exemplos

Confira os exemplos completos no diretório [`examples/`](./examples/):

- [Cache Básico](./examples/basic/)
- [Cache Redis](./examples/redis/)
- [Cache Distribuído](./examples/distributed/)
- [Cache Hierárquico](./examples/hierarchical/)
- [Web Server](./examples/webserver/)
- [Microserviços](./examples/microservices/)
- [CLI Tool](./examples/cli/)

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor, leia o [CONTRIBUTING.md](./CONTRIBUTING.md) para detalhes sobre nosso processo de desenvolvimento.

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Apache License 2.0 - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Agradecimentos

- [Go Redis](https://github.com/redis/go-redis) - Cliente Redis para Go
- [GoMemcache](https://github.com/bradfitz/gomemcache) - Cliente Memcached
- [Prometheus](https://prometheus.io/) - Monitoramento e alertas
- [OpenTelemetry](https://opentelemetry.io/) - Observabilidade

## 🔗 Links Úteis

- [Documentação](https://godoc.org/github.com/chmenegatti/gocachex)
- [Exemplos](./examples/)
- [Changelog](./CHANGELOG.md)
- [Issues](https://github.com/chmenegatti/gocachex/issues)
- [Discussions](https://github.com/chmenegatti/gocachex/discussions)

---

### Feito com ❤️ em Go
