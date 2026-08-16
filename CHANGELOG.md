# Changelog

## Unreleased

### Added

- Importação automática de decks públicos do Archidekt pelo `source_link`.
- Extração de nome, formato, cores, comandante e cartas.
- Catálogo normalizado de cartas e relacionamento com decks por Oracle ID.
- Validação e enriquecimento de listas manuais pelo Scryfall.
- Cadastro manual de cartas por lista no formato `quantidade nome`.
- Endpoint para adicionar cartas a um deck existente.
- Testes unitários e integração com um deck real do Archidekt.
- Comando `make publish` para publicar a imagem de produção no GHCR.

### Changed

- Criação e atualização de decks agora aceitam dados obtidos pelo link.
- Imagens de produção aceitam tags através de `VERSION`.

## 2026-08-15

- Publicação automatizada da imagem da API no GHCR.
- Ajustes de Docker, configuração HTTP, health check e CORS.

## 2026-07-06

- Ampliação dos formatos de deck aceitos.
- Suporte ao comandante opcional em formatos que não o utilizam.

## 2026-05-28

- Configuração da aplicação por ambiente.
- Migrações automáticas opcionais antes da inicialização.
- Comandos Make e Docker Compose para executar migrações.

## 2026-05-16

- Persistência de usuários e decks no PostgreSQL.
- Runner de testes de integração com banco de dados.
- Migrações iniciais de usuários e decks.

## 2026-05-15

- Autenticação JWT com access e refresh tokens.
- Middleware para proteger as rotas da API.
- Cobertura de testes dos endpoints HTTP.

## 2025-08-17

- Validação estruturada das requisições de usuários e decks.
- Respostas de erro de validação padronizadas.

## 2025-08-03

- CRUD completo de decks.
- Repositório em memória e testes de handlers, serviços e repositórios.

## 2025-05-01

- CRUD de usuários com separação entre handler, serviço e repositório.
- Testes unitários e integração com Docker Compose.

## 2024-03-31

- Estrutura inicial da API em Go com Gin.
- Configuração, servidor HTTP, logger, Docker e Makefile.
- Pipeline de CI com GitHub Actions.
