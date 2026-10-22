package daos

import (
	"database/sql"
	"errors"

	"falkcroche/modelos"
)

var ErrNotificacaoNaoEncontrada = errors.New("notificação não encontrada")

type NotificacaoDAO struct {
	db *sql.DB
}

func NovoNotificacaoDAO(db *sql.DB) *NotificacaoDAO {
	return &NotificacaoDAO{db: db}
}

func (dao *NotificacaoDAO) ListarVencidas(dataHoje string) ([]modelos.Notificacao, error) {
	if _, err := dao.db.Exec(`INSERT OR IGNORE INTO notificacoes (pagamento_id)
		SELECT pagamentos.id FROM pagamentos
		JOIN pedidos ON pedidos.id = pagamentos.pedido_id
		WHERE pagamentos.status = 'Pendente' AND pagamentos.data_vencimento <= ?
			AND pedidos.status <> 'Cancelado'`, dataHoje); err != nil {
		return nil, err
	}

	linhas, err := dao.db.Query(`SELECT notificacoes.id, pagamentos.id, pedidos.id, clientes.nome,
		COALESCE((SELECT group_concat(CASE WHEN quantidade > 1 THEN quantidade || 'x ' || nome_produto ELSE nome_produto END, ', ')
			FROM itens_pedido WHERE itens_pedido.pedido_id = pedidos.id), produtos.nome_peca),
		pagamentos.descricao, pagamentos.valor, pagamentos.data_vencimento
		FROM notificacoes
		JOIN pagamentos ON pagamentos.id = notificacoes.pagamento_id
		JOIN pedidos ON pedidos.id = pagamentos.pedido_id
		JOIN clientes ON clientes.id = pedidos.cliente_id
		JOIN produtos ON produtos.id = pedidos.produto_id
		WHERE notificacoes.lida = 0 AND pagamentos.status = 'Pendente' AND pedidos.status <> 'Cancelado'
		ORDER BY pagamentos.data_vencimento ASC, notificacoes.id ASC`)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	notificacoes := make([]modelos.Notificacao, 0)
	for linhas.Next() {
		var notificacao modelos.Notificacao
		if err := linhas.Scan(&notificacao.ID, &notificacao.PagamentoID, &notificacao.PedidoID,
			&notificacao.Cliente, &notificacao.Produto, &notificacao.Descricao,
			&notificacao.Valor, &notificacao.DataVencimento); err != nil {
			return nil, err
		}
		notificacoes = append(notificacoes, notificacao)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return notificacoes, nil
}

func (dao *NotificacaoDAO) MarcarComoLida(id int) error {
	resultado, err := dao.db.Exec(`UPDATE notificacoes SET lida = 1 WHERE id = ? AND lida = 0`, id)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrNotificacaoNaoEncontrada
	}
	return nil
}
