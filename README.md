# falkcroche

Mini-ERP do ateliê Falk Crochê, feito com Go, SQLite e templates `templ`.

## Executar

Crie um arquivo `.env` na raiz do projeto usando `.env.example` como referência:

```dotenv
ADMIN_PASSWORD=troque-por-uma-senha-forte
ADMIN_SESSION_KEY=troque-por-uma-chave-aleatoria-com-32-caracteres
WHATSAPP_NUMERO=55DDDSEUNUMERO
```

O arquivo `.env` é local e ignorado pelo Git. Depois de configurá-lo uma vez, inicie o sistema normalmente:

```sh
go run .
```

O painel fica em `http://127.0.0.1:8080/admin/login`. O catálogo local fica em `http://127.0.0.1:8080/`. O servidor aceita conexões somente do próprio computador.
As sessões expiram após 12 horas. Se a senha ou a chave não estiver configurada, o painel permanece bloqueado.
O número de WhatsApp deve incluir o código do país e DDD; ele é usado nos botões de contato do catálogo público.

Para regenerar os arquivos Go depois de editar um arquivo `.templ`:

```sh
go run github.com/a-h/templ/cmd/templ generate
```

## Publicar a vitrine no GitHub Pages

O exportador lê o banco local e gera somente a página pública e o logo em `docs/`. Ele também carrega o número de WhatsApp do `.env`:

```sh
go run ./cmd/exportar-vitrine
```

No GitHub, abra **Settings > Pages**, escolha **Deploy from a branch**, selecione a branch `main` e a pasta `/docs`. O endereço público será exibido nessa tela. A exportação contém os produtos publicados no momento em que o comando foi executado; depois de alterar o catálogo, execute o comando novamente e envie a pasta `docs/` atualizada.

Não envie `falkcroche.db`, arquivos `.env` nem dados privados de clientes ao repositório. O `.gitignore` do projeto exclui o banco SQLite local e arquivos de ambiente.