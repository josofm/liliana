# Liliana — Spec de Grupos, Temporadas e Partidas (MVP)

## 1. Objetivo

Implementar no Liliana o conceito de grupos de jogo (`playgroups`), permitindo que jogadores participem de vários grupos, registrem partidas de Commander, definam regras próprias e mantenham rankings de jogadores por temporada.

Esta entrega deve permitir que um grupo como o **Commander Night**:

- reúna seus membros;
- documente regras próprias;
- verifique regras objetivas aplicáveis aos decks;
- crie e encerre temporadas;
- registre partidas com jogadores, decks e guests;
- calcule a pontuação dos jogadores na temporada;
- consulte rankings e estatísticas básicas.

## 2. Escopo do MVP

### Incluído

- Criação e gerenciamento de grupos.
- Participação de um jogador em vários grupos.
- Papéis básicos de proprietário e membro.
- Convite de usuários cadastrados para um grupo.
- Regras textuais e regras estruturadas do grupo.
- Validação objetiva de decks quando houver dados suficientes.
- Múltiplas temporadas por grupo.
- No máximo uma temporada ativa por grupo.
- Registro de partidas com dois ou mais participantes.
- Participantes cadastrados ou guests sem identidade persistente.
- Uso de deck próprio, emprestado ou guest.
- Pontuação de jogadores por temporada.
- Estatísticas básicas de jogadores e decks.
- Sugestão de rotação de decks por jogador dentro de cada grupo.

### Fora do escopo

- Elo, Glicko ou qualquer rating dinâmico.
- Pontuação ou ranking sazonal de decks.
- Rating perpétuo de decks.
- Colocações de segundo, terceiro ou quarto lugar.
- Pontos por eliminações, dano, first blood ou achievements.
- Limite de partidas pontuáveis por noite ou período.
- Identidade ou histórico acumulado de guests.
- Reivindicação posterior de uma partida por um guest.
- Múltiplas temporadas ativas simultaneamente.
- Editor genérico de fórmulas de pontuação.
- Automação de regras subjetivas, como “bom senso” ou “deck frustrante”.

## 3. Premissas técnicas

O projeto existente é uma API em Go com Gin e organização inspirada em Clean/Hexagonal Architecture. O Codex deve primeiro inspecionar o repositório e preservar:

- convenções de nomes e pacotes;
- padrões existentes de entidade, service/use case e repository;
- estratégia atual de IDs e datas;
- tratamento de erros;
- padrões de testes;
- contratos HTTP já adotados.

Esta spec define comportamento e domínio. Nomes de pacotes, arquivos e interfaces devem ser adaptados às convenções reais do repositório.

## 4. Modelo de domínio

### 4.1 Group

Representa um playgroup.

Campos mínimos:

```text
id
name
description opcional
owner_id
created_at
updated_at
```

Regras:

- O criador torna-se proprietário e membro do grupo.
- Um jogador pode participar de vários grupos.
- Um grupo pode possuir vários jogadores.
- O nome deve ser obrigatório.
- Somente o proprietário pode editar os dados e regras do grupo no MVP.
- O grupo não é removido fisicamente se possuir partidas; nesse caso, deve ser arquivado ou a remoção rejeitada conforme o padrão existente do projeto.

### 4.2 GroupMember

Relacionamento entre grupo e jogador.

```text
group_id
player_id
role: owner | member
joined_at
```

Restrições:

- A combinação `group_id + player_id` deve ser única.
- Um grupo deve possuir exatamente um proprietário no MVP.
- O proprietário não pode sair enquanto não houver transferência de propriedade.

### 4.3 GroupInvitation

Convite para um usuário cadastrado tornar-se membro.

```text
id
group_id
invited_player_id
invited_by_player_id
status: pending | accepted | declined | cancelled
created_at
responded_at opcional
```

Restrições:

- Não criar convite pendente duplicado para o mesmo jogador e grupo.
- Não convidar alguém que já seja membro.
- Apenas o proprietário pode criar ou cancelar convites no MVP.
- Ao aceitar, criar `GroupMember` de forma atômica e marcar o convite como aceito.

> Guest não utiliza `GroupInvitation`. Guest é apenas um participante avulso registrado dentro de uma partida.

### 4.4 GroupRule

Regras podem ser textuais ou estruturadas.

```text
id
group_id
name
description
rule_type: informational | deck_validation
rule_key opcional
configuration opcional
enabled
created_at
updated_at
```

