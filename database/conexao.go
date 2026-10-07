package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // o underline importa o driver apenas para os seus efeitos colaterais
)

// Conectar() inicia a conexao com o banco e cria as tabelas se nao existirem
func Conectar() *sql.DB {
	caminho := os.Getenv("DATABASE_PATH")
	if caminho == "" {
		caminho = "./falkcroche.db"
	}
	db, err := sql.Open("sqlite3", caminho)
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}

	// testa a conexao para garantir que está funcionando
	if err = db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// garante que as tabelas existam
	criarTabelas(db)

	return db
}

// criarTabelas() cria as tabelas necessarias para o sistema caso nao existam
func criarTabelas(db *sql.DB) {
	// query para criar a tabela de clientes
	tabelaClientes := `
	CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT NOT NULL,
		whatsapp TEXT,
		cidade TEXT DEFAULT '',
		estado TEXT DEFAULT ''
	);`

	// query para criar a tabela de produtos
	tabelaProdutos := `
	CREATE TABLE IF NOT EXISTS produtos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome_peca TEXT NOT NULL,
		preco REAL NOT NULL,
		foto_url TEXT,
		pronta_entrega BOOLEAN NOT NULL DEFAULT 0,
		publicado BOOLEAN NOT NULL DEFAULT 1
	);`

	// query para criar a tabela de pedidos
	tabelaPedidos := `
	CREATE TABLE IF NOT EXISTS pedidos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cliente_id INTEGER,
		produto_id INTEGER,
		status TEXT NOT NULL,
		data_pedido TEXT,
		dias_prazo INTEGER,
		data_entrega TEXT,
		observacoes TEXT DEFAULT '',
		mensagem_orcamento TEXT DEFAULT '',
		FOREIGN KEY (cliente_id) REFERENCES clientes(id),
		FOREIGN KEY (produto_id) REFERENCES produtos(id)
	);`

	// query para criar a tabela de transacoes
	tabelaTransacoes := `
	CREATE TABLE IF NOT EXISTS transacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo TEXT NOT NULL,
		descricao TEXT NOT NULL,
		valor REAL NOT NULL,
		data TEXT NOT NULL,
		pedido_id INTEGER,
		etapa TEXT
	);`

	// executa as queries de criacao
	if _, err := db.Exec(tabelaClientes); err != nil {
		log.Fatalf("Erro ao criar tabela clientes: %v", err)
	}
	if _, err := db.Exec(tabelaProdutos); err != nil {
		log.Fatalf("Erro ao criar tabela produtos: %v", err)
	}
	if _, err := db.Exec(tabelaPedidos); err != nil {
		log.Fatalf("Erro ao criar tabela pedidos: %v", err)
	}
	if _, err := db.Exec(tabelaTransacoes); err != nil {
		log.Fatalf("Erro ao criar tabela transacoes: %v", err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_transacoes_pedido_etapa ON transacoes(pedido_id, etapa) WHERE pedido_id IS NOT NULL`); err != nil {
		log.Fatalf("Erro ao criar índice de recebimentos do pedido: %v", err)
	}

	log.Println("Tabelas verificadas/criadas com sucesso!")
}
