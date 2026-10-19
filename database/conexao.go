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
		preco_base REAL NOT NULL DEFAULT 0,
		desconto_cartao REAL NOT NULL DEFAULT 0,
		desconto_pix REAL NOT NULL DEFAULT 0,
		desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1,
		pix_parcelado BOOLEAN NOT NULL DEFAULT 0,
		forma_pagamento TEXT NOT NULL DEFAULT '',
		total_acordado REAL NOT NULL DEFAULT 0,
		brinde BOOLEAN NOT NULL DEFAULT 0,
		pedido_principal_id INTEGER,
		FOREIGN KEY (cliente_id) REFERENCES clientes(id),
		FOREIGN KEY (produto_id) REFERENCES produtos(id),
		FOREIGN KEY (pedido_principal_id) REFERENCES pedidos(id)
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

	tabelaPagamentos := `
	CREATE TABLE IF NOT EXISTS pagamentos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pedido_id INTEGER NOT NULL,
		etapa TEXT NOT NULL,
		descricao TEXT NOT NULL,
		valor REAL NOT NULL,
		data_vencimento TEXT NOT NULL,
		data_recebimento TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'Pendente',
		FOREIGN KEY (pedido_id) REFERENCES pedidos(id),
		UNIQUE (pedido_id, etapa)
	);`

	tabelaNotificacoes := `
	CREATE TABLE IF NOT EXISTS notificacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pagamento_id INTEGER NOT NULL UNIQUE,
		lida BOOLEAN NOT NULL DEFAULT 0,
		FOREIGN KEY (pagamento_id) REFERENCES pagamentos(id)
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
	if _, err := db.Exec(tabelaPagamentos); err != nil {
		log.Fatalf("Erro ao criar tabela pagamentos: %v", err)
	}
	if _, err := db.Exec(tabelaNotificacoes); err != nil {
		log.Fatalf("Erro ao criar tabela notificacoes: %v", err)
	}
	migrarColunasPedidos(db)
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_transacoes_pedido_etapa ON transacoes(pedido_id, etapa) WHERE pedido_id IS NOT NULL`); err != nil {
		log.Fatalf("Erro ao criar índice de recebimentos do pedido: %v", err)
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento, data_recebimento, status)
		SELECT pedido_id, etapa, descricao, valor, data, data, 'Recebido'
		FROM transacoes WHERE pedido_id IS NOT NULL AND etapa IS NOT NULL AND tipo = 'Entrada'`); err != nil {
		log.Fatalf("Erro ao migrar recebimentos existentes: %v", err)
	}
	migrarPagamentosPendentes(db)

	log.Println("Tabelas verificadas/criadas com sucesso!")
}

func migrarPagamentosPendentes(db *sql.DB) {
	if _, err := db.Exec(`UPDATE pedidos
		SET preco_base = (SELECT preco FROM produtos WHERE produtos.id = pedidos.produto_id)
		WHERE preco_base = 0 AND brinde = 0 AND EXISTS (SELECT 1 FROM produtos WHERE produtos.id = pedidos.produto_id)`); err != nil {
		log.Fatalf("Erro ao recuperar valores dos pedidos existentes: %v", err)
	}
	if _, err := db.Exec(`UPDATE pedidos SET total_acordado = preco_base, forma_pagamento = 'Condições antigas'
		WHERE total_acordado = 0 AND status IN ('Aguardando sinal', 'Sinal recebido', 'Em produção', 'Aguardando pagamento final')
		AND preco_base > 0`); err != nil {
		log.Fatalf("Erro ao definir valores de pedidos existentes: %v", err)
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento)
		SELECT id, 'sinal', 'Sinal (30%)', ROUND(total_acordado * 0.30, 2), date('now')
		FROM pedidos WHERE status = 'Aguardando sinal' AND total_acordado > 0
		AND NOT EXISTS (SELECT 1 FROM pagamentos WHERE pagamentos.pedido_id = pedidos.id AND pagamentos.etapa = 'sinal')`); err != nil {
		log.Fatalf("Erro ao migrar sinais pendentes: %v", err)
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento)
		SELECT pedidos.id, 'saldo-migrado', 'Saldo pendente', ROUND(
			MAX(pedidos.total_acordado - COALESCE((SELECT SUM(valor) FROM pagamentos WHERE pagamentos.pedido_id = pedidos.id), 0), 0), 2), date('now')
		FROM pedidos WHERE pedidos.status IN ('Aguardando sinal', 'Sinal recebido', 'Em produção', 'Aguardando pagamento final')
		AND pedidos.total_acordado > COALESCE((SELECT SUM(valor) FROM pagamentos WHERE pagamentos.pedido_id = pedidos.id), 0)`); err != nil {
		log.Fatalf("Erro ao migrar saldos pendentes: %v", err)
	}
}

func migrarColunasPedidos(db *sql.DB) {
	colunas := map[string]string{
		"preco_base":          `ALTER TABLE pedidos ADD COLUMN preco_base REAL NOT NULL DEFAULT 0`,
		"desconto_cartao":     `ALTER TABLE pedidos ADD COLUMN desconto_cartao REAL NOT NULL DEFAULT 0`,
		"desconto_pix":        `ALTER TABLE pedidos ADD COLUMN desconto_pix REAL NOT NULL DEFAULT 0`,
		"desconto_pix_avista": `ALTER TABLE pedidos ADD COLUMN desconto_pix_avista BOOLEAN NOT NULL DEFAULT 1`,
		"pix_parcelado":       `ALTER TABLE pedidos ADD COLUMN pix_parcelado BOOLEAN NOT NULL DEFAULT 0`,
		"forma_pagamento":     `ALTER TABLE pedidos ADD COLUMN forma_pagamento TEXT NOT NULL DEFAULT ''`,
		"total_acordado":      `ALTER TABLE pedidos ADD COLUMN total_acordado REAL NOT NULL DEFAULT 0`,
		"brinde":              `ALTER TABLE pedidos ADD COLUMN brinde BOOLEAN NOT NULL DEFAULT 0`,
		"pedido_principal_id": `ALTER TABLE pedidos ADD COLUMN pedido_principal_id INTEGER REFERENCES pedidos(id)`,
	}
	for coluna, alteracao := range colunas {
		var encontrada bool
		linhas, err := db.Query(`PRAGMA table_info(pedidos)`)
		if err != nil {
			log.Fatalf("Erro ao verificar coluna %s de pedidos: %v", coluna, err)
		}
		for linhas.Next() {
			var cid int
			var nome, tipo string
			var notNull, primaryKey int
			var valorPadrao sql.NullString
			if err := linhas.Scan(&cid, &nome, &tipo, &notNull, &valorPadrao, &primaryKey); err != nil {
				linhas.Close()
				log.Fatalf("Erro ao ler colunas de pedidos: %v", err)
			}
			if nome == coluna {
				encontrada = true
				break
			}
		}
		if err := linhas.Close(); err != nil {
			log.Fatalf("Erro ao fechar verificação de colunas de pedidos: %v", err)
		}
		if !encontrada {
			if _, err := db.Exec(alteracao); err != nil {
				log.Fatalf("Erro ao migrar coluna %s de pedidos: %v", coluna, err)
			}
		}
	}
}
