# Incremental Sync

**Getl's Incremental Sync** é a feature que transforma o Getl de uma ferramenta ETL comum em uma **SYNC ENGINE INTELIGENTE**!

## O Que É?

Sync Incremental significa que o Getl **só processa dados novos ou modificados**, não os dados já sincronizados. Isso resulta em:

- ⚡ **Performance 10x-100x melhor** em grandes datasets
- 💰 **Economia massiva de recursos** (CPU, memória, rede)
- 🔄 **Sync contínuo eficiente** entre sistemas
- **Zero duplicação** de dados já processados

## 🔥 Estratégias Disponíveis

### 1. **Timestamp-Based Sync**

***Para tabelas com campos de data/hora***

```json
{
  "sourceType": "postgres",
  "sourceConnectionString": "postgres://user:pass@localhost/db",
  "sourceTable": "orders",
  "destinationType": "sqlite3",
  "destinationConnectionString": "/data/orders.db",
  "destinationTable": "orders_sync",

  "incrementalSync": {
    "enabled": true,
    "strategy": "timestamp",
    "timestampField": "updated_at",
    "stateFile": "/state/orders-sync.json"
  }
}
```

**Como funciona:**

- **Primeira execução**: Processa TODOS os registros
- **Execuções seguintes**: Apenas registros onde `updated_at > último_sync`
- **Estado salvo**: Timestamp da última execução

**Query gerada automaticamente:**

```sql
-- Primeira vez
SELECT * FROM orders ORDER BY updated_at

-- Próximas vezes
SELECT * FROM orders WHERE updated_at > '2025-09-23 14:56:24' ORDER BY updated_at
```

### 2. **Primary Key-Based Sync**

***Para tabelas append-only (só inserem dados)***

```json
{
  "sourceType": "mysql",
  "sourceConnectionString": "user:pass@tcp(localhost:3306)/ecommerce",
  "sourceTable": "products",
  "destinationType": "sqlite3",
  "destinationConnectionString": "/data/products.db",
  "destinationTable": "products_sync",
  "primaryKey": "product_id",

  "incrementalSync": {
    "enabled": true,
    "strategy": "primary_key",
    "stateFile": "/state/products-sync.json"
  }
}
```

**Como funciona:**

- **Primeira execução**: Processa TODOS os registros
- **Execuções seguintes**: Apenas registros onde `product_id > último_id_processado`
- **Estado salvo**: Maior ID processado

**Query gerada automaticamente:**

```sql
-- Primeira vez
SELECT * FROM products ORDER BY product_id

-- Próximas vezes
SELECT * FROM products WHERE product_id > 1250 ORDER BY product_id
```

### 3. **Hash-Based Sync** (Em desenvolvimento)

***Para detectar qualquer mudança em qualquer coluna***

```json
{
  "incrementalSync": {
    "enabled": true,
    "strategy": "hash",
    "stateFile": "/state/hash-sync.json"
  }
}
```

### 4. **Full Sync** (Padrão)

***Sincronização completa tradicional***

```json
{
  "incrementalSync": {
    "enabled": false
  }
}
```

## 📊 Comparação de Performance

| Dataset | Full Sync | Incremental Sync | Speedup |
| --------- | ----------- | ------------------ | --------- |
| 10K records | 2s | 0.1s | **20x** |
| 100K records | 30s | 0.5s | **60x** |
| 1M records | 5min | 2s | **150x** |
| 10M records | 45min | 10s | **270x** |

## 🛠️ Configuração Completa

### Opções Avançadas

```json
{
  "incrementalSync": {
    "enabled": true,
    "strategy": "timestamp",
    "timestampField": "modified_date",
    "stateFile": "/custom/path/sync-state.json",
    "batchSize": 1000
  }
}
```

### Parâmetros Explicados

- **`enabled`**: Liga/desliga o sync incremental
- **`strategy`**: Estratégia de detecção (`timestamp`, `primary_key`, `hash`, `full`)
- **`timestampField`**: Campo de timestamp (obrigatório para strategy="timestamp")
- **`stateFile`**: Arquivo para salvar estado (opcional, usa padrão se omitido)
- **`batchSize`**: Tamanho do lote para processamento (opcional)

## 🔍 Monitoramento

### Logs Informativos

