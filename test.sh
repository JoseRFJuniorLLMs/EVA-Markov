#!/bin/bash

# Script para executar EVA-Markov localmente em modo de teste

set -e

echo "🧪 Executando EVA-Markov em modo de teste..."

# Verificar se .env existe
if [ ! -f .env ]; then
    echo "❌ Arquivo .env não encontrado!"
    echo "📝 Copie .env.example para .env e configure as variáveis"
    exit 1
fi

# Carregar variáveis de ambiente
export $(cat .env | xargs)

# Executar imediatamente (sem esperar cronjob)
export RUN_NOW=true

echo "🚀 Iniciando análise..."
go run cmd/scheduler/main.go

echo "✅ Teste concluído!"
