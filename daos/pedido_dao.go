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
var ErrPagamentoNaoEncontrado = errors.New("pagamento não encontrado")
var ErrPagamentoJaRecebido = errors.New("pagamento já recebido")
var ErrPedidoNaoEncontrado = errors.New("pedido não encontrado")
var ErrEtiquetaPedidoNaoEncontrada = errors.New("etiqueta do pedido não encontrada")
var ErrPedidoPrincipalNaoEncontrado = errors.New("pedido principal não encontrado")
var ErrPedidoPrincipalInvalido = errors.New("o pedido principal não pode ser um brinde")
var ErrCondicoesPedidoBloqueadas = errors.New("as condições financeiras do pedido não podem ser alteradas após gerar pagamentos")
var ErrFormaPagamentoInvalida = errors.New("forma de pagamento inválida")
var ErrPixParceladoNaoOferecido = errors.New("Pix parcelado não foi oferecido neste orçamento")
var ErrNumeroParcelasInvalido = errors.New("número de parcelas inválido")
var ErrDataParcelaInvalida = errors.New("data da primeira parcela inválida")
var ErrValorTotalInvalido = errors.New("valor total do pedido inválido")
var ErrPedidoSemItens = errors.New("o pedido precisa conter ao menos um produto")

type PedidoDAO struct {
	db *sql.DB
}

func NovoPedidoDAO(db *sql.DB) *PedidoDAO {
	return &PedidoDAO{db: db}
}

func (dao *PedidoDAO) Inserir(pedido dtos.PedidoInputDTO) (int, error) {
	itens, precoBase, err := prepararItensPedido(pedido.Itens, pedido.ProdutoID, pedido.PrecoBase, pedido.Brinde)
	if err != nil {
		return 0, err
	}
	pedido.Itens = itens
	pedido.ProdutoID = itens[0].ProdutoID
	pedido.PrecoBase = precoBase
	if pedido.PedidoPrincipalID > 0 {
		var principalBrinde bool
		if err := dao.db.QueryRow(`SELECT brinde FROM pedidos WHERE id = ?`, pedido.PedidoPrincipalID).Scan(&principalBrinde); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrPedidoPrincipalNaoEncontrado
			}
			return 0, err
		}
		if principalBrinde {
			return 0, ErrPedidoPrincipalInvalido
		}
	}
	if pedido.Brinde {
		pedido.Status = modelos.StatusEmProducao
		pedido.PrecoBase = 0
		pedido.DescontoCartao = 0
		pedido.DescontoPix = 0
		pedido.DescontoPixAvista = false
		pedido.PixParcelado = false
		pedido.IncluirBrinde = false
	} else {
		pedido.PedidoPrincipalID = 0
	}
	tx, err := dao.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if pedido.PedidoPrincipalID > 0 {
		var principalBrinde bool
		if err := tx.QueryRow(`SELECT brinde FROM pedidos WHERE id = ?`, pedido.PedidoPrincipalID).Scan(&principalBrinde); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrPedidoPrincipalNaoEncontrado
			}
			return 0, err
		}
		if principalBrinde {
			return 0, ErrPedidoPrincipalInvalido
		}
	}
	query := `INSERT INTO pedidos (cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, observacoes, mensagem_orcamento, linha_cores_nome, linha_cores_link, preco_base, desconto_cartao, desconto_pix, desconto_pix_avista, pix_parcelado, brinde, incluir_brinde, pedido_principal_id)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, 0))`

	resultado, err := tx.Exec(query, pedido.ClienteID, pedido.ProdutoID, pedido.Status,
		pedido.DataPedido, pedido.DiasPrazo, pedido.DataEntrega, pedido.Observacoes, pedido.MensagemOrcamento, pedido.LinhaCoresNome, pedido.LinhaCoresLink,
		pedido.PrecoBase, pedido.DescontoCartao, pedido.DescontoPix, pedido.DescontoPixAvista, pedido.PixParcelado,
		pedido.Brinde, pedido.IncluirBrinde, pedido.PedidoPrincipalID)
	if err != nil {
		log.Printf("Erro ao inserir pedido: %v", err)
		return 0, err
	}

	id, err := resultado.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := inserirItensPedido(tx, int(id), pedido.Itens); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(id), nil
}

