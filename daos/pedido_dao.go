package daos

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"strconv"
	"time"

	"falkcroche/dtos"
	"falkcroche/modelos"
)

var ErrTransicaoPedidoInvalida = errors.New("transição de status inválida")
var ErrPedidoNaoEncontrado = errors.New("pedido não encontrado")

type PedidoDAO struct {
	db *sql.DB
}

func NovoPedidoDAO(db *sql.DB) *PedidoDAO {
	return &PedidoDAO{db: db}
}

func (dao *PedidoDAO) Inserir(pedido dtos.PedidoInputDTO) (int, error) {
	query := `INSERT INTO pedidos (cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, observacoes, mensagem_orcamento) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	resultado, err := dao.db.Exec(query, pedido.ClienteID, pedido.ProdutoID, pedido.Status,
		pedido.DataPedido, pedido.DiasPrazo, pedido.DataEntrega, pedido.Observacoes, pedido.MensagemOrcamento)
	if err != nil {
		log.Printf("Erro ao inserir pedido: %v", err)
		return 0, err
	}

	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (dao *PedidoDAO) BuscarTodos() ([]modelos.Pedido, error) {
	query := `SELECT id, cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, COALESCE(observacoes, ''), COALESCE(mensagem_orcamento, '') FROM pedidos ORDER BY id DESC`

	linhas, err := dao.db.Query(query)
	if err != nil {
		log.Printf("Erro ao buscar pedidos: %v", err)
		return nil, err
	}
	defer linhas.Close()

	var pedidos []modelos.Pedido

	for linhas.Next() {
		var p modelos.Pedido
		err := linhas.Scan(&p.ID, &p.ClienteID, &p.ProdutoID, &p.Status,
			&p.DataPedido, &p.DiasPrazo, &p.DataEntrega, &p.Observacoes, &p.MensagemOrcamento)
		if err != nil {
			log.Printf("Erro ao ler pedido: %v", err)
			continue
		}
		pedidos = append(pedidos, p)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}

	return pedidos, nil
}

func (dao *PedidoDAO) AtualizarStatus(id int, novoStatus string) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var statusAtual string
	var preco float64
	if err := tx.QueryRow(`SELECT pedidos.status, produtos.preco FROM pedidos JOIN produtos ON produtos.id = pedidos.produto_id WHERE pedidos.id = ?`, id).Scan(&statusAtual, &preco); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPedidoNaoEncontrado
		}
		return err
	}

	if !statusPedidoPermitido(statusAtual, novoStatus) {
		return ErrTransicaoPedidoInvalida
	}

	if _, err := tx.Exec(`UPDATE pedidos SET status = ? WHERE id = ?`, novoStatus, id); err != nil {
		return err
	}

	var etapa string
	var percentual float64
	if novoStatus == modelos.StatusSinalRecebido {
		etapa = "sinal"
		percentual = 0.30
	} else if novoStatus == modelos.StatusPago {
		etapa = "pagamento final"
		percentual = 0.70
	}
	if etapa != "" {
		valor := math.Round(preco*percentual*100) / 100
		descricao := "Pedido #" + strconv.Itoa(id) + " - " + etapa
		if _, err := tx.Exec(`INSERT INTO transacoes (tipo, descricao, valor, data, pedido_id, etapa) VALUES ('Entrada', ?, ?, ?, ?, ?) ON CONFLICT(pedido_id, etapa) WHERE pedido_id IS NOT NULL DO NOTHING`, descricao, valor, time.Now().Format("2006-01-02"), id, etapa); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func statusPedidoPermitido(atual string, proximo string) bool {
	for _, permitido := range modelos.ProximosStatusPedido(atual) {
		if proximo == permitido {
			return true
		}
	}
	return false
}
