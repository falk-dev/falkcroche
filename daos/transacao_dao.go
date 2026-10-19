package daos

import (
	"database/sql"
	"errors"
	"falkcroche/dtos"
	"falkcroche/modelos"
	"log"
)

var ErrTransacaoNaoEncontrada = errors.New("transação não encontrada")
var ErrTipoTransacaoInvalido = errors.New("recebimentos de pedidos devem permanecer como entrada")

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
	query := `SELECT id, tipo, descricao, valor, data, COALESCE(pedido_id, 0) FROM transacoes ORDER BY data DESC, id DESC`
	linhas, err := dao.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var lista []modelos.Transacao
	for linhas.Next() {
		var t modelos.Transacao
		if err := linhas.Scan(&t.ID, &t.Tipo, &t.Descricao, &t.Valor, &t.Data, &t.PedidoID); err != nil {
			return nil, err
		}
		lista = append(lista, t)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return lista, nil
}

func (dao *TransacaoDAO) Atualizar(id int, t dtos.TransacaoInputDTO) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var pedidoID sql.NullInt64
	var etapa sql.NullString
	if err := tx.QueryRow(`SELECT pedido_id, etapa FROM transacoes WHERE id = ?`, id).Scan(&pedidoID, &etapa); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTransacaoNaoEncontrada
		}
		return err
	}
	if pedidoID.Valid && t.Tipo != "Entrada" {
		return ErrTipoTransacaoInvalido
	}

	resultado, err := tx.Exec(`UPDATE transacoes SET tipo = ?, descricao = ?, valor = ?, data = ? WHERE id = ?`,
		t.Tipo, t.Descricao, t.Valor, t.Data, id)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrTransacaoNaoEncontrada
	}
	if pedidoID.Valid && etapa.Valid {
		resultado, err := tx.Exec(`UPDATE pagamentos SET valor = ?, data_recebimento = ?
			WHERE pedido_id = ? AND etapa = ? AND status = 'Recebido'`,
			t.Valor, t.Data, pedidoID.Int64, etapa.String)
		if err != nil {
			return err
		}
		afetadas, err := resultado.RowsAffected()
		if err != nil {
			return err
		}
		if afetadas == 0 {
			return ErrTransacaoNaoEncontrada
		}
	}
	return tx.Commit()
}
