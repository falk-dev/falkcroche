package daos

import (
	"database/sql"
	"math"
	"testing"
	"time"

	"falkcroche/dtos"
	"falkcroche/modelos"

	_ "github.com/mattn/go-sqlite3"
)

func TestPedidoAgendaERegistraRecebimentosSeparadamente(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, query := range []string{
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY, preco REAL NOT NULL)`,
		`CREATE TABLE pedidos (
			id INTEGER PRIMARY KEY,
			cliente_id INTEGER,
			produto_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			data_pedido TEXT,
			dias_prazo INTEGER,
			data_entrega TEXT,
			observacoes TEXT DEFAULT '',
			mensagem_orcamento TEXT DEFAULT '',
			linha_cores_nome TEXT NOT NULL DEFAULT '',
			linha_cores_link TEXT NOT NULL DEFAULT '',
			preco_base REAL NOT NULL DEFAULT 0,
			desconto_cartao REAL NOT NULL DEFAULT 0,
			desconto_pix REAL NOT NULL DEFAULT 0,
			desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1,
			pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
			forma_pagamento TEXT NOT NULL DEFAULT '',
			total_acordado REAL NOT NULL DEFAULT 0,
			brinde BOOLEAN NOT NULL DEFAULT 0,
			incluir_brinde BOOLEAN NOT NULL DEFAULT 0,
			pedido_principal_id INTEGER
		)`,
		`CREATE TABLE transacoes (id INTEGER PRIMARY KEY, tipo TEXT NOT NULL, descricao TEXT NOT NULL, valor REAL NOT NULL, data TEXT NOT NULL, pedido_id INTEGER, etapa TEXT)`,
		`CREATE UNIQUE INDEX idx_transacoes_pedido_etapa ON transacoes(pedido_id, etapa) WHERE pedido_id IS NOT NULL`,
		`CREATE TABLE pagamentos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pedido_id INTEGER NOT NULL,
			etapa TEXT NOT NULL,
			descricao TEXT NOT NULL,
			valor REAL NOT NULL,
			data_vencimento TEXT NOT NULL,
			data_recebimento TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Pendente',
			UNIQUE (pedido_id, etapa)
		)`,
		`CREATE TABLE notificacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pagamento_id INTEGER NOT NULL UNIQUE,
			lida BOOLEAN NOT NULL DEFAULT 0
		)`,
		`INSERT INTO produtos (id, preco) VALUES (1, 100.01)`,
		`INSERT INTO pedidos (id, produto_id, status, preco_base, desconto_pix, desconto_pix_avista, pix_parcelado) VALUES (1, 1, 'Orçamento', 100.01, 5, 1, 1)`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}

	dao := NovoPedidoDAO(db)
	var quantidade int
	var total float64
	if err := dao.AtualizarStatusComPagamento(1, modelos.StatusSinalRecebido, "", 1, "2026-02-15"); err != ErrTransicaoPedidoInvalida {
		t.Fatalf("transição direta para sinal recebido deveria ser recusada, recebeu: %v", err)
	}
	if err := dao.AtualizarStatus(1, modelos.StatusAguardandoSinal); err != nil {
		t.Fatalf("falha ao avançar para aguardando sinal: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM pagamentos WHERE pedido_id = 1`).Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 0 {
		t.Fatalf("pagamentos devem ser configurados somente ao confirmar o sinal; encontrados: %d", quantidade)
	}
	if err := dao.AvancarStatusComRecebimentoSinal(1, "Pix parcelado", 2, "2026-02-15", "2026-01-20"); err != nil {
		t.Fatalf("falha ao confirmar condições e recebimento do sinal: %v", err)
	}

	if err := db.QueryRow(`SELECT COUNT(*), SUM(valor) FROM pagamentos WHERE pedido_id = 1`).Scan(&quantidade, &total); err != nil {
		t.Fatal(err)
	}
	if quantidade != 3 {
		t.Fatalf("quantidade de pagamentos agendados = %d; esperado 3", quantidade)
	}
	if math.Abs(total-95.01) > 0.001 {
		t.Fatalf("total acordado = %.2f; esperado 95.01", total)
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM transacoes WHERE pedido_id = 1`).Scan(&quantidade); err != nil {
		t.Fatal(err)
	}
	if quantidade != 1 {
		t.Fatalf("lançamentos após confirmar o sinal = %d; esperado 1", quantidade)
	}

	var sinalID int
	if err := db.QueryRow(`SELECT id FROM pagamentos WHERE pedido_id = 1 AND etapa = 'sinal'`).Scan(&sinalID); err != nil {
		t.Fatal(err)
	}
	if err := dao.RegistrarRecebimento(sinalID, "2026-01-21"); err != ErrPagamentoJaRecebido {
		t.Fatalf("confirmar o mesmo pagamento novamente deveria falhar; recebeu: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*), SUM(valor) FROM transacoes WHERE pedido_id = 1`).Scan(&quantidade, &total); err != nil {
		t.Fatal(err)
	}
	if quantidade != 1 || math.Abs(total-28.50) > 0.001 {
		t.Fatalf("lançamentos confirmados = %d, total %.2f; esperado um lançamento de 28.50", quantidade, total)
	}
	var descricao, data string
	if err := db.QueryRow(`SELECT descricao, data FROM transacoes WHERE pedido_id = 1`).Scan(&descricao, &data); err != nil {
		t.Fatal(err)
	}
	if descricao != "Pedido #1 - Sinal (30%) (Pix parcelado)" || data != "2026-01-20" {
		t.Fatalf("lançamento financeiro = (%q, %q); esperado método e data reais do recebimento", descricao, data)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM pedidos WHERE id = 1`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != modelos.StatusSinalRecebido {
		t.Fatalf("status após confirmar sinal = %q; esperado %q", status, modelos.StatusSinalRecebido)
	}

	if err := dao.AtualizarStatus(1, modelos.StatusEmProducao); err != nil {
		t.Fatalf("falha ao avançar para produção: %v", err)
	}
	if err := dao.AtualizarStatus(1, modelos.StatusAguardandoPagamento); err != nil {
		t.Fatalf("falha ao aguardar pagamento final: %v", err)
	}
	var parcela1ID, parcela2ID int
	if err := db.QueryRow(`SELECT id FROM pagamentos WHERE pedido_id = 1 AND etapa = 'parcela 1'`).Scan(&parcela1ID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM pagamentos WHERE pedido_id = 1 AND etapa = 'parcela 2'`).Scan(&parcela2ID); err != nil {
		t.Fatal(err)
	}
	if err := dao.RegistrarRecebimento(parcela1ID, "2026-02-15"); err != nil {
		t.Fatalf("falha ao confirmar a primeira parcela: %v", err)
	}
	if err := db.QueryRow(`SELECT status FROM pedidos WHERE id = 1`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != modelos.StatusAguardandoPagamento {
		t.Fatalf("pedido avançou antes de quitar todas as parcelas: %q", status)
	}
	if err := dao.RegistrarRecebimento(parcela2ID, "2026-03-15"); err != nil {
		t.Fatalf("falha ao confirmar a última parcela: %v", err)
	}
	if err := db.QueryRow(`SELECT status FROM pedidos WHERE id = 1`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != modelos.StatusPago {
		t.Fatalf("status após confirmar a última parcela = %q; esperado %q", status, modelos.StatusPago)
	}

	if _, err := db.Exec(`INSERT INTO pedidos (id, produto_id, status, preco_base, desconto_pix_avista) VALUES (2, 1, 'Orçamento', 100.01, 0)`); err != nil {
		t.Fatal(err)
	}

	if err := dao.AtualizarStatus(2, modelos.StatusAguardandoSinal); err != nil {
		t.Fatal(err)
	}
	if err := dao.AvancarStatusComRecebimentoSinal(2, "Pix à vista", 1, "2026-02-15", "2026-01-20"); err != nil {
		t.Fatalf("falha ao confirmar Pix à vista sem desconto: %v", err)
	}
	if err := db.QueryRow(`SELECT total_acordado FROM pedidos WHERE id = 2`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if math.Abs(total-100.01) > 0.001 {
		t.Fatalf("total acordado no Pix à vista sem desconto = %.2f; esperado 100.01", total)
	}
}

func TestPedidoBrindeZeraPrecoEPermiteVinculoOpcional(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE pedidos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cliente_id INTEGER, produto_id INTEGER, status TEXT NOT NULL,
		data_pedido TEXT, dias_prazo INTEGER, data_entrega TEXT,
		observacoes TEXT, mensagem_orcamento TEXT, linha_cores_nome TEXT NOT NULL DEFAULT '',
		linha_cores_link TEXT NOT NULL DEFAULT '', preco_base REAL NOT NULL DEFAULT 0,
		desconto_cartao REAL NOT NULL DEFAULT 0, desconto_pix REAL NOT NULL DEFAULT 0,
		desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1, pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
		brinde BOOLEAN NOT NULL DEFAULT 0, incluir_brinde BOOLEAN NOT NULL DEFAULT 0,
		pedido_principal_id INTEGER
	);
	CREATE TABLE pagamentos (id INTEGER PRIMARY KEY, pedido_id INTEGER);
	CREATE TABLE itens_pedido (id INTEGER PRIMARY KEY, pedido_id INTEGER, produto_id INTEGER, nome_produto TEXT, categoria TEXT, quantidade INTEGER, preco_unitario REAL)`); err != nil {
		t.Fatal(err)
	}
	dao := NovoPedidoDAO(db)
	principalID, err := dao.Inserir(dtos.PedidoInputDTO{
		ClienteID: 1, ProdutoID: 1, Status: modelos.StatusOrcamento, PrecoBase: 80,
	})
	if err != nil {
		t.Fatal(err)
	}
	brindeID, err := dao.Inserir(dtos.PedidoInputDTO{
		ClienteID: 1, ProdutoID: 2, Status: modelos.StatusOrcamento, PrecoBase: 45,
		DescontoCartao: 10, PixParcelado: true, Brinde: true, PedidoPrincipalID: principalID,
	})
	if err != nil {
		t.Fatalf("falha ao cadastrar brinde relacionado: %v", err)
	}
	var status string
	var preco float64
	var brinde bool
	var principal sql.NullInt64
	if err := db.QueryRow(`SELECT status, preco_base, brinde, pedido_principal_id FROM pedidos WHERE id = ?`, brindeID).
		Scan(&status, &preco, &brinde, &principal); err != nil {
		t.Fatal(err)
	}
	if status != modelos.StatusEmProducao || preco != 0 || !brinde || !principal.Valid || int(principal.Int64) != principalID {
		t.Fatalf("pedido de brinde = (%q, %.2f, %t, %+v); esperava produção, valor zero e vínculo", status, preco, brinde, principal)
	}
	var pagamentos int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pagamentos WHERE pedido_id = ?`, brindeID).Scan(&pagamentos); err != nil {
		t.Fatal(err)
	}
	if pagamentos != 0 {
		t.Fatalf("pedido de brinde gerou %d pagamentos; esperado zero", pagamentos)
	}
	if err := dao.AtualizarStatus(brindeID, modelos.StatusAguardandoPagamento); err != ErrTransicaoPedidoInvalida {
		t.Fatalf("brinde não deveria avançar para pagamento: %v", err)
	}
	if err := dao.AtualizarStatus(brindeID, modelos.StatusEntregue); err != nil {
		t.Fatalf("brinde em produção deveria poder ser entregue sem cobrança: %v", err)
	}
	brindeSemVinculo, err := dao.Inserir(dtos.PedidoInputDTO{ClienteID: 1, ProdutoID: 3, Brinde: true})
	if err != nil {
		t.Fatalf("brinde sem vínculo deveria ser permitido: %v", err)
	}
	if err := db.QueryRow(`SELECT preco_base, pedido_principal_id FROM pedidos WHERE id = ?`, brindeSemVinculo).
		Scan(&preco, &principal); err != nil {
		t.Fatal(err)
	}
	if preco != 0 || principal.Valid {
		t.Fatalf("brinde avulso ficou com valor %.2f ou vínculo %+v", preco, principal)
	}
	if _, err := dao.Inserir(dtos.PedidoInputDTO{ProdutoID: 4, Brinde: true, PedidoPrincipalID: 99}); err != ErrPedidoPrincipalNaoEncontrado {
		t.Fatalf("pedido principal inexistente deveria ser rejeitado, recebeu: %v", err)
	}
	if _, err := dao.Inserir(dtos.PedidoInputDTO{ProdutoID: 4, Brinde: true, PedidoPrincipalID: brindeID}); err != ErrPedidoPrincipalInvalido {
		t.Fatalf("brinde não deveria aceitar outro brinde como principal, recebeu: %v", err)
	}
}

