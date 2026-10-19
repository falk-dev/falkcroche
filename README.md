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

O painel local fica em `http://127.0.0.1:8080/admin/login` e a vitrine local em `http://127.0.0.1:8080/`. Por padrão, o servidor escuta em `0.0.0.0:8080`; em produção, publique-o atrás de um proxy reverso com HTTPS e não exponha diretamente a porta da aplicação.
As sessões expiram após 12 horas. Se a senha ou a chave não estiver configurada, o painel permanece bloqueado.
O número de WhatsApp deve incluir o código do país e DDD; ele é usado nos botões de contato do catálogo público.

## Deploy em domínio próprio

O app pode ser servido na raiz do domínio, por exemplo `https://falkcroche.com.br`; os links internos usam caminhos relativos ao próprio domínio. Configure `ADMIN_PASSWORD`, `ADMIN_SESSION_KEY` e `WHATSAPP_NUMERO` no ambiente do serviço, e encaminhe o tráfego HTTPS pelo proxy reverso para a porta configurada em `APP_PORT` (8080 por padrão). O proxy deve enviar `X-Forwarded-Proto: https` e substituir qualquer valor recebido do cliente; o cookie de sessão usa essa informação para ativar `Secure`.

O banco SQLite fica em `./falkcroche.db` por padrão. Em hospedagens com sistema de arquivos efêmero, configure `DATABASE_PATH` para um volume persistente; sem isso, reinicializações ou novos deploys podem apagar os dados de pedidos, clientes, produtos e financeiro.

## Pedidos e pagamentos

Ao criar um pedido, informe a data original da encomenda para registrar pedidos antigos. Quando a cliente aceitar o orçamento, avance o pedido para “Aguardando sinal”. Ao avançar dessa etapa, escolha no modal a forma de pagamento e, para Pix parcelado ou cartão, de uma a três parcelas após o sinal e a data de vencimento da primeira parcela. Confirme também que recebeu o sinal e informe a data: o sistema salva o valor acordado com o desconto correspondente, agenda as parcelas e lança o sinal no caixa.
Um pedido de brinde pode ser registrado separadamente pelo checkbox no formulário; ele fica com valor zero, inicia em produção e não cria pagamentos. O vínculo com um pedido principal é opcional, mas ajuda a manter os brindes rastreáveis junto à encomenda.

Cada recebimento fica pendente até ser confirmado na página de pedidos. Ao confirmar, informe a data em que o dinheiro entrou; só então o valor é lançado no financeiro. A entrega pode ser registrada mesmo com parcelas pendentes, e o lembrete de cada parcela pode ser copiado para o WhatsApp. O histórico financeiro aparece da movimentação mais recente para a mais antiga.
Depois de confirmar a última parcela, o pedido passa automaticamente para “Pago” quando ainda estiver aguardando o pagamento final. É possível editar datas e observações do pedido sem alterar valores ou parcelas já acordados. Movimentações financeiras também podem ser editadas; recebimentos vinculados a pedidos mantêm o tipo “Entrada” e atualizam junto a data e o valor do pagamento correspondente.

O sino no painel reúne notificações não lidas para pagamentos vencidos hoje ou antes. Elas permanecem no centro de lembretes até serem marcadas como lidas ou o pagamento ser confirmado.

Para regenerar os arquivos Go depois de editar um arquivo `.templ`:

```sh
go run github.com/a-h/templ/cmd/templ generate
```

## Exportar uma vitrine estática opcional

O exportador lê o banco local e gera somente a página pública e o logo em `docs/`. Ele também carrega o número de WhatsApp do `.env`:

```sh
go run ./cmd/exportar-vitrine
```

No GitHub, abra **Settings > Pages**, escolha **Deploy from a branch**, selecione a branch `main` e a pasta `/docs`. O endereço público será exibido nessa tela. A exportação contém os produtos publicados no momento em que o comando foi executado; depois de alterar o catálogo, execute o comando novamente e envie a pasta `docs/` atualizada.

Não envie `falkcroche.db`, arquivos `.env` nem dados privados de clientes ao repositório. O `.gitignore` do projeto exclui o banco SQLite local e arquivos de ambiente.