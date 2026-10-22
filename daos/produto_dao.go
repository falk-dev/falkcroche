package daos

import (
	"database/sql"
	"log"
	"strings"

	"falkcroche/dtos"
	"falkcroche/modelos"
)

// ProdutoDAO é a estrutura que acessa o banco de dados para a entidade Produto.
type ProdutoDAO struct {
	db *sql.DB
}

// NovoProdutoDAO cria uma nova instância de ProdutoDAO recebendo a conexão.
func NovoProdutoDAO(db *sql.DB) *ProdutoDAO {
	return &ProdutoDAO{db: db}
}

// Inserir adiciona um novo produto usando os dados do DTO e retorna o ID criado.
func (dao *ProdutoDAO) Inserir(produto dtos.ProdutoInputDTO) (int, error) {
	tx, err := dao.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	query := `INSERT INTO produtos (nome_peca, categoria, preco, foto_url, pronta_entrega, publicado, imagem_ilustrativa) VALUES (?, ?, ?, ?, ?, ?, ?)`

	// db.Exec executa a query sem retornar linhas (ideal para INSERT, UPDATE, DELETE)
	resultado, err := tx.Exec(query, produto.NomePeca, produto.Categoria, produto.Preco, produto.FotoURL, produto.ProntaEntrega, produto.Publicado, produto.ImagemIlustrativa)
	if err != nil {
		log.Printf("Erro ao inserir produto: %v", err)
		return 0, err
	}

	// Pega o ID gerado automaticamente pelo SQLite
	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := garantirCategoria(tx, produto.Categoria); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(id), nil
}

func (dao *ProdutoDAO) Atualizar(id int, produto dtos.ProdutoInputDTO) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	resultado, err := tx.Exec(`UPDATE produtos SET nome_peca = ?, categoria = ?, preco = ?, foto_url = ?, pronta_entrega = ?, publicado = ?, imagem_ilustrativa = ? WHERE id = ?`, produto.NomePeca, produto.Categoria, produto.Preco, produto.FotoURL, produto.ProntaEntrega, produto.Publicado, produto.ImagemIlustrativa, id)
	if err != nil {
		return err
	}
	quantidade, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if quantidade == 0 {
		return sql.ErrNoRows
	}
	if err := garantirCategoria(tx, produto.Categoria); err != nil {
		return err
	}
	return tx.Commit()
}

func (dao *ProdutoDAO) BuscarCategorias() ([]modelos.CategoriaProduto, error) {
	linhas, err := dao.db.Query(`SELECT nome, prioridade FROM categorias_produtos ORDER BY nome COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var categorias []modelos.CategoriaProduto
	for linhas.Next() {
		var categoria modelos.CategoriaProduto
		if err := linhas.Scan(&categoria.Nome, &categoria.Prioridade); err != nil {
			return nil, err
		}
		categorias = append(categorias, categoria)
	}
	return categorias, linhas.Err()
}

func (dao *ProdutoDAO) CriarCategoria(nome string, prioridade string) error {
	nome = strings.TrimSpace(nome)
	prioridade = normalizarPrioridadeCategoria(prioridade)
	_, err := dao.db.Exec(`INSERT INTO categorias_produtos (nome, prioridade) VALUES (?, ?)`, nome, prioridade)
	return err
}

func (dao *ProdutoDAO) AtualizarPrioridadeCategoria(nome string, prioridade string) error {
	nome = strings.TrimSpace(nome)
	prioridade = normalizarPrioridadeCategoria(prioridade)
	resultado, err := dao.db.Exec(`UPDATE categorias_produtos SET prioridade = ? WHERE nome = ?`, prioridade, nome)
	if err != nil {
		return err
	}
	quantidade, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if quantidade == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// BuscarTodos retorna todos os produtos cadastrados.
func (dao *ProdutoDAO) BuscarTodos() ([]modelos.Produto, error) {
	return dao.buscar(`SELECT id, nome_peca, categoria, COALESCE((SELECT prioridade FROM categorias_produtos WHERE nome = TRIM(produtos.categoria)), 'media'), preco, foto_url, pronta_entrega, publicado, imagem_ilustrativa FROM produtos ORDER BY CASE WHEN TRIM(categoria) = '' THEN 1 ELSE 0 END, categoria COLLATE NOCASE, nome_peca COLLATE NOCASE, id DESC`)
}

func (dao *ProdutoDAO) BuscarPublicados() ([]modelos.Produto, error) {
	return dao.buscar(`SELECT id, nome_peca, categoria, COALESCE((SELECT prioridade FROM categorias_produtos WHERE nome = TRIM(produtos.categoria)), 'media'), preco, foto_url, pronta_entrega, publicado, imagem_ilustrativa FROM produtos WHERE publicado = 1 ORDER BY CASE WHEN TRIM(categoria) = '' THEN 1 ELSE 0 END, categoria COLLATE NOCASE, nome_peca COLLATE NOCASE, id DESC`)
}

func (dao *ProdutoDAO) buscar(query string) ([]modelos.Produto, error) {
	linhas, err := dao.db.Query(query)
	if err != nil {
		log.Printf("Erro ao buscar produtos: %v", err)
		return nil, err
	}
	defer linhas.Close() // Muito importante fechar as linhas após o uso!

	var produtos []modelos.Produto

	// Itera sobre os resultados
	for linhas.Next() {
		var p modelos.Produto
		// Scan copia os valores das colunas para os campos da struct
		err := linhas.Scan(&p.ID, &p.NomePeca, &p.Categoria, &p.PrioridadeCategoria, &p.Preco, &p.FotoURL, &p.ProntaEntrega, &p.Publicado, &p.ImagemIlustrativa)
		if err != nil {
			log.Printf("Erro ao ler produto: %v", err)
			continue
		}
		produtos = append(produtos, p)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return produtos, nil
}

func normalizarPrioridadeCategoria(prioridade string) string {
	prioridade = strings.ToLower(strings.TrimSpace(prioridade))
	switch prioridade {
	case "alta", "media", "baixa":
		return prioridade
	default:
		return "media"
	}
}

func garantirCategoria(tx *sql.Tx, categoria string) error {
	categoria = strings.TrimSpace(categoria)
	if categoria == "" {
		return nil
	}
	_, err := tx.Exec(`INSERT OR IGNORE INTO categorias_produtos (nome, prioridade) VALUES (?, 'media')`, categoria)
	return err
}
