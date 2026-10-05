package daos

import (
	"database/sql"
	"math"
	"testing"

	"falkcroche/dtos"
	"falkcroche/modelos"

	_ "github.com/mattn/go-sqlite3"
)

func TestAtualizarStatusRegistraParcelasUmaVez(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, query := range []string{
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY, preco REAL NOT NULL)`,
		`CREATE TABLE pedidos (id INTEGER PRIMARY KEY, produto_id INTEGER NOT NULL, status TEXT NOT NULL)`,
		`CREATE TABLE transacoes (id INTEGER PRIMARY KEY, tipo TEXT NOT NULL, descricao TEXT NOT NULL, valor REAL NOT NULL, data TEXT NOT NULL, pedido_id INTEGER, etapa TEXT)`,
		`CREATE UNIQUE INDEX idx_transacoes_pedido_etapa ON transacoes(pedido_id, etapa) WHERE pedido_id IS NOT NULL`,
		`INSERT INTO produtos (id, preco) VALUES (1, 100.01)`,
		`INSERT INTO pedidos (id, produto_id, status) VALUES (1, 1, 'Orçamento')`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}

	dao := NovoPedidoDAO(db)
	if err := dao.AtualizarStatus(1, modelos.StatusSinalRecebido); err != ErrTransicaoPedidoInvalida {
		t.Fatalf("transição direta para sinal recebido deveria ser recusada, recebeu: %v", err)
	}
	for _, status := range []string{
		modelos.StatusAguardandoSinal,
		modelos.StatusSinalRecebido,
		modelos.StatusEmProducao,
		modelos.StatusAguardandoPagamento,
		modelos.StatusPago,
	} {
		if err := dao.AtualizarStatus(1, status); err != nil {
			t.Fatalf("falha ao avançar para %q: %v", status, err)
		}
	}
	if err := dao.AtualizarStatus(1, modelos.StatusSinalRecebido); err != ErrTransicaoPedidoInvalida {
		t.Fatalf("repetir o sinal deveria ser recusado, recebeu: %v", err)
	}

	var quantidade int
	var total float64
	if err := db.QueryRow(`SELECT COUNT(*), SUM(valor) FROM transacoes WHERE pedido_id = 1`).Scan(&quantidade, &total); err != nil {
		t.Fatal(err)
	}
	if quantidade != 2 {
		t.Fatalf("quantidade de lançamentos = %d; esperado 2", quantidade)
	}
	if math.Abs(total-100.01) > 0.001 {
		t.Fatalf("total recebido = %.2f; esperado 100.01", total)
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
		mensagem_orcamento TEXT DEFAULT ''
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
