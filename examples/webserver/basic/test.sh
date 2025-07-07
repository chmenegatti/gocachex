#!/bin/bash

# Script de teste para demonstrar o exemplo de web server do GoCacheX

set -e

echo "🚀 Testando GoCacheX Web Server Example"
echo "========================================"

# Verificar se o servidor está rodando
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Servidor não está rodando!"
    echo "   Execute primeiro: go run main.go"
    exit 1
fi

echo "✅ Servidor está rodando!"
echo ""

# 1. Health check
echo "1️⃣ Health Check:"
curl -s http://localhost:8080/health | jq .
echo ""

# 2. Listar usuários (primeira vez - cache miss)
echo "2️⃣ Listar usuários (primeira vez - cache miss):"
curl -s -i http://localhost:8080/api/v1/users | head -n 10
echo ""

# 3. Listar usuários (segunda vez - cache hit)
echo "3️⃣ Listar usuários (segunda vez - cache hit):"
curl -s -i http://localhost:8080/api/v1/users | head -n 10
echo ""

# 4. Buscar usuário específico (primeira vez - cache miss)
echo "4️⃣ Buscar usuário 1 (primeira vez - cache miss):"
curl -s -i http://localhost:8080/api/v1/users/1 | head -n 10
echo ""

# 5. Buscar usuário específico (segunda vez - cache hit)
echo "5️⃣ Buscar usuário 1 (segunda vez - cache hit):"
curl -s -i http://localhost:8080/api/v1/users/1 | head -n 10
echo ""

# 6. Estatísticas do cache
echo "6️⃣ Estatísticas do cache:"
curl -s http://localhost:8080/api/v1/cache/stats | jq .
echo ""

# 7. Criar novo usuário
echo "7️⃣ Criando novo usuário:"
NEW_USER=$(curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"David","email":"david@example.com"}')
echo $NEW_USER | jq .
NEW_USER_ID=$(echo $NEW_USER | jq -r .id)
echo ""

# 8. Buscar novo usuário criado
echo "8️⃣ Buscar novo usuário criado (ID: $NEW_USER_ID):"
curl -s http://localhost:8080/api/v1/users/$NEW_USER_ID | jq .
echo ""

# 9. Atualizar usuário
echo "9️⃣ Atualizando usuário $NEW_USER_ID:"
curl -s -X PUT http://localhost:8080/api/v1/users/$NEW_USER_ID \
  -H "Content-Type: application/json" \
  -d '{"name":"David Updated","email":"david.updated@example.com"}' | jq .
echo ""

# 10. Estatísticas finais do cache
echo "🔟 Estatísticas finais do cache:"
curl -s http://localhost:8080/api/v1/cache/stats | jq .
echo ""

# 11. Limpar cache
echo "1️⃣1️⃣ Limpando cache:"
curl -s -X POST http://localhost:8080/api/v1/cache/clear
echo "Cache limpo!"
echo ""

# 12. Verificar cache após limpeza
echo "1️⃣2️⃣ Estatísticas após limpeza:"
curl -s http://localhost:8080/api/v1/cache/stats | jq .
echo ""

# 13. Buscar usuário após limpeza (deve ser cache miss)
echo "1️⃣3️⃣ Buscar usuário 1 após limpeza (cache miss):"
curl -s -i http://localhost:8080/api/v1/users/1 | head -n 10
echo ""

echo "✅ Teste completo! Verifique os logs do servidor para ver os HITs e MISSes."
