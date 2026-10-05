package daos

import (
	"database/sql"
	"falkcroche/dtos"
	"falkcroche/modelos"
	"log"
)

type TransacaoDAO struct {
	db *sql.DB
}

func NovoTransacaoDAO(db *sql.DB) *TransacaoDAO {
	return &TransacaoDAO{db: db}
}

func (dao *TransacaoDAO) Inserir(t dtos.TransacaoInputDTO) (int, error) {
	query := `INSERT INTO transacoes (tipo, descricao, valor, data) VALUES (?, ?, ?, ?)`
	res, err := dao.db.Exec(query, t.Tipo, t.Descricao, t.Valor, t.Data)
	if err != nil {
		log.Printf("Erro ao inserir transação: %v", err)
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (dao *TransacaoDAO) BuscarTodas() ([]modelos.Transacao, error) {
	query := `SELECT id, tipo, descricao, valor, data FROM transacoes ORDER BY data DESC`
	linhas, err := dao.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var lista []modelos.Transacao
	for linhas.Next() {
		var t modelos.Transacao
		if err := linhas.Scan(&t.ID, &t.Tipo, &t.Descricao, &t.Valor, &t.Data); err == nil {
			lista = append(lista, t)
		}
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return lista, nil
}
