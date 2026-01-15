# 🚀 Guia de Início Rápido - EVA-Markov

## Pré-requisitos

- Go 1.21+
- PostgreSQL com extensão pgvector
- Google Gemini API Key
- Acesso ao banco de dados EVA

## Instalação

### 1. Clone e Configure

```bash
cd d:\dev\EVA\EVA-Markov
cp .env.example .env
```

### 2. Edite o `.env`

```env
GOOGLE_API_KEY=sua_chave_aqui
DATABASE_URL=postgresql://user:pass@host:5432/eva_db
```

### 3. Instale Dependências

```bash
go mod download
```

### 4. Execute as Migrations

```bash
# Conecte ao PostgreSQL e execute:
psql -U user -d eva_db -f migrations/001_initial_schema.sql
```

## Executar Localmente

### Modo de Teste (Execução Imediata)

```bash
# Windows
$env:RUN_NOW="true"
go run cmd/scheduler/main.go

# Linux/Mac
RUN_NOW=true go run cmd/scheduler/main.go
```

### Modo Produção (Cronjob)

```bash
go run cmd/scheduler/main.go
# Aguarda até 23:00 para executar
```

## Deploy para Cloud Run

```bash
chmod +x deploy.sh
./deploy.sh
```

## Verificar Resultados

### Ver Notas Comportamentais

```sql
SELECT * FROM behavioral_notes 
WHERE idoso_id = 1 
ORDER BY created_at DESC;
```

### Ver Prompts Otimizados

```sql
SELECT * FROM prompt_templates_personalized 
WHERE idoso_id = 1 AND active = true;
```

### Dashboard de Status

```sql
SELECT * FROM v_idoso_optimization_status;
```

## Troubleshooting

### Erro: "GOOGLE_API_KEY é obrigatório"
- Verifique se o `.env` está configurado corretamente

### Erro: "Erro ao conectar ao banco"
- Confirme que o PostgreSQL está rodando
- Verifique a `DATABASE_URL`

### Nenhum insight gerado
- Verifique se há conversas nas últimas 24h
- Confirme que a tabela `episodic_memories` existe

## Próximos Passos

1. ✅ Execute o teste inicial
2. ✅ Verifique os logs
3. ✅ Analise os insights gerados
4. ✅ Configure o cronjob para produção
5. ✅ Monitore os resultados

## Suporte

Para dúvidas, consulte o README.md principal ou os comentários no código.
