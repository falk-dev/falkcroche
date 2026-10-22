package daos

import (
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestLembretesDeVencimentoFicamNaoLidosAteSeremMarcados(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, query := range []string{
		`CREATE TABLE clientes (id INTEGER PRIMARY KEY, nome TEXT NOT NULL)`,
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY, nome_peca TEXT NOT NULL)`,
		`CREATE TABLE pedidos (id INTEGER PRIMARY KEY, cliente_id INTEGER, produto_id INTEGER, status TEXT)`,
		`CREATE TABLE itens_pedido (pedido_id INTEGER, nome_produto TEXT, quantidade INTEGER)`,
		`CREATE TABLE pagamentos (
			id INTEGER PRIMARY KEY,
			pedido_id INTEGER NOT NULL,
			descricao TEXT NOT NULL,
			valor REAL NOT NULL,
			data_vencimento TEXT NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE notificacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pagamento_id INTEGER NOT NULL UNIQUE,
			lida BOOLEAN NOT NULL DEFAULT 0
		)`,
		`INSERT INTO clientes (id, nome) VALUES (1, 'Camila')`,
		`INSERT INTO produtos (id, nome_peca) VALUES (1, 'Amigurumi Coelho')`,
		`INSERT INTO itens_pedido (pedido_id, nome_produto, quantidade) VALUES (12, 'Amigurumi Coelho', 1), (12, 'Touca', 2)`,
		`INSERT INTO pedidos (id, cliente_id, produto_id, status)
			VALUES (12, 1, 1, 'Em produção'), (13, 1, 1, 'Cancelado')`,
		`INSERT INTO pagamentos (id, pedido_id, descricao, valor, data_vencimento, status)
			VALUES (5, 12, 'Sinal', 30, '2026-10-14', 'Pendente'),
				(6, 12, 'Parcela 1', 35, '2026-10-15', 'Pendente'),
				(7, 12, 'Parcela 2', 35, '2026-10-16', 'Pendente'),
				(8, 13, 'Parcela cancelada', 35, '2026-10-14', 'Pendente')`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}

	dao := NovoNotificacaoDAO(db)
	notificacoes, err := dao.ListarVencidas("2026-10-14")
	if err != nil {
		t.Fatal(err)
	}
	if len(notificacoes) != 1 || notificacoes[0].PagamentoID != 5 {
		t.Fatalf("lembretes no vencimento = %+v; esperado apenas pagamento 5", notificacoes)
	}
	notificacoes, err = dao.ListarVencidas("2026-10-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(notificacoes) != 2 || notificacoes[1].Cliente != "Camila" || notificacoes[1].Produto != "Amigurumi Coelho, 2x Touca" {
		t.Fatalf("lembretes não lidos após novo vencimento = %+v; esperado dois com dados do pedido", notificacoes)
	}
	if err := dao.MarcarComoLida(notificacoes[0].ID); err != nil {
		t.Fatal(err)
	}
	notificacoes, err = dao.ListarVencidas("2026-10-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(notificacoes) != 1 || notificacoes[0].PagamentoID != 6 {
		t.Fatalf("lembretes restantes após marcar como lido = %+v; esperado pagamento 6", notificacoes)
	}
	if err := dao.MarcarComoLida(notificacoes[0].ID + 100); !errors.Is(err, ErrNotificacaoNaoEncontrada) {
		t.Fatalf("marcar notificação inexistente deveria retornar erro; recebeu %v", err)
	}
}
