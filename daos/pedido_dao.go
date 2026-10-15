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
var ErrPagamentoNaoEncontrado = errors.New("pagamento não encontrado")
var ErrPagamentoJaRecebido = errors.New("pagamento já recebido")
var ErrFormaPagamentoInvalida = errors.New("forma de pagamento inválida")
var ErrPixParceladoNaoOferecido = errors.New("Pix parcelado não foi oferecido neste orçamento")
var ErrNumeroParcelasInvalido = errors.New("número de parcelas inválido")
var ErrDataParcelaInvalida = errors.New("data da primeira parcela inválida")
var ErrValorTotalInvalido = errors.New("valor total do pedido inválido")

type PedidoDAO struct {
	db *sql.DB
}

func NovoPedidoDAO(db *sql.DB) *PedidoDAO {
	return &PedidoDAO{db: db}
}

func (dao *PedidoDAO) Inserir(pedido dtos.PedidoInputDTO) (int, error) {
	query := `INSERT INTO pedidos (cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, observacoes, mensagem_orcamento, preco_base, desconto_cartao, desconto_pix, pix_parcelado)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	resultado, err := dao.db.Exec(query, pedido.ClienteID, pedido.ProdutoID, pedido.Status,
		pedido.DataPedido, pedido.DiasPrazo, pedido.DataEntrega, pedido.Observacoes, pedido.MensagemOrcamento,
		pedido.PrecoBase, pedido.DescontoCartao, pedido.DescontoPix, pedido.PixParcelado)
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
	query := `SELECT id, cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, COALESCE(observacoes, ''), COALESCE(mensagem_orcamento, ''), preco_base, desconto_cartao, desconto_pix, pix_parcelado, forma_pagamento, total_acordado
		FROM pedidos ORDER BY id DESC`

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
			&p.DataPedido, &p.DiasPrazo, &p.DataEntrega, &p.Observacoes, &p.MensagemOrcamento,
			&p.PrecoBase, &p.DescontoCartao, &p.DescontoPix, &p.PixParcelado, &p.FormaPagamento, &p.TotalAcordado)
		if err != nil {
			log.Printf("Erro ao ler pedido: %v", err)
			return nil, err
		}
		pedidos = append(pedidos, p)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	linhas.Close()

	pagamentos, err := dao.buscarPagamentos()
	if err != nil {
		return nil, err
	}
	pedidosPorID := make(map[int]int, len(pedidos))
	for i := range pedidos {
		pedidosPorID[pedidos[i].ID] = i
	}
	for _, pagamento := range pagamentos {
		if i, ok := pedidosPorID[pagamento.PedidoID]; ok {
			pedidos[i].Pagamentos = append(pedidos[i].Pagamentos, pagamento)
		}
	}

	return pedidos, nil
}

func (dao *PedidoDAO) AtualizarStatus(id int, novoStatus string) error {
	return dao.AtualizarStatusComPagamento(id, novoStatus, "", 1, time.Now().Format("2006-01-02"))
}

func (dao *PedidoDAO) AtualizarStatusComPagamento(id int, novoStatus string, formaPagamento string, numeroParcelas int, primeiraData string) error {
	return dao.avancarStatus(id, novoStatus, formaPagamento, numeroParcelas, primeiraData, primeiraData)
}

func (dao *PedidoDAO) AvancarStatusComRecebimentoSinal(id int, formaPagamento string, numeroParcelas int, primeiraData string, dataRecebimento string) error {
	return dao.avancarStatus(id, modelos.StatusSinalRecebido, formaPagamento, numeroParcelas, primeiraData, dataRecebimento)
}

func (dao *PedidoDAO) avancarStatus(id int, novoStatus string, formaPagamento string, numeroParcelas int, primeiraData string, dataRecebimento string) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var statusAtual string
	var precoBase, descontoCartao, descontoPix float64
	var pixParcelado bool
	if err := tx.QueryRow(`SELECT status, preco_base, desconto_cartao, desconto_pix, pix_parcelado FROM pedidos WHERE id = ?`, id).
		Scan(&statusAtual, &precoBase, &descontoCartao, &descontoPix, &pixParcelado); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPedidoNaoEncontrado
		}
		return err
	}
	if precoBase == 0 {
		if err := tx.QueryRow(`SELECT preco FROM produtos WHERE id = (SELECT produto_id FROM pedidos WHERE id = ?)`, id).Scan(&precoBase); err != nil {
			return err
		}
	}

	if !statusPedidoPermitido(statusAtual, novoStatus) {
		return ErrTransicaoPedidoInvalida
	}

	if novoStatus == modelos.StatusSinalRecebido && statusAtual == modelos.StatusAguardandoSinal {
		if formaPagamento != "Pix à vista" && formaPagamento != "Pix parcelado" && formaPagamento != "Cartão de crédito" {
			return ErrFormaPagamentoInvalida
		}
		if formaPagamento == "Pix parcelado" && !pixParcelado {
			return ErrPixParceladoNaoOferecido
		}
		if numeroParcelas < 1 || numeroParcelas > 3 {
			return ErrNumeroParcelasInvalido
		}
		if formaPagamento == "Pix à vista" {
			numeroParcelas = 1
		}
		if _, err := time.Parse("2006-01-02", primeiraData); err != nil {
			return ErrDataParcelaInvalida
		}
		desconto := 0.0
		switch formaPagamento {
		case "Pix à vista":
			desconto = 10
		case "Pix parcelado":
			desconto = descontoPix
		case "Cartão de crédito":
			desconto = descontoCartao
		}
		total := math.Round(precoBase*(1-desconto/100)*100) / 100
		if total <= 0 {
			return ErrValorTotalInvalido
		}
		if _, err := tx.Exec(`UPDATE pedidos SET forma_pagamento = ?, total_acordado = ? WHERE id = ?`, formaPagamento, total, id); err != nil {
			return err
		}
		if err := criarPagamentos(tx, id, formaPagamento, total, numeroParcelas, primeiraData); err != nil {
			return err
		}
		etapaRecebida := "sinal"
		if formaPagamento == "Pix à vista" {
			etapaRecebida = "avista"
		}
		resultado, err := tx.Exec(`UPDATE pagamentos SET status = 'Recebido', data_recebimento = ?
			WHERE pedido_id = ? AND etapa = ? AND status = 'Pendente'`, dataRecebimento, id, etapaRecebida)
		if err != nil {
			return err
		}
		afetadas, err := resultado.RowsAffected()
		if err != nil {
			return err
		}
		if afetadas == 0 {
			return ErrPagamentoNaoEncontrado
		}
		var descricao string
		var valor float64
		if err := tx.QueryRow(`SELECT descricao, valor FROM pagamentos WHERE pedido_id = ? AND etapa = ?`, id, etapaRecebida).
			Scan(&descricao, &valor); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO transacoes (tipo, descricao, valor, data, pedido_id, etapa) VALUES ('Entrada', ?, ?, ?, ?, ?)`,
			"Pedido #"+strconv.Itoa(id)+" - "+descricao+" ("+formaPagamento+")", valor, dataRecebimento, id, etapaRecebida); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE notificacoes SET lida = 1
			WHERE pagamento_id = (SELECT id FROM pagamentos WHERE pedido_id = ? AND etapa = ?)`, id, etapaRecebida); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE pedidos SET status = ? WHERE id = ?`, novoStatus, id); err != nil {
		return err
	}

	return tx.Commit()
}

