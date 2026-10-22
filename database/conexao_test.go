package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCriarTabelasMigraPedidosERetemRecebimentosAnteriores(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, query := range []string{
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY, nome_peca TEXT NOT NULL, categoria TEXT NOT NULL DEFAULT '', preco REAL NOT NULL)`,
		`CREATE TABLE pedidos (
			id INTEGER PRIMARY KEY,
			cliente_id INTEGER,
			produto_id INTEGER,
			status TEXT NOT NULL,
			data_pedido TEXT,
			dias_prazo INTEGER,
			data_entrega TEXT,
			observacoes TEXT DEFAULT '',
			mensagem_orcamento TEXT DEFAULT ''
		)`,
		`CREATE TABLE transacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tipo TEXT NOT NULL,
			descricao TEXT NOT NULL,
			valor REAL NOT NULL,
			data TEXT NOT NULL,
			pedido_id INTEGER,
			etapa TEXT
		)`,
		`INSERT INTO produtos (id, nome_peca, preco) VALUES (2, 'Peça antiga', 100)`,
		`INSERT INTO pedidos (id, produto_id, status) VALUES (7, 2, 'Sinal recebido')`,
		`INSERT INTO transacoes (tipo, descricao, valor, data, pedido_id, etapa)
			VALUES ('Entrada', 'Pedido #7 - sinal', 30, '2026-08-01', 7, 'sinal')`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}

	criarTabelas(db)

	var tabelasNotificacoes int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'notificacoes'`).Scan(&tabelasNotificacoes); err != nil {
		t.Fatal(err)
	}
	if tabelasNotificacoes != 1 {
		t.Fatal("a tabela de notificações deveria ser criada na migração")
	}
	var tabelasEtiquetas int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'etiquetas_pedido'`).Scan(&tabelasEtiquetas); err != nil {
		t.Fatal(err)
	}
	if tabelasEtiquetas != 1 {
		t.Fatal("a tabela de etiquetas dos pedidos deveria ser criada na migração")
	}
	var categoria string
	if err := db.QueryRow(`SELECT categoria FROM produtos WHERE id = 2`).Scan(&categoria); err != nil {
		t.Fatal(err)
	}
	if categoria != "" {
		t.Fatalf("categoria migrada = %q; esperada vazia para produto antigo", categoria)
	}
	var quantidadeItens int
	var nomeItem string
	if err := db.QueryRow(`SELECT COUNT(*), MAX(nome_produto) FROM itens_pedido WHERE pedido_id = 7`).Scan(&quantidadeItens, &nomeItem); err != nil {
		t.Fatal(err)
	}
	if quantidadeItens != 1 || nomeItem != "Peça antiga" {
		t.Fatalf("itens migrados = %d, %q; esperado um item Peça antiga", quantidadeItens, nomeItem)
	}

	var quantidade int
	var recebido, pendente float64
	var status, data string
	if err := db.QueryRow(`SELECT COUNT(*), SUM(CASE WHEN status = 'Recebido' THEN valor ELSE 0 END),
		SUM(CASE WHEN status = 'Pendente' THEN valor ELSE 0 END),
		MAX(CASE WHEN status = 'Pendente' THEN status ELSE '' END), MAX(data_recebimento)
		FROM pagamentos WHERE pedido_id = 7`).Scan(&quantidade, &recebido, &pendente, &status, &data); err != nil {
		t.Fatal(err)
	}
	if quantidade != 2 || recebido != 30 || pendente != 70 || status != "Pendente" || data != "2026-08-01" {
		t.Fatalf("pagamentos migrados = (%d, recebido %.2f, pendente %.2f, %q, %q); esperado (2, 30, 70, Pendente, 2026-08-01)",
			quantidade, recebido, pendente, status, data)
	}
	if _, err := db.Exec(`INSERT INTO pedidos (id, produto_id, status, preco_base, brinde)
		VALUES (8, 2, 'Em produção', 0, 1)`); err != nil {
		t.Fatal(err)
	}
	criarTabelas(db)
	var precoBrinde float64
	if err := db.QueryRow(`SELECT preco_base FROM pedidos WHERE id = 8`).Scan(&precoBrinde); err != nil {
		t.Fatal(err)
	}
	if precoBrinde != 0 {
		t.Fatalf("migração alterou o preço zero do brinde para %.2f", precoBrinde)
	}
}
