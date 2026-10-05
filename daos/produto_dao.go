package daos

import (
	"database/sql"
	"log"

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
	query := `INSERT INTO produtos (nome_peca, preco, foto_url, pronta_entrega, publicado) VALUES (?, ?, ?, ?, ?)`

	// db.Exec executa a query sem retornar linhas (ideal para INSERT, UPDATE, DELETE)
	resultado, err := dao.db.Exec(query, produto.NomePeca, produto.Preco, produto.FotoURL, produto.ProntaEntrega, produto.Publicado)
	if err != nil {
		log.Printf("Erro ao inserir produto: %v", err)
		return 0, err
	}

	// Pega o ID gerado automaticamente pelo SQLite
	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (dao *ProdutoDAO) Atualizar(id int, produto dtos.ProdutoInputDTO) error {
	resultado, err := dao.db.Exec(`UPDATE produtos SET nome_peca = ?, preco = ?, foto_url = ?, pronta_entrega = ?, publicado = ? WHERE id = ?`, produto.NomePeca, produto.Preco, produto.FotoURL, produto.ProntaEntrega, produto.Publicado, id)
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
	return dao.buscar(`SELECT id, nome_peca, preco, foto_url, pronta_entrega, publicado FROM produtos ORDER BY id DESC`)
}

func (dao *ProdutoDAO) BuscarPublicados() ([]modelos.Produto, error) {
	return dao.buscar(`SELECT id, nome_peca, preco, foto_url, pronta_entrega, publicado FROM produtos WHERE publicado = 1 ORDER BY id DESC`)
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
		err := linhas.Scan(&p.ID, &p.NomePeca, &p.Preco, &p.FotoURL, &p.ProntaEntrega, &p.Publicado)
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