Diretrizes:

- `informational`: regra exibida aos membros, sem validação automática.
- `deck_validation`: regra objetiva que pode validar dados do deck.
- `rule_key` identifica um validador conhecido pela aplicação.
- `configuration` armazena somente a configuração tipada necessária ao validador.
- Não implementar execução arbitrária de expressões ou scripts.

Validadores inicialmente previstos:

```text
max_game_changers
max_salt_score
forbid_commander_one_card_infinite_combo
forbid_test_proxy_for_game_changer
```

Caso os dados necessários não estejam disponíveis, a avaliação não deve reprovar silenciosamente o deck. Deve retornar `review_required`.

### 4.5 DeckEligibility

Decks pertencem aos jogadores, não aos grupos. Quando o proprietário do deck é membro, o deck fica disponível para seleção nas partidas do grupo.

A elegibilidade é calculada no contexto do grupo:

```text
status: approved | rejected | review_required
violations: lista de regras objetivamente violadas
review_reasons: lista de informações ausentes ou regras que exigem revisão
```

Não criar associação persistente `group_decks` no MVP, salvo se o modelo existente exigir isso. A elegibilidade pode ser calculada sob demanda.

### 4.6 Season

Uma temporada pertence a um grupo.

```text
id
group_id
name
starts_at opcional
ends_at opcional
status: draft | active | finished
win_points, padrão 3
loss_points, padrão 0
draw_points, padrão 1
created_at
updated_at
finished_at opcional
```

Regras:

- Um grupo pode possuir várias temporadas.
- Apenas uma temporada pode estar `active` por grupo.
- A temporada inicial pode ser criada como `draft` ou `active` explicitamente.
- Ativar uma temporada deve falhar se outra estiver ativa.
- Temporada finalizada não pode receber novas partidas pontuáveis.
- Os valores de pontuação devem ser inteiros maiores ou iguais a zero.
- Não existe limite de partidas pontuáveis.
- Alterações no sistema de pontos não podem recalcular silenciosamente partidas antigas.

### 4.7 Match

Representa uma partida realizada dentro de um grupo.

```text
id
group_id
season_id opcional
played_at
status: completed | cancelled
result: winner | draw
notes opcional
created_by_player_id
created_at
updated_at
```

Regras:

- Deve possuir no mínimo dois participantes.
- Uma partida com `season_id` deve pertencer ao mesmo grupo da temporada.
- Apenas temporada ativa pode receber uma nova partida.
- Partida sem temporada registra somente estatísticas gerais.
- Uma partida concluída deve possuir exatamente um vencedor ou ser declarada empate.
- No MVP, não existe colocação entre os perdedores.
- Cancelar ou corrigir uma partida deve reverter seus efeitos no ranking de maneira consistente.

### 4.8 MatchParticipant

Representa um assento/piloto em uma partida.

```text
id
match_id
player_id opcional
guest_name opcional
deck_id opcional
guest_deck_name opcional
result: winner | loser | draw
points_awarded
seat_position opcional
```

Regras de identidade:

- Participante cadastrado: `player_id` preenchido e `guest_name` vazio.
- Guest: `player_id` vazio e `guest_name` preenchido.
- Guest não cria conta, perfil ou histórico persistente.
- Nomes iguais em partidas diferentes não representam a mesma pessoa.
- `guest_name` é apenas um snapshot daquela partida.

Regras de deck:

- Deck cadastrado: `deck_id` preenchido e `guest_deck_name` vazio.
- Deck guest: `deck_id` vazio e `guest_deck_name` preenchido.
- Um guest pode pilotar um deck cadastrado emprestado.
- Um membro pode pilotar um deck cadastrado pertencente a outro jogador.
- O piloto recebe o resultado pessoal; o proprietário do deck não recebe pontos por tê-lo emprestado.
- O mesmo jogador cadastrado não pode ocupar dois assentos na mesma partida.
- O mesmo deck cadastrado não pode ocupar dois assentos na mesma partida.

Regras de pontuação:

- Participante cadastrado vencedor recebe `season.win_points`.
- Participante cadastrado derrotado recebe `season.loss_points`.
- Participante cadastrado em empate recebe `season.draw_points`.
- Guest sempre recebe zero pontos e nunca aparece no ranking.
- `points_awarded` deve ser salvo como snapshot no participante.
- Partida sem temporada atribui zero pontos a todos.

