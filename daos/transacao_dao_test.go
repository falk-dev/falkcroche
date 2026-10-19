package daos

import (
	"database/sql"
	"errors"
	"testing"

	"falkcroche/dtos"

	_ "github.com/mattn/go-sqlite3"
)

func TestAtualizarTransacaoSincronizaRecebimentoDoPedido(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, query := range []string{
		`CREATE TABLE transacoes (
			id INTEGER PRIMARY KEY, tipo TEXT NOT NULL, descricao TEXT NOT NULL, valor REAL NOT NULL,
			data TEXT NOT NULL, pedido_id INTEGER, etapa TEXT
		)`,
		`CREATE TABLE pagamentos (
			id INTEGER PRIMARY KEY, pedido_id INTEGER, etapa TEXT, descricao TEXT, valor REAL,
			data_vencimento TEXT, data_recebimento TEXT, status TEXT
		)`,
		`INSERT INTO transacoes VALUES (1, 'Saída', 'Material', 20, '2026-01-01', NULL, NULL)`,
		`INSERT INTO transacoes VALUES (2, 'Entrada', 'Pedido #3 - Parcela 1 (Pix parcelado)', 30, '2026-01-02', 3, 'parcela 1')`,
		`INSERT INTO pagamentos VALUES (1, 3, 'parcela 1', 'Parcela 1 de 2', 30, '2026-01-02', '2026-01-02', 'Recebido')`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	dao := NovoTransacaoDAO(db)
	if err := dao.Atualizar(1, dtos.TransacaoInputDTO{
		Tipo: "Entrada", Descricao: "Reembolso", Valor: 42.50, Data: "2026-01-05",
	}); err != nil {
		t.Fatalf("falha ao editar movimentação manual: %v", err)
	}
	if err := dao.Atualizar(2, dtos.TransacaoInputDTO{
		Tipo: "Entrada", Descricao: "Parcela corrigida", Valor: 35.25, Data: "2026-01-06",
	}); err != nil {
		t.Fatalf("falha ao editar recebimento de pedido: %v", err)
	}
	var valor float64
	var data string
	if err := db.QueryRow(`SELECT valor, data_recebimento FROM pagamentos WHERE id = 1`).Scan(&valor, &data); err != nil {
		t.Fatal(err)
	}
	if valor != 35.25 || data != "2026-01-06" {
		t.Fatalf("pagamento vinculado = %.2f em %s; esperava 35.25 em 2026-01-06", valor, data)
	}
	if err := dao.Atualizar(2, dtos.TransacaoInputDTO{
		Tipo: "Saída", Descricao: "Inconsistente", Valor: 35.25, Data: "2026-01-06",
	}); !errors.Is(err, ErrTipoTransacaoInvalido) {
		t.Fatalf("alterar recebimento vinculado para saída deveria falhar, recebeu: %v", err)
	}
	if err := dao.Atualizar(99, dtos.TransacaoInputDTO{
		Tipo: "Entrada", Descricao: "Não existe", Valor: 1, Data: "2026-01-06",
	}); !errors.Is(err, ErrTransacaoNaoEncontrada) {
		t.Fatalf("movimentação inexistente deveria falhar, recebeu: %v", err)
	}
	lista, err := dao.BuscarTodas()
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) != 2 || lista[0].PedidoID != 3 || lista[1].PedidoID != 0 {
		t.Fatalf("movimentações não foram carregadas com seus vínculos: %+v", lista)
	}
}
