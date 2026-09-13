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

Convites e aceite ficam para o próximo incremento. `AddMember` é uma operação
interna de persistência preparada para esse fluxo; não há rota de entrada direta.
Remoção/arquivamento de grupos, transferência de propriedade, regras, temporadas,
partidas, rankings e rotação de decks também ficam para próximas etapas. Não há
associação persistente entre grupos e decks. Os metadados atuais de cartas não
incluem SALT, Game Changers ou histórico de proxies para os validadores previstos.
Rankings futuros poderão ser calculados por consulta ao histórico de partidas.