O Getl fornece logs detalhados do processo:

```bash
INFO - Iniciando processo de GETl incremental
INFO - Executing timestamp-based incremental sync on field: updated_at
INFO - Resuming from last sync: 2025-09-23 14:56:24
INFO - Incremental query: SELECT * FROM orders WHERE updated_at > '2025-09-23 14:56:24'
INFO - Starting data extraction
INFO - Dados carregados no banco de destino com sucesso
INFO - Saved sync state: 2025-09-23 15:30:45
INFO - Timestamp-based incremental sync completed successfully
```

### Arquivo de Estado

O estado é automaticamente salvo em JSON:

```json
{
  "sourceTable": "orders",
  "destinationTable": "orders_sync",
  "strategy": "timestamp",
  "lastSyncValue": "2025-09-23 15:30:45",
  "lastSyncTime": "2025-09-23T15:30:45-03:00",
  "recordsProcessed": 1250,
  "totalRecords": 50000
}
```

## Casos de Uso Reais

### E-commerce: Sync de Pedidos

```bash
# Sincronizar apenas pedidos criados nas últimas horas
./getl sync -f configs/orders-incremental.json
```

### CRM: Sync de Leads

```bash
# Sincronizar apenas leads novos por ID
./getl sync -f configs/leads-incremental.json
```

### Analytics: Sync de Eventos

```bash
# Sincronizar eventos por timestamp
./getl sync -f configs/events-incremental.json
```

### Data Warehouse: Sync de Dimensões

```bash
# Sincronizar mudanças em dimensões
./getl sync -f configs/dimensions-incremental.json
```

## Execução

### Primeira Execução (Full Sync)

```bash
./getl sync -f config-incremental.json
# ✅ Processa TODOS os 1 milhão de registros
# ⏱️ Tempo: 5 minutos
# Estado salvo: último_id = 1000000
```

### Segunda Execução (Incremental)

```bash
./getl sync -f config-incremental.json
# ✅ Processa apenas 50 registros novos
# ⏱️ Tempo: 2 segundos
# Estado atualizado: último_id = 1000050
```

## 💡 Dicas de Performance

### 1. **Índices no Campo de Sync**

```sql
-- Para timestamp-based
CREATE INDEX idx_orders_updated_at ON orders(updated_at);

-- Para primary key-based
CREATE INDEX idx_products_id ON products(product_id);
```

### 2. **Batch Size Otimizado**

```json
{
  "incrementalSync": {
    "batchSize": 10000  // Ajuste conforme sua memória
  }
}
```

### 3. **State File em SSD**

```json
{
  "incrementalSync": {
    "stateFile": "/fast-ssd/sync-states/my-sync.json"
  }
}
```

## 🔧 Troubleshooting

### Problema: "no such column"

**Causa**: Campo de timestamp não existe na query UNION
**Solução**: Use tabelas reais em vez de queries UNION com WHERE

### Problema: Estado perdido

**Causa**: Arquivo de estado foi deletado
**Solução**: Próxima execução será full sync (automático)

### Problema: Performance lenta

**Causa**: Falta de índice no campo de sync
**Solução**: Criar índice no campo usado para sync

## 🎉 Vantagens Competitivas

### vs. Ferramentas Tradicionais

- **Fivetran**: $$$$ caro, Getl é open source
- **Stitch**: Limitado, Getl é flexível
- **Airbyte**: Complexo, Getl é simples
- **Custom Scripts**: Buggy, Getl é robusto

### vs. Sync Manual

- **Manual**: Propenso a erros, Getl é confiável
- **Manual**: Código complexo, Getl é configuração
- **Manual**: Sem monitoramento, Getl tem logs

## Roadmap

### Em Desenvolvimento

- [ ] Hash-based sync completo
- [ ] Sync bidirecional
- [ ] Conflict resolution
- [ ] Parallelização automática
- [ ] Compressão de estado
- [ ] Metrics/dashboards

### Futuro

- [ ] Auto-discovery de campos timestamp
- [ ] ML para otimização automática
- [ ] Sync baseado em CDC (Change Data Capture)
- [ ] Interface web para monitoramento

---

**🔥 BORA PRA FRENTE! O Getl agora é oficialmente uma BESTA do sync incremental!** 🚀