func (dao *PedidoDAO) AtualizarDetalhes(id int, pedido dtos.PedidoEdicaoDTO) error {
	itens, precoBase, err := prepararItensPedido(pedido.Itens, pedido.ProdutoID, pedido.PrecoBase, pedido.Brinde)
	if err != nil {
		return err
	}
	pedido.Itens = itens
	pedido.ProdutoID = itens[0].ProdutoID
	pedido.PrecoBase = precoBase
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var brindeAtual, descontoPixAvistaAtual, pixParceladoAtual bool
	var precoBaseAtual, descontoCartaoAtual, descontoPixAtual float64
	if err := tx.QueryRow(`SELECT status, brinde, preco_base, desconto_cartao, desconto_pix, desconto_pix_avista, pix_parcelado
		FROM pedidos WHERE id = ?`, id).Scan(&status, &brindeAtual, &precoBaseAtual, &descontoCartaoAtual,
		&descontoPixAtual, &descontoPixAvistaAtual, &pixParceladoAtual); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPedidoNaoEncontrado
		}
		return err
	}
	var quantidadePagamentos int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM pagamentos WHERE pedido_id = ?`, id).Scan(&quantidadePagamentos); err != nil {
		return err
	}
	itensAtuais, err := buscarItensPedido(tx, id)
	if err != nil {
		return err
	}
	itensFinanceirosAlterados := !itensPedidoIguais(pedido.Itens, itensAtuais)
	if quantidadePagamentos > 0 && (itensFinanceirosAlterados || pedido.Brinde != brindeAtual || pedido.DescontoCartao != descontoCartaoAtual ||
		pedido.DescontoPix != descontoPixAtual || pedido.DescontoPixAvista != descontoPixAvistaAtual ||
		pedido.PixParcelado != pixParceladoAtual || pedido.PrecoBase != precoBaseAtual) {
		return ErrCondicoesPedidoBloqueadas
	}
	if pedido.Brinde {
		pedido.PrecoBase = 0
		pedido.DescontoCartao = 0
		pedido.DescontoPix = 0
		pedido.DescontoPixAvista = false
		pedido.PixParcelado = false
	} else {
		pedido.PedidoPrincipalID = 0
	}
	if pedido.Brinde && pedido.PedidoPrincipalID > 0 {
		if pedido.PedidoPrincipalID == id {
			return ErrPedidoPrincipalInvalido
		}
		var principalBrinde bool
		if err := tx.QueryRow(`SELECT brinde FROM pedidos WHERE id = ?`, pedido.PedidoPrincipalID).Scan(&principalBrinde); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrPedidoPrincipalNaoEncontrado
			}
			return err
		}
		if principalBrinde {
			return ErrPedidoPrincipalInvalido
		}
	}
	statusAtualizado := status
	if !brindeAtual && pedido.Brinde && status == modelos.StatusOrcamento {
		statusAtualizado = modelos.StatusEmProducao
	} else if brindeAtual && !pedido.Brinde && status == modelos.StatusEmProducao {
		statusAtualizado = modelos.StatusOrcamento
	}
	resultado, err := tx.Exec(`UPDATE pedidos SET cliente_id = ?, produto_id = ?, status = ?, data_pedido = ?, dias_prazo = ?, data_entrega = ?, observacoes = ?, mensagem_orcamento = ?, linha_cores_nome = ?, linha_cores_link = ?, preco_base = ?, desconto_cartao = ?, desconto_pix = ?, desconto_pix_avista = ?, pix_parcelado = ?, brinde = ?, incluir_brinde = ?, pedido_principal_id = NULLIF(?, 0) WHERE id = ?`,
		pedido.ClienteID, pedido.ProdutoID, statusAtualizado, pedido.DataPedido, pedido.DiasPrazo, pedido.DataEntrega,
		pedido.Observacoes, pedido.MensagemOrcamento, pedido.LinhaCoresNome, pedido.LinhaCoresLink, pedido.PrecoBase,
		pedido.DescontoCartao, pedido.DescontoPix, pedido.DescontoPixAvista, pedido.PixParcelado, pedido.Brinde,
		pedido.IncluirBrinde, pedido.PedidoPrincipalID, id)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrPedidoNaoEncontrado
	}
	if quantidadePagamentos == 0 {
		if _, err := tx.Exec(`DELETE FROM itens_pedido WHERE pedido_id = ?`, id); err != nil {
			return err
		}
		if err := inserirItensPedido(tx, id, pedido.Itens); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func prepararItensPedido(itens []dtos.ItemPedidoInputDTO, produtoID int, precoBase float64, brinde bool) ([]dtos.ItemPedidoInputDTO, float64, error) {
	if len(itens) == 0 && produtoID > 0 {
		itens = []dtos.ItemPedidoInputDTO{{ProdutoID: produtoID, NomeProduto: "Produto", Quantidade: 1, PrecoUnitario: precoBase}}
	}
	if len(itens) == 0 {
		return nil, 0, ErrPedidoSemItens
	}
	total := 0.0
	for indice := range itens {
		item := &itens[indice]
		if item.ProdutoID < 1 || item.Quantidade < 1 || item.PrecoUnitario < 0 || math.IsNaN(item.PrecoUnitario) || math.IsInf(item.PrecoUnitario, 0) {
			return nil, 0, ErrValorTotalInvalido
		}
		if item.NomeProduto == "" {
			item.NomeProduto = "Produto"
		}
		if brinde {
			item.PrecoUnitario = 0
		}
		total += float64(item.Quantidade) * item.PrecoUnitario
	}
	if brinde {
		total = 0
	}
	return itens, math.Round(total*100) / 100, nil
}

func inserirItensPedido(tx *sql.Tx, pedidoID int, itens []dtos.ItemPedidoInputDTO) error {
	for _, item := range itens {
		if _, err := tx.Exec(`INSERT INTO itens_pedido (pedido_id, produto_id, nome_produto, categoria, quantidade, preco_unitario) VALUES (?, ?, ?, ?, ?, ?)`,
			pedidoID, item.ProdutoID, item.NomeProduto, item.Categoria, item.Quantidade, item.PrecoUnitario); err != nil {
			return err
		}
	}
	return nil
}

func buscarItensPedido(tx *sql.Tx, pedidoID int) ([]modelos.ItemPedido, error) {
	linhas, err := tx.Query(`SELECT produto_id, nome_produto, categoria, quantidade, preco_unitario FROM itens_pedido WHERE pedido_id = ? ORDER BY id`, pedidoID)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()
	var itens []modelos.ItemPedido
	for linhas.Next() {
		var item modelos.ItemPedido
		item.PedidoID = pedidoID
		if err := linhas.Scan(&item.ProdutoID, &item.NomeProduto, &item.Categoria, &item.Quantidade, &item.PrecoUnitario); err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, linhas.Err()
}

func itensPedidoIguais(novos []dtos.ItemPedidoInputDTO, atuais []modelos.ItemPedido) bool {
	if len(novos) != len(atuais) {
		return false
	}
	for indice, novo := range novos {
		atual := atuais[indice]
		if novo.ProdutoID != atual.ProdutoID || novo.Quantidade != atual.Quantidade || novo.PrecoUnitario != atual.PrecoUnitario {
			return false
		}
	}
	return true
}

func (dao *PedidoDAO) BuscarTodos() ([]modelos.Pedido, error) {
	query := `SELECT id, cliente_id, produto_id, status, data_pedido, dias_prazo, data_entrega, COALESCE(observacoes, ''), COALESCE(mensagem_orcamento, ''), COALESCE(linha_cores_nome, ''), COALESCE(linha_cores_link, ''), preco_base, desconto_cartao, desconto_pix, desconto_pix_avista, pix_parcelado, forma_pagamento, total_acordado, brinde, incluir_brinde, COALESCE(pedido_principal_id, 0)
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
			&p.DataPedido, &p.DiasPrazo, &p.DataEntrega, &p.Observacoes, &p.MensagemOrcamento, &p.LinhaCoresNome, &p.LinhaCoresLink,
			&p.PrecoBase, &p.DescontoCartao, &p.DescontoPix, &p.DescontoPixAvista, &p.PixParcelado, &p.FormaPagamento, &p.TotalAcordado,
			&p.Brinde, &p.IncluirBrinde, &p.PedidoPrincipalID)
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

	etiquetas, err := dao.buscarEtiquetasPedidos()
	if err != nil {
		return nil, err
	}
	for _, etiqueta := range etiquetas {
		if i, ok := pedidosPorID[etiqueta.PedidoID]; ok {
			pedidos[i].Etiquetas = append(pedidos[i].Etiquetas, etiqueta)
		}
	}

	itens, err := dao.buscarItensPedidos()
	if err != nil {
		return nil, err
	}
	for _, item := range itens {
		if i, ok := pedidosPorID[item.PedidoID]; ok {
			pedidos[i].Itens = append(pedidos[i].Itens, item)
		}
	}

	return pedidos, nil
}

