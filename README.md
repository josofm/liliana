# liliana
Gideon died for her to live

O ambiente Docker Compose usa PostgreSQL 18 (`postgres:18-alpine`). O banco de
desenvolvimento/testes usa `tmpfs` em `/var/lib/postgresql`, compatível com o
`PGDATA` padrão `/var/lib/postgresql/18/docker` da imagem. Os dados são temporários
e não persistem quando o container do banco é parado.

## Grupos — primeiro incremento

Jogadores são os usuários existentes (`users`); `player_id` referencia `users.id`.
As rotas de grupos exigem JWT. IDs são inteiros e as datas são UTC/RFC 3339.

| Método | Rota | Comportamento |
| --- | --- | --- |
| POST | `/groups/` | Cria grupo com `name` obrigatório (1–100 caracteres) e `description` opcional. O usuário autenticado torna-se proprietário e membro atomicamente. |
| GET | `/groups/` | Lista somente os grupos do usuário autenticado, por ID. |
| GET | `/groups/:id` | Consulta permitida aos membros. |
| PATCH | `/groups/:id` | Proprietário altera `name` e/ou `description`; string vazia limpa a descrição. |
| GET | `/groups/:id/members` | Membros consultam os vínculos, papéis e datas de entrada. |
| DELETE | `/groups/:id/members/:playerID` | Proprietário remove um membro, ou o próprio membro sai. O proprietário não pode sair (409). |

`owner_id` e papéis enviados pelo cliente não são usados. Usuários externos recebem
403 ao consultar um grupo existente. Listagens vazias retornam `[]`, sem paginação,
seguindo o padrão atual. Erros usam `{"error":"..."}`.

A migration `000004_create_groups.up.sql` cria `groups` e `group_members`, com
vínculo único por grupo/jogador e chave estrangeira diferida que garante a presença
do proprietário entre os membros. `owner_id` é a fonte do papel `owner`; os demais
vínculos retornam `member`. Usuários vinculados não podem ser excluídos do banco.

Convites e aceite de usuários cadastrados estão implementados conforme a seção abaixo.
`AddMember` permanece uma operação interna de persistência; não há rota de entrada direta.
Remoção/arquivamento de grupos, transferência de propriedade, regras, temporadas,
partidas, rankings e rotação de decks também ficam para próximas etapas. Não há
associação persistente entre grupos e decks. Os metadados atuais de cartas não
incluem SALT, Game Changers ou histórico de proxies para os validadores previstos.
Rankings futuros poderão ser calculados por consulta ao histórico de partidas.

## Convites para usuários cadastrados

Todas as rotas exigem JWT. A busca usa somente o e-mail; busca por nome fica
para uma etapa futura com username próprio. A correspondência é parcial, sem distinguir maiúsculas e
minúsculas. `q` deve ter de 2 a 100 caracteres após remover espaços externos.
Retorna até 20 usuários, ordenados por nome e ID, somente com `id`, `name` e `email`.
Para refinar resultados, informe mais caracteres. A busca inclui membros existentes;
a criação do convite rejeita quem já pertence ao grupo.

| Método | Rota | Comportamento |
| --- | --- | --- |
| GET | `/users/search?q=marina@example.com` | Busca usuários cadastrados somente pelo e-mail. |
| POST | `/groups/:id/invitations` | Proprietário convida com `{"invited_player_id": 2}`. Retorna 201. |
| GET | `/groups/:id/invitations` | Proprietário lista convites pendentes do grupo. |
| GET | `/group-invitations/` | Usuário lista seus convites recebidos e pendentes. |
| POST | `/group-invitations/:id/accept` | Destinatário aceita e torna-se membro. Retorna 200 com o convite. |
| POST | `/group-invitations/:id/decline` | Destinatário recusa. Retorna 200 com o convite. |
| DELETE | `/group-invitations/:id` | Proprietário cancela um convite pendente. Retorna 204. |

O fluxo da interface é buscar um usuário, selecionar seu ID e enviar o convite.
O destinatário vê `group_name` na caixa de convites, sem obter acesso ao grupo antes
de aceitar. O servidor define autor, status e datas; campos extras enviados pelo
cliente não substituem esses valores. Listagens retornam `[]` quando vazias.

Estados: `pending` → `accepted`, `declined` ou `cancelled`. `responded_at` é nulo
enquanto pendente e preenchido em qualquer encerramento. Um convite encerrado não
pode receber outra resposta (409). É possível criar novo convite após recusa ou
cancelamento, ou após o membro sair. Convites duplicados pendentes e convites para
membros retornam 409; IDs inexistentes retornam 404; operações sem permissão, 403.

A migration `000005_create_group_invitations.up.sql` adiciona o histórico de convites
e a unicidade dos pendentes. Aceite e inclusão do membro são atômicos; requisições
concorrentes não duplicam vínculos nem sobrescrevem respostas já concluídas.
Nesta etapa não há envio de e-mail, convite de pessoa sem conta ou expiração automática.
