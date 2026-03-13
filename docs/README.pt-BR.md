# GETL

English version: [../README.md](../README.md)

## Sumário

- [Visão Geral](#visão-geral)
- [Escopo Atual do Produto](#escopo-atual-do-produto)
- [Estado Operacional Atual](#estado-operacional-atual)
- [Capacidades Principais](#capacidades-principais)
- [Visão Geral da Arquitetura](#visão-geral-da-arquitetura)
- [Estrutura do Repositório](#estrutura-do-repositório)
- [Notas de Instalação](#notas-de-instalação)
- [Comandos Principais](#comandos-principais)
- [Modelo de Configuração](#modelo-de-configuração)
- [Caso de Uso com PostgreSQL e Sync de Catálogo](#caso-de-uso-com-postgresql-e-sync-de-catálogo)
- [Papel Atual no Ecossistema](#papel-atual-no-ecossistema)
- [Limitações Atuais](#limitações-atuais)
- [Screenshots](#screenshots)

## Visão Geral

`GETL` é o utilitário de ETL e sincronização do ecossistema Kubex.

Ele foi desenhado para mover dados entre fontes e destinos heterogêneos por meio de pipelines configuráveis, mantendo praticidade suficiente para ser usado em fluxos reais de repositório para runtime.

No estágio atual, o `GETL` já não é apenas uma ideia genérica de ETL. Ele já é usado materialmente como parte do caminho de ingestão do catálogo Sankhya que sustenta a geração BI guiada por metadados no `GNyx`.

## Escopo Atual do Produto

Áreas práticas atuais incluem:

- execução configurável de ETL e sync
- fluxos de origem e destino orientados a SQL
- caminhos de ingestão para banco e arquivos
- materialização full-refresh ou em estilo batch
- suporte utilitário para sistemas heterogêneos
- código Go reutilizável mais superfícies orientadas a CLI

Ele também inclui uma superfície mais ampla de longo prazo envolvendo incremental sync, múltiplos backends e estratégias de sincronização mais ricas.

## Estado Operacional Atual

Verdades operacionais relevantes hoje:

- o `GETL` já está sendo usado para carregar CSVs de metadados Sankhya no PostgreSQL
- o schema alvo real dessa frente é `sankhya_catalog`
- arquivos de configuração agora podem expandir variáveis de ambiente, o que tornou configs versionadas realmente práticas
- o repositório é valioso tanto como ferramenta quanto como dependência importável

## Capacidades Principais

Capacidades concretas atuais incluem:

- fluxos configuráveis de extração e carga
- helpers de ingestão e carga orientados a SQL
- materialização de tabelas de destino
- execução batch a partir de inputs como CSV
- suporte mais amplo a padrões heterogêneos de movimentação de dados
- loading de config com expansão de env

## Visão Geral da Arquitetura

O `GETL` se organiza em torno de:

- entrypoints de CLI
- orquestração de config e sync
- camadas utilitárias de SQL/dados
- suporte a extração e transformação
- tipos compartilhados orientados a ETL

## Estrutura do Repositório

```text
cmd/                    entrypoints da CLI
sql/                    caminhos ETL orientados a SQL
sync/                   lógica de sincronização
utils/                  loading e helpers de apoio
extr/                   código relacionado a extração
etypes/                 tipos compartilhados orientados a ETL
```

## Notas de Instalação

Restrição importante de build:

- como o `GETL` depende de `godror` para suporte a Oracle, alguns ambientes de build exigem `CGO=1`

Exemplo:

```bash
CGO=1 go build ./...
```

Isso importa não só ao compilar o `GETL` diretamente, mas também quando outro projeto o importa transitivamente.

## Comandos Principais

Compilar:

```bash
go build ./...
```

Rodar testes:

```bash
go test ./...
```

A superfície exata de comandos operacionais pode variar por fluxo, mas o `GETL` já está sendo consumido hoje por orquestração de nível mais alto no `GNyx`.

## Modelo de Configuração

O `GETL` é orientado por configuração.

Melhoria prática recente:

- a expansão de variáveis de ambiente no loading de configs agora funciona, o que significa que manifests versionados de sync não precisam mais hardcodar DSNs locais ou valores específicos de máquina

Isso viabilizou manter configs reutilizáveis no versionamento e ainda assim executá-las em ambientes diferentes.

## Caso de Uso com PostgreSQL e Sync de Catálogo

Agora existe um caso de uso real importante:

- CSVs de metadados BI do Sankhya são carregados no PostgreSQL
- schema de destino: `sankhya_catalog`
- registry/governança ficam no `Domus`
- o comando de orquestração atualmente vive no `GNyx`

Na prática:

- o `GETL` cuida da ingestão/materialização
- o `Domus` hospeda o runtime PostgreSQL ativo e o external metadata registry
- o `GNyx` orquestra o fluxo de sync orientado ao domínio

Este é um bom exemplo de o `GETL` ser usado como ferramenta real do ecossistema, e não como experimento isolado.

## Papel Atual no Ecossistema

Hoje o `GETL` é especialmente relevante para:

- frentes de ingestão de metadados
- carga de catálogos no PostgreSQL
- cenários futuros mais amplos de ETL e sync entre projetos Kubex

Seu papel cresceu materialmente quando a prova de conceito de BI guiada por metadados se tornou uma feature real no `GNyx`.

## Limitações Atuais

Limitações atuais incluem:

- o uso real no ecossistema ainda está concentrado em alguns fluxos concretos, e não em toda capacidade teórica possível
- a superfície total é mais ampla do que a slice hoje battle-tested
- `godror` e os requisitos de CGO adicionam restrições práticas de build em alguns ambientes

## Screenshots

Sugestões de placeholders:

- `[Screenshot Placeholder: saída do comando de sync]`
- `[Screenshot Placeholder: tabelas alvo no PostgreSQL]`
- `[Screenshot Placeholder: exemplo de config]`
