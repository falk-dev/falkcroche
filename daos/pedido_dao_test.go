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
			preco_base REAL NOT NULL DEFAULT 0,
			desconto_cartao REAL NOT NULL DEFAULT 0,
			desconto_pix REAL NOT NULL DEFAULT 0,
			desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1,
			pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
			forma_pagamento TEXT NOT NULL DEFAULT '',
			total_acordado REAL NOT NULL DEFAULT 0
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
		preco_base REAL NOT NULL DEFAULT 0,
		desconto_cartao REAL NOT NULL DEFAULT 0,
		desconto_pix REAL NOT NULL DEFAULT 0,
		desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1,
		pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
		forma_pagamento TEXT NOT NULL DEFAULT '',
		total_acordado REAL NOT NULL DEFAULT 0
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

	mensagem := "Orçamento individual da linha escolhida"
	dao := NovoPedidoDAO(db)
	id, err := dao.Inserir(dtos.PedidoInputDTO{
		ClienteID:         1,
		ProdutoID:         1,
		Status:            modelos.StatusOrcamento,
		MensagemOrcamento: mensagem,
	})
	if err != nil {
		t.Fatal(err)
	}

	pedidos, err := dao.BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	if len(pedidos) != 1 || pedidos[0].ID != id || pedidos[0].MensagemOrcamento != mensagem {
		t.Fatalf("mensagem não foi recuperada para o pedido: %+v", pedidos)
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