func criarPagamentos(tx *sql.Tx, pedidoID int, formaPagamento string, total float64, numeroParcelas int, primeiraData string) error {
	if formaPagamento == "Pix à vista" {
		_, err := tx.Exec(`INSERT INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento) VALUES (?, 'avista', 'Pagamento à vista', ?, ?)`,
			pedidoID, total, time.Now().Format("2006-01-02"))
		return err
	}

	sinal := math.Round(total*0.30*100) / 100
	descricaoSinal := "Sinal (30%)"
	if _, err := tx.Exec(`INSERT INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento) VALUES (?, 'sinal', ?, ?, ?)`,
		pedidoID, descricaoSinal, sinal, time.Now().Format("2006-01-02")); err != nil {
		return err
	}
	restante := total - sinal
	parcelaBase := math.Round(restante/float64(numeroParcelas)*100) / 100
	for i := 1; i <= numeroParcelas; i++ {
		valor := parcelaBase
		if i == numeroParcelas {
			valor = restante - parcelaBase*float64(numeroParcelas-1)
		}
		data, err := time.Parse("2006-01-02", primeiraData)
		if err != nil {
			return err
		}
		data = adicionarMeses(data, i-1)
		etapa := "parcela " + strconv.Itoa(i)
		descricao := "Parcela " + strconv.Itoa(i) + " de " + strconv.Itoa(numeroParcelas)
		if _, err := tx.Exec(`INSERT INTO pagamentos (pedido_id, etapa, descricao, valor, data_vencimento) VALUES (?, ?, ?, ?, ?)`,
			pedidoID, etapa, descricao, valor, data.Format("2006-01-02")); err != nil {
			return err
		}
	}
	return nil
}