func (dao *PedidoDAO) buscarItensPedidos() ([]modelos.ItemPedido, error) {
	linhas, err := dao.db.Query(`SELECT pedido_id, produto_id, nome_produto, categoria, quantidade, preco_unitario FROM itens_pedido ORDER BY pedido_id, id`)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()
	var itens []modelos.ItemPedido
	for linhas.Next() {
		var item modelos.ItemPedido
		if err := linhas.Scan(&item.PedidoID, &item.ProdutoID, &item.NomeProduto, &item.Categoria, &item.Quantidade, &item.PrecoUnitario); err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, linhas.Err()
}

func (dao *PedidoDAO) AdicionarEtiquetaPedido(pedidoID int, texto string, cor string) error {
	var existe bool
	if err := dao.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pedidos WHERE id = ?)`, pedidoID).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return ErrPedidoNaoEncontrado
	}
	_, err := dao.db.Exec(`INSERT INTO etiquetas_pedido (pedido_id, texto, cor) VALUES (?, ?, ?)`, pedidoID, texto, cor)
	return err
}

func (dao *PedidoDAO) ExcluirEtiquetaPedido(pedidoID int, etiquetaID int) error {
	resultado, err := dao.db.Exec(`DELETE FROM etiquetas_pedido WHERE id = ? AND pedido_id = ?`, etiquetaID, pedidoID)
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrEtiquetaPedidoNaoEncontrada
	}
	return nil
}

func (dao *PedidoDAO) buscarEtiquetasPedidos() ([]modelos.EtiquetaPedido, error) {
	linhas, err := dao.db.Query(`SELECT id, pedido_id, texto, cor FROM etiquetas_pedido ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var etiquetas []modelos.EtiquetaPedido
	for linhas.Next() {
		var etiqueta modelos.EtiquetaPedido
		if err := linhas.Scan(&etiqueta.ID, &etiqueta.PedidoID, &etiqueta.Texto, &etiqueta.Cor); err != nil {
			return nil, err
		}
		etiquetas = append(etiquetas, etiqueta)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}
	return etiquetas, nil
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
	var pixParcelado, descontoPixAvista, brinde bool
	if err := tx.QueryRow(`SELECT status, preco_base, desconto_cartao, desconto_pix, desconto_pix_avista, pix_parcelado, brinde FROM pedidos WHERE id = ?`, id).
		Scan(&statusAtual, &precoBase, &descontoCartao, &descontoPix, &descontoPixAvista, &pixParcelado, &brinde); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPedidoNaoEncontrado
		}
		return err
	}
	if precoBase == 0 && !brinde {
		if err := tx.QueryRow(`SELECT preco FROM produtos WHERE id = (SELECT produto_id FROM pedidos WHERE id = ?)`, id).Scan(&precoBase); err != nil {
			return err
		}
	}

	if !statusPedidoPermitido(statusAtual, novoStatus, brinde) {
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
			if descontoPixAvista {
				desconto = 10
			}
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
	if _, err := tx.Exec(`UPDATE pedidos SET status = ? WHERE id = ? AND status = ?
		AND NOT EXISTS (SELECT 1 FROM pagamentos WHERE pedido_id = ? AND status != 'Recebido')`,
		modelos.StatusPago, pedidoID, modelos.StatusAguardandoPagamento, pedidoID); err != nil {
		return err
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

func statusPedidoPermitido(atual string, proximo string, brinde bool) bool {
	for _, permitido := range modelos.ProximosStatusPedidoComBrinde(atual, brinde) {
		if proximo == permitido {
			return true
		}
	}
	return false
}