## 5. Ranking e estatísticas

### 5.1 Ranking da temporada

O ranking inclui somente jogadores cadastrados que participaram de partidas daquela temporada.

Campos retornados:

```text
player_id
player_name
points
games_played
wins
losses
draws
win_rate
```

Ordem:

1. Maior quantidade de pontos.
2. Maior quantidade de vitórias.
3. Maior taxa de vitória.
4. Maior quantidade de partidas.
5. Nome ou ID como ordenação estável final.

Não implementar confronto direto no MVP, pois mesas multiplayer tornam esse desempate ambíguo.

### 5.2 Estatísticas de deck

Decks não recebem pontos nem rating no MVP.

Estatísticas permitidas:

```text
games_played
wins
losses
draws
win_rate
last_played_at
```

Regras:

- As estatísticas acompanham o deck, independentemente do piloto.
- Um deck emprestado participa normalmente das estatísticas do deck.
- O proprietário não recebe estatísticas pessoais pelo resultado de outro piloto.
- Deck guest não possui histórico agregado.
- Permitir consulta geral e filtro por grupo quando for simples dentro da arquitetura atual.
- Não criar ranking sazonal de decks.

### 5.3 Sugestão de rotação de decks

O sistema deve ajudar cada jogador a descobrir qual deck está há mais tempo sem ser utilizado naquele grupo. A rotação é apenas uma sugestão de escolha e nunca deve bloquear o jogador ou obrigá-lo a seguir a fila.

A fila é individual por jogador e por grupo. Ela considera somente partidas em que o próprio jogador pilotou um deck cadastrado que lhe pertence.

Ordenação sugerida:

1. Decks do jogador que nunca foram pilotados por ele naquele grupo.
2. Decks ordenados pelo uso mais antigo (`last_played_at` crescente).
3. Nome ou ID do deck como critério estável de desempate.

Cada item pode retornar:

```text
deck_id
deck_name
last_played_at opcional
games_played_by_player_in_group
suggested_next: boolean
```

Regras:

- Apenas o primeiro deck elegível da fila recebe `suggested_next = true`.
- Decks reprovados pelas regras objetivas do grupo não entram na sugestão principal.
- Decks com `review_required` podem aparecer identificados como pendentes, depois dos decks aprovados, mas não devem ser a recomendação principal enquanto existir deck aprovado.
- Deck emprestado pilotado pelo jogador não altera a fila dos decks próprios dele.
- Se outro participante pilotar um deck emprestado do jogador, isso também não altera a fila pessoal do proprietário.
- Deck guest não participa da rotação.
- Partidas canceladas não contam como uso.
- Partidas com ou sem temporada contam para a rotação, desde que pertençam ao grupo.
- A sugestão deve ser derivada do histórico de partidas; não criar uma fila persistida como segunda fonte de verdade.
- A interface pode mostrar algo como: `Próximo da fila: Saruman Army — último uso há 1 mês`.

## 6. Regras iniciais do Commander Night

O sistema deve conseguir representar estas regras:

### Estruturadas ou candidatas à validação

- Máximo de 3 Game Changers por deck.
- SALT máximo de 50.
- Sem combo infinito de comandante mais uma carta.
- Game Changers não podem ser utilizados como proxy de teste.

### Informativas

- Combos nas 99 são permitidos.
- Deck com combo infinito não usa tutor para buscar suas peças.
- Deck sem combo pode usar tutores normalmente.
- Quem usar tutor deve saber o que procura; aproximadamente 30 segundos como referência.
- Mass Land Destruction é permitido quando fizer parte da estratégia e encaminhar o fim do jogo.
- Turnos extras são permitidos; decks focados em monopolizar turnos, não.
- Quem possui a carta original pode usar proxy dela em outros decks.
- Carta não possuída pode ser testada como proxy por até três partidas.
- Locks, stax e cartas extremamente frustrantes dependem de bom senso.
- Deck apontado repetidamente como muito acima dos demais deve ser revisto.
- Deck que atropela consistentemente as mesas pode ser revisto mesmo respeitando as demais regras.

As regras informativas devem ser armazenadas e exibidas, mas não devem bloquear automaticamente um deck.

## 7. Casos de uso

### Grupos

- Criar grupo.
- Consultar grupo.
- Listar grupos do jogador autenticado.
- Atualizar grupo.
- Arquivar ou remover grupo conforme as restrições do domínio.
- Listar membros.

