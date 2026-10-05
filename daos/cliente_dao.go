package daos

import (
	"database/sql"
	"log"

	"falkcroche/dtos"
	"falkcroche/modelos"
)

type ClienteDAO struct {
	db *sql.DB
}

func NovoClienteDAO(db *sql.DB) *ClienteDAO {
	return &ClienteDAO{db: db}
}

func (dao *ClienteDAO) Inserir(cliente dtos.ClienteInputDTO) (int, error) {
	query := `INSERT INTO clientes (nome, whatsapp, cidade, estado) VALUES (?, ?, ?, ?)`
	res, err := dao.db.Exec(query, cliente.Nome, cliente.Whatsapp, cliente.Cidade, cliente.Estado)
	if err != nil {
		log.Printf("Erro ao inserir cliente: %v", err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (dao *ClienteDAO) Atualizar(id int, cliente dtos.ClienteInputDTO) error {
	resultado, err := dao.db.Exec(`UPDATE clientes SET nome = ?, whatsapp = ?, cidade = ?, estado = ? WHERE id = ?`, cliente.Nome, cliente.Whatsapp, cliente.Cidade, cliente.Estado, id)
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

func (dao *ClienteDAO) BuscarTodos() ([]modelos.Cliente, error) {
	query := `SELECT id, nome, whatsapp, COALESCE(cidade, ''), COALESCE(estado, '') FROM clientes`
	linhas, err := dao.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var clientes []modelos.Cliente
	for linhas.Next() {
		var c modelos.Cliente
		if err := linhas.Scan(&c.ID, &c.Nome, &c.Whatsapp, &c.Cidade, &c.Estado); err != nil {
			return nil, err
		}
		clientes = append(clientes, c)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return clientes, nil
}
