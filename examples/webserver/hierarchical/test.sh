#!/bin/bash

# Script de teste para demonstrar o cache hierárquico do GoCacheX

set -e

echo "🚀 Testando GoCacheX Cache Hierárquico (L1 + L2)"
echo "================================================"

# Verificar se o servidor está rodando
if ! curl -s http://localhost:8081/api/v1/products > /dev/null; then
    echo "❌ Servidor hierárquico não está rodando!"
    echo "   Execute primeiro: go run hierarchical_cache.go"
    exit 1
fi

echo "✅ Servidor hierárquico está rodando!"
echo ""

# 1. Listar produtos (primeira vez - database)
echo "1️⃣ Listar produtos (primeira vez - database):"
curl -s -i http://localhost:8081/api/v1/products | head -n 10
echo ""

# 2. Listar produtos (segunda vez - L1 cache hit)
echo "2️⃣ Listar produtos (segunda vez - L1 cache hit):"
curl -s -i http://localhost:8081/api/v1/products | head -n 10
echo ""

# 3. Buscar produto específico (primeira vez - database)
echo "3️⃣ Buscar produto 1 (primeira vez - database):"
curl -s -i http://localhost:8081/api/v1/products/1 | head -n 10
echo ""

# 4. Buscar produto específico (segunda vez - L1 cache hit)
echo "4️⃣ Buscar produto 1 (segunda vez - L1 cache hit):"
curl -s -i http://localhost:8081/api/v1/products/1 | head -n 10
echo ""

# 5. Estatísticas do cache
echo "5️⃣ Estatísticas do cache (L1 e L2):"
curl -s http://localhost:8081/api/v1/cache/stats | jq .
echo ""

# 6. Buscar diferentes produtos para popular o cache
echo "6️⃣ Populando cache com diferentes produtos:"
for i in {2..5}; do
    echo "   Buscando produto $i..."
    curl -s http://localhost:8081/api/v1/products/$i > /dev/null
done
echo ""

# 7. Estatísticas após popular o cache
echo "7️⃣ Estatísticas após popular o cache:"
curl -s http://localhost:8081/api/v1/cache/stats | jq .
echo ""

# 8. Atualizar produto (deve invalidar cache)
echo "8️⃣ Atualizando produto 1 (deve invalidar cache):"
curl -s -X PUT http://localhost:8081/api/v1/products/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop Updated","price":1099.99,"category":"Electronics","description":"Updated high-performance laptop"}' | jq .
echo ""

# 9. Buscar produto atualizado (deve ser database miss novamente)
echo "9️⃣ Buscar produto 1 atualizado (cache miss):"
curl -s -i http://localhost:8081/api/v1/products/1 | head -n 10
echo ""

# 10. Estatísticas finais
echo "🔟 Estatísticas finais:"
curl -s http://localhost:8081/api/v1/cache/stats | jq .
echo ""

echo "✅ Teste de cache hierárquico completo!"
echo ""
echo "📊 Observações sobre o cache hierárquico:"
echo "   • L1 (Memory): Cache rápido, menor capacidade"
echo "   • L2 (Redis): Cache persistente, maior capacidade"
echo "   • Ordem de busca: L1 → L2 → Database"
echo "   • Dados são promovidos do L2 para L1 automaticamente"
echo "   • Invalidação remove dados de ambos os níveis"