### Convites e membros

- Convidar usuário cadastrado.
- Listar convites pendentes.
- Aceitar ou recusar convite.
- Cancelar convite.
- Remover membro.
- Sair do grupo.

### Regras

- Criar regra informativa.
- Criar regra estruturada suportada.
- Atualizar, ativar ou desativar regra.
- Listar regras do grupo.
- Avaliar elegibilidade de um deck.

### Temporadas

- Criar temporada.
- Listar temporadas do grupo.
- Consultar temporada.
- Ativar temporada.
- Finalizar temporada.
- Consultar ranking.

### Partidas

- Registrar partida com ou sem temporada.
- Incluir membros e guests.
- Selecionar deck próprio ou emprestado.
- Informar deck guest sem cadastro.
- Consultar partida.
- Listar partidas do grupo.
- Corrigir ou cancelar partida.

### Rotação

- Consultar a fila sugerida de decks de um jogador no grupo.
- Destacar o próximo deck sugerido ao preparar uma nova partida.
- Permitir selecionar qualquer outro deck elegível sem confirmação adicional.

## 8. Contratos HTTP sugeridos

Adaptar os paths aos padrões existentes do projeto.

```text
POST   /groups
GET    /groups
GET    /groups/:groupID
PATCH  /groups/:groupID
DELETE /groups/:groupID

GET    /groups/:groupID/members
DELETE /groups/:groupID/members/:playerID

POST   /groups/:groupID/invitations
GET    /groups/:groupID/invitations
POST   /group-invitations/:invitationID/accept
POST   /group-invitations/:invitationID/decline
DELETE /group-invitations/:invitationID

POST   /groups/:groupID/rules
GET    /groups/:groupID/rules
PATCH  /groups/:groupID/rules/:ruleID
DELETE /groups/:groupID/rules/:ruleID
GET    /groups/:groupID/decks/:deckID/eligibility
GET    /groups/:groupID/players/:playerID/decks/rotation

POST   /groups/:groupID/seasons
GET    /groups/:groupID/seasons
GET    /groups/:groupID/seasons/:seasonID
POST   /groups/:groupID/seasons/:seasonID/activate
POST   /groups/:groupID/seasons/:seasonID/finish
GET    /groups/:groupID/seasons/:seasonID/ranking

POST   /groups/:groupID/matches
GET    /groups/:groupID/matches
GET    /groups/:groupID/matches/:matchID
PATCH  /groups/:groupID/matches/:matchID
POST   /groups/:groupID/matches/:matchID/cancel
```

## 9. Exemplo de registro de partida

```json
{
  "season_id": "season-uuid",
  "played_at": "2026-09-15T23:30:00Z",
  "result": "winner",
  "notes": "Commander Night",
  "participants": [
    {
      "player_id": "player-1",
      "deck_id": "deck-y",
      "result": "winner"
    },
    {
      "player_id": "player-2",
      "deck_id": "deck-b",
      "result": "loser"
    },
    {
      "guest_name": "Rafael",
      "deck_id": "borrowed-deck-c",
      "result": "loser"
    },
    {
      "guest_name": "Marina",
      "guest_deck_name": "Elfos da Lathril",
      "result": "loser"
    }
  ]
}
```

O backend calcula `points_awarded`; o cliente não deve enviar nem controlar esse valor.

## 10. Consistência e transações

Devem ocorrer atomicamente:

- criação do grupo e inclusão do proprietário;
- aceite do convite e inclusão do membro;
- criação da partida e de todos os seus participantes;
- cálculo e persistência dos pontos concedidos;
- correção ou cancelamento de partida e atualização consistente das consultas agregadas;
- ativação de temporada garantindo unicidade da temporada ativa.

Preferir rankings e estatísticas calculados a partir das partidas na primeira versão, salvo se o volume ou a arquitetura existente justificarem agregados persistidos. Não manter duas fontes de verdade sem necessidade.

## 11. Validações essenciais