func TestAtualizarDetalhesPedidoPreservaCondicoesAcordadas(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE pedidos (
		id INTEGER PRIMARY KEY, cliente_id INTEGER, produto_id INTEGER, status TEXT, dias_prazo INTEGER,
		data_pedido TEXT, data_entrega TEXT, observacoes TEXT, mensagem_orcamento TEXT,
		linha_cores_nome TEXT, linha_cores_link TEXT, preco_base REAL, desconto_cartao REAL, desconto_pix REAL,
		desconto_pix_avista BOOLEAN, pix_parcelado BOOLEAN, total_acordado REAL, forma_pagamento TEXT,
		brinde BOOLEAN, incluir_brinde BOOLEAN, pedido_principal_id INTEGER
	);
		CREATE TABLE pagamentos (id INTEGER PRIMARY KEY, pedido_id INTEGER);
		CREATE TABLE itens_pedido (id INTEGER PRIMARY KEY, pedido_id INTEGER, produto_id INTEGER, nome_produto TEXT, categoria TEXT, quantidade INTEGER, preco_unitario REAL);
		INSERT INTO pedidos VALUES (1, 1, 1, 'Orçamento', 20, '2026-01-01', '2026-02-01', 'Antes', 'Orçamento antigo', 'Linha', 'https://exemplo.com', 120, 0, 0, 1, 0, 120, 'Pix parcelado', 0, 0, NULL)`); err != nil {
		t.Fatal(err)
	}
	dao := NovoPedidoDAO(db)
	pedidoEdicao := dtos.PedidoEdicaoDTO{
		ClienteID: 2, ProdutoID: 3, DataPedido: "2026-01-02", DiasPrazo: 25,
		DataEntrega: "2026-02-10", Observacoes: "Novo detalhe", MensagemOrcamento: "Novo orçamento",
		LinhaCoresNome: "Nova linha", LinhaCoresLink: "https://novo.exemplo", PrecoBase: 120,
		DescontoPixAvista: true,
	}
	if err := dao.AtualizarDetalhes(1, pedidoEdicao); err != nil {
		t.Fatal(err)
	}
	var dataPedido, dataEntrega, observacoes, mensagemOrcamento, linhaCoresNome, linhaCoresLink, formaPagamento string
	var clienteID, produtoID, diasPrazo int
	var total float64
	if err := db.QueryRow(`SELECT cliente_id, produto_id, data_pedido, dias_prazo, data_entrega, observacoes, mensagem_orcamento, linha_cores_nome, linha_cores_link, total_acordado, forma_pagamento FROM pedidos WHERE id = 1`).
		Scan(&clienteID, &produtoID, &dataPedido, &diasPrazo, &dataEntrega, &observacoes, &mensagemOrcamento, &linhaCoresNome, &linhaCoresLink, &total, &formaPagamento); err != nil {
		t.Fatal(err)
	}
	if clienteID != 2 || produtoID != 3 || dataPedido != "2026-01-02" || diasPrazo != 25 ||
		dataEntrega != "2026-02-10" || observacoes != "Novo detalhe" || mensagemOrcamento != "Novo orçamento" ||
		linhaCoresNome != "Nova linha" || linhaCoresLink != "https://novo.exemplo" ||
		total != 120 || formaPagamento != "Pix parcelado" {
		t.Fatalf("edição alterou dados incorretos do pedido: %d, %d, %q, %d, %q, %q, %q, %q, %q, %.2f, %q",
			clienteID, produtoID, dataPedido, diasPrazo, dataEntrega, observacoes, mensagemOrcamento, linhaCoresNome, linhaCoresLink, total, formaPagamento)
	}
	if err := dao.AtualizarDetalhes(2, pedidoEdicao); err != ErrPedidoNaoEncontrado {
		t.Fatalf("pedido inexistente deveria ser reportado, recebeu: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO pagamentos (pedido_id) VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	condicoesAlteradas := pedidoEdicao
	condicoesAlteradas.PrecoBase = 130
	if err := dao.AtualizarDetalhes(1, condicoesAlteradas); err != ErrCondicoesPedidoBloqueadas {
		t.Fatalf("condições financeiras com pagamentos deveriam ser bloqueadas, recebeu: %v", err)
	}
}

func TestPedidoPersisteMensagemOrcamento(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`CREATE TABLE pedidos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cliente_id INTEGER,
		produto_id INTEGER,
		status TEXT NOT NULL,
		data_pedido TEXT,
		dias_prazo INTEGER,
		data_entrega TEXT,
		observacoes TEXT DEFAULT '',
		mensagem_orcamento TEXT DEFAULT '',
		linha_cores_nome TEXT NOT NULL DEFAULT '',
		linha_cores_link TEXT NOT NULL DEFAULT '',
		preco_base REAL NOT NULL DEFAULT 0,
		desconto_cartao REAL NOT NULL DEFAULT 0,
		desconto_pix REAL NOT NULL DEFAULT 0,
		desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1,
		pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
		forma_pagamento TEXT NOT NULL DEFAULT '',
		total_acordado REAL NOT NULL DEFAULT 0,
		brinde BOOLEAN NOT NULL DEFAULT 0,
		incluir_brinde BOOLEAN NOT NULL DEFAULT 0,
		pedido_principal_id INTEGER
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE pagamentos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pedido_id INTEGER NOT NULL,
		etapa TEXT NOT NULL,
		descricao TEXT NOT NULL,
		valor REAL NOT NULL,
		data_vencimento TEXT NOT NULL,
		data_recebimento TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'Pendente',
		UNIQUE (pedido_id, etapa)
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE etiquetas_pedido (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pedido_id INTEGER NOT NULL,
		texto TEXT NOT NULL,
		cor TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE itens_pedido (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pedido_id INTEGER NOT NULL,
		produto_id INTEGER NOT NULL,
		nome_produto TEXT NOT NULL,
		categoria TEXT NOT NULL DEFAULT '',
		quantidade INTEGER NOT NULL DEFAULT 1,
		preco_unitario REAL NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatal(err)
	}

	mensagem := "Orçamento individual da linha escolhida"
	dao := NovoPedidoDAO(db)
	id, err := dao.Inserir(dtos.PedidoInputDTO{
		ClienteID: 1,
		ProdutoID: 1,
		Itens: []dtos.ItemPedidoInputDTO{
			{ProdutoID: 1, NomeProduto: "Boina", Categoria: "Acessórios", Quantidade: 2, PrecoUnitario: 30},
			{ProdutoID: 2, NomeProduto: "Bolsa", Categoria: "Bolsas", Quantidade: 1, PrecoUnitario: 20},
		},
		Status:            modelos.StatusOrcamento,
		MensagemOrcamento: mensagem,
		LinhaCoresNome:    "Linha de teste",
		LinhaCoresLink:    "https://exemplo.com/cores",
		IncluirBrinde:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := dao.AdicionarEtiquetaPedido(id, "Presente", "rosa"); err != nil {
		t.Fatal(err)
	}

	pedidos, err := dao.BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	if len(pedidos) != 1 || pedidos[0].ID != id || pedidos[0].MensagemOrcamento != mensagem ||
		pedidos[0].LinhaCoresNome != "Linha de teste" || pedidos[0].LinhaCoresLink != "https://exemplo.com/cores" || !pedidos[0].IncluirBrinde {
		t.Fatalf("mensagem não foi recuperada para o pedido: %+v", pedidos)
	}
	if len(pedidos[0].Itens) != 2 || pedidos[0].PrecoBase != 80 || pedidos[0].Itens[0].Quantidade != 2 || pedidos[0].Itens[1].NomeProduto != "Bolsa" {
		t.Fatalf("itens ou total do pedido foram recuperados incorretamente: %+v", pedidos[0])
	}
	if len(pedidos[0].Etiquetas) != 1 || pedidos[0].Etiquetas[0].Texto != "Presente" || pedidos[0].Etiquetas[0].Cor != "rosa" {
		t.Fatalf("etiqueta não foi recuperada para o pedido: %+v", pedidos[0].Etiquetas)
	}
	if err := dao.ExcluirEtiquetaPedido(id, pedidos[0].Etiquetas[0].ID); err != nil {
		t.Fatalf("não foi possível excluir a etiqueta: %v", err)
	}
	if err := dao.ExcluirEtiquetaPedido(id, pedidos[0].Etiquetas[0].ID); err != ErrEtiquetaPedidoNaoEncontrada {
		t.Fatalf("esperava etiqueta não encontrada após exclusão, recebeu: %v", err)
	}
}

func TestAdicionarMesesMantemDataValidaNoFimDoMes(t *testing.T) {
	inicio := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	resultado := adicionarMeses(inicio, 1)
	esperado := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)
	if !resultado.Equal(esperado) {
		t.Fatalf("próximo vencimento = %s; esperado %s", resultado.Format("2006-01-02"), esperado.Format("2006-01-02"))
	}
}