func adicionarMeses(data time.Time, quantidade int) time.Time {
	ano, mes, dia := data.Date()
	primeiroDiaMes := time.Date(ano, mes+time.Month(quantidade), 1, 0, 0, 0, 0, data.Location())
	ultimoDia := time.Date(primeiroDiaMes.Year(), primeiroDiaMes.Month()+1, 0, 0, 0, 0, 0, data.Location()).Day()
	if dia > ultimoDia {
		dia = ultimoDia
	}
	return time.Date(primeiroDiaMes.Year(), primeiroDiaMes.Month(), dia, 0, 0, 0, 0, data.Location())
}

func (dao *PedidoDAO) RegistrarRecebimento(pagamentoID int, dataRecebimento string) error {
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var pedidoID int
	var etapa, descricao, status, formaPagamento string
	var valor float64
	if err := tx.QueryRow(`SELECT pagamentos.pedido_id, pagamentos.etapa, pagamentos.descricao, pagamentos.valor, pagamentos.status, pedidos.forma_pagamento
		FROM pagamentos JOIN pedidos ON pedidos.id = pagamentos.pedido_id WHERE pagamentos.id = ?`, pagamentoID).
		Scan(&pedidoID, &etapa, &descricao, &valor, &status, &formaPagamento); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPagamentoNaoEncontrado
		}
		return err
	}
	if status != "Pendente" {
		return ErrPagamentoJaRecebido
	}
	resultado, err := tx.Exec(`UPDATE pagamentos SET status = 'Recebido', data_recebimento = ? WHERE id = ? AND status = 'Pendente'`,
		dataRecebimento, pagamentoID)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrPagamentoJaRecebido
	}
	if _, err := tx.Exec(`INSERT INTO transacoes (tipo, descricao, valor, data, pedido_id, etapa) VALUES ('Entrada', ?, ?, ?, ?, ?)`,
		"Pedido #"+strconv.Itoa(pedidoID)+" - "+descricao+" ("+formaPagamento+")", valor, dataRecebimento, pedidoID, etapa); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE notificacoes SET lida = 1 WHERE pagamento_id = ?`, pagamentoID); err != nil {
		return err
	}
	if etapa == "sinal" || etapa == "avista" {
		if _, err := tx.Exec(`UPDATE pedidos SET status = ? WHERE id = ? AND status = ?`,
			modelos.StatusSinalRecebido, pedidoID, modelos.StatusAguardandoSinal); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (dao *PedidoDAO) buscarPagamentos() ([]modelos.Pagamento, error) {
	linhas, err := dao.db.Query(`SELECT id, pedido_id, etapa, descricao, valor, data_vencimento, data_recebimento, status FROM pagamentos ORDER BY pedido_id, id`)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var pagamentos []modelos.Pagamento
	for linhas.Next() {
		var pagamento modelos.Pagamento
		if err := linhas.Scan(&pagamento.ID, &pagamento.PedidoID, &pagamento.Etapa, &pagamento.Descricao,
			&pagamento.Valor, &pagamento.DataVencimento, &pagamento.DataRecebimento, &pagamento.Status); err != nil {
			return nil, err
		}
		pagamentos = append(pagamentos, pagamento)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return pagamentos, nil
}

func statusPedidoPermitido(atual string, proximo string) bool {
	for _, permitido := range modelos.ProximosStatusPedido(atual) {
		if proximo == permitido {
			return true
		}
	}
	return false
}