- Grupo, temporada, regra, jogador e deck referenciados devem existir.
- Usuários devem ter autorização para operar no grupo.
- `season_id`, quando informado, deve pertencer ao grupo da partida.
- Jogador cadastrado participante deve ser membro do grupo.
- Guest não precisa ser membro.
- `player_id` e `guest_name` são mutuamente exclusivos.
- `deck_id` e `guest_deck_name` são mutuamente exclusivos.
- Cada participante deve possuir uma forma válida de identificação e uma forma válida de identificar o deck.
- Deck emprestado deve existir, mas seu proprietário não precisa participar da partida.
- Uma partida deve possuir exatamente um vencedor ou todos os participantes com resultado de empate.
- Não aceitar combinações parciais ou contraditórias de resultados.
- Pontos devem ser calculados pelo servidor usando o snapshot da configuração da temporada.

## 12. Segurança e autorização

- Somente o proprietário gerencia dados, regras, membros e temporadas do grupo no MVP.
- Membros podem consultar o grupo, regras, temporadas, rankings e partidas.
- Membros podem registrar partidas do grupo.
- Apenas o criador da partida ou o proprietário pode corrigi-la ou cancelá-la.
- Usuários externos não devem acessar grupos privados.
- Nunca confiar em `points_awarded`, `owner_id` ou papéis enviados pelo cliente.

Se autenticação ainda não estiver implementada no projeto, preservar essas regras no domínio/interfaces e documentar o bloqueio, sem criar uma solução improvisada incompatível com a arquitetura planejada.

## 13. Critérios de aceitação

1. Um jogador consegue criar um grupo e torna-se seu proprietário.
2. O mesmo jogador consegue participar de mais de um grupo.
3. O proprietário consegue convidar um usuário cadastrado.
4. O convidado consegue aceitar e tornar-se membro.
5. O grupo consegue armazenar regras informativas e estruturadas.
6. Um deck pode retornar `approved`, `rejected` ou `review_required` conforme as regras estruturadas.
7. Um grupo consegue possuir várias temporadas históricas.
8. Não é possível manter duas temporadas ativas no mesmo grupo.
9. Uma partida pode ser registrada na temporada ativa ou sem temporada.
10. Uma partida aceita membros, guests, decks próprios, decks emprestados e decks guest.
11. Guest não cria identidade persistente nem aparece no ranking.
12. O vencedor cadastrado recebe três pontos por padrão; derrotados recebem zero e empates recebem um.
13. Alterar a configuração futura não modifica `points_awarded` de partidas anteriores.
14. O ranking ordena jogadores conforme os critérios definidos.
15. Decks exibem apenas estatísticas básicas e não recebem pontos ou rating.
16. Corrigir ou cancelar uma partida altera corretamente ranking e estatísticas.
17. O jogador consegue consultar seus decks ordenados pelos que nunca foram usados no grupo e depois pelo uso mais antigo.
18. O uso de deck emprestado não altera a fila pessoal do proprietário nem a do piloto.
19. A sugestão de rotação não impede a seleção de outro deck elegível.
20. Casos de uso, domínio, repositórios e handlers possuem testes compatíveis com o padrão atual do projeto.

## 14. Estratégia de implementação sugerida

Implementar em incrementos revisáveis:

1. Inspecionar arquitetura, autenticação, entidades e persistência existentes.
2. Implementar `Group`, `GroupMember` e permissões.
3. Implementar convites de usuários cadastrados.
4. Implementar temporadas e garantia de apenas uma ativa.
5. Implementar regras informativas.
6. Implementar partidas e participantes, incluindo guests e decks emprestados.
7. Implementar pontuação e ranking de jogadores.
8. Implementar estatísticas básicas e sugestão de rotação de decks.
9. Implementar regras estruturadas e elegibilidade de decks gradualmente, conforme os metadados disponíveis.
10. Adicionar testes unitários, de repositório e HTTP em cada incremento.

Não implementar Elo, rating de decks ou novos mecanismos de pontuação durante esta entrega.

## 15. Questões que o Codex deve resolver pela inspeção do repositório

Antes de alterar código, responder no plano de implementação:

- Qual entidade atual representa jogador: `User`, `Player` ou ambas?
- Já existe entidade e CRUD de deck?
- Qual mecanismo de autenticação/autorização está ativo?
- Qual banco e estratégia de migrations estão sendo usados atualmente?
- Rankings devem ser calculados por query ou pelo service com o volume previsto?
- Como o projeto representa timestamps, UUIDs, paginação e erros HTTP?
- Quais validadores de deck são possíveis com os dados atuais?

Se algum requisito depender de estrutura inexistente, o Codex deve apontar a dependência e propor a menor adaptação possível antes de implementar.
