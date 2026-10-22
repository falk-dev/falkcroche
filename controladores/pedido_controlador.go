package controladores

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"falkcroche/daos"
	"falkcroche/dtos"
	"falkcroche/modelos"
	"falkcroche/views"
)

const prazoPadraoDiasUteis = 20

type PedidoControlador struct {
	dao        *daos.PedidoDAO
	clienteDAO *daos.ClienteDAO
	produtoDAO *daos.ProdutoDAO
}

func NovoPedidoControlador(dao *daos.PedidoDAO, cDAO *daos.ClienteDAO, pDAO *daos.ProdutoDAO) *PedidoControlador {
	return &PedidoControlador{
		dao:        dao,
		clienteDAO: cDAO,
		produtoDAO: pDAO,
	}
}

func (c *PedidoControlador) Inserir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}

	brinde := r.FormValue("pedido_brinde") == "on"
	temDescontoPixAvista := !brinde && r.FormValue("desconto_pix_avista") == "on"
	clienteID, erroCliente := strconv.Atoi(r.FormValue("cliente_id"))
	if erroCliente != nil || clienteID < 1 {
		http.Error(w, "Selecione um cliente.", http.StatusBadRequest)
		return
	}

	dataPedido := time.Now()
	if valorDataPedido := r.FormValue("data_pedido"); valorDataPedido != "" {
		var err error
		dataPedido, err = time.Parse("2006-01-02", valorDataPedido)
		if err != nil {
			http.Error(w, "Informe uma data de pedido válida.", http.StatusBadRequest)
			return
		}
	}

	temDesconto := !brinde && r.FormValue("tem_desconto") == "on"
	percentualDesconto := 0.0
	if temDesconto {
		var erroPercentual error
		percentualDesconto, erroPercentual = strconv.ParseFloat(r.FormValue("percentual_desconto"), 64)
		if erroPercentual != nil || percentualDesconto <= 0 || percentualDesconto > 100 {
			http.Error(w, "Informe um percentual de desconto entre 0 e 100.", http.StatusBadRequest)
			return
		}
	}
	temPixParcelado := !brinde && r.FormValue("pix_parcelado") == "on"
	temDescontoPixParcelado := r.FormValue("tem_desconto_pix_parcelado") == "on" && temPixParcelado
	percentualDescontoPixParcelado := 0.0
	if temDescontoPixParcelado {
		var erroPercentualPix error
		percentualDescontoPixParcelado, erroPercentualPix = strconv.ParseFloat(r.FormValue("percentual_desconto_pix_parcelado"), 64)
		if erroPercentualPix != nil || percentualDescontoPixParcelado <= 0 || percentualDescontoPixParcelado > 100 {
			http.Error(w, "Informe um percentual de desconto do Pix parcelado entre 0 e 100.", http.StatusBadRequest)
			return
		}
	}
	incluirBrinde := r.FormValue("incluir_brinde") == "on"
	pedidoPrincipalID := 0
	if brinde && r.FormValue("pedido_principal_id") != "" {
		idPrincipal, err := strconv.Atoi(r.FormValue("pedido_principal_id"))
		if err != nil || idPrincipal < 1 {
			http.Error(w, "Selecione um pedido principal válido.", http.StatusBadRequest)
			return
		}
		pedidoPrincipalID = idPrincipal
	}

	dataEntregaCalculada := calcularDataEntrega(dataPedido, prazoPadraoDiasUteis)
	clientes, err := c.clienteDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar clientes.", http.StatusInternalServerError)
		return
	}
	produtos, err := c.produtoDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar produtos.", http.StatusInternalServerError)
		return
	}
	var clienteNome string
	for _, cliente := range clientes {
		if cliente.ID == clienteID {
			clienteNome = cliente.Nome
			break
		}
	}
	itens, produtoPreco, err := construirItensPedido(r.Form, produtos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if clienteNome == "" {
		http.Error(w, "Cliente não encontrado.", http.StatusBadRequest)
		return
	}
	produtoNome := nomesItensPedidoDTO(itens)
	produtoID := itens[0].ProdutoID
	status := modelos.StatusOrcamento
	mensagem := ""
	if brinde {
		produtoPreco = 0
		status = modelos.StatusEmProducao
		mensagem = montarMensagemBrinde(clienteNome, produtoNome, pedidoPrincipalID, dataEntregaCalculada)
	} else {
		mensagem = montarMensagemOrcamento(clienteNome, produtoNome, produtoPreco, prazoPadraoDiasUteis, dataEntregaCalculada, r.FormValue("linha_cores_nome"), r.FormValue("linha_cores_link"), temDesconto, percentualDesconto, temPixParcelado, incluirBrinde, temDescontoPixParcelado, percentualDescontoPixParcelado, temDescontoPixAvista)
	}

	pedidoDTO := dtos.PedidoInputDTO{
		ClienteID:         clienteID,
		ProdutoID:         produtoID,
		Itens:             itens,
		Status:            status,
		DataPedido:        dataPedido.Format("2006-01-02"),
		DiasPrazo:         prazoPadraoDiasUteis,
		DataEntrega:       dataEntregaCalculada.Format("2006-01-02"),
		Observacoes:       r.FormValue("observacoes"),
		MensagemOrcamento: mensagem,
		LinhaCoresNome:    r.FormValue("linha_cores_nome"),
		LinhaCoresLink:    r.FormValue("linha_cores_link"),
		PrecoBase:         produtoPreco,
		DescontoCartao:    percentualDesconto,
		DescontoPix:       percentualDescontoPixParcelado,
		DescontoPixAvista: temDescontoPixAvista,
		PixParcelado:      temPixParcelado,
		Brinde:            brinde,
		IncluirBrinde:     incluirBrinde,
		PedidoPrincipalID: pedidoPrincipalID,
	}

	if _, err := c.dao.Inserir(pedidoDTO); err != nil {
		if errors.Is(err, daos.ErrPedidoPrincipalNaoEncontrado) {
			http.Error(w, "O pedido principal selecionado não existe.", http.StatusBadRequest)
			return
		}
		if errors.Is(err, daos.ErrPedidoPrincipalInvalido) {
			http.Error(w, "Um brinde não pode ser usado como pedido principal.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Erro ao registrar pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
}

func montarMensagemBrinde(clienteNome string, produtoNome string, pedidoPrincipalID int, entrega time.Time) string {
	relacao := ""
	if pedidoPrincipalID > 0 {
		relacao = fmt.Sprintf(" relacionado ao pedido #%d", pedidoPrincipalID)
	}
	return fmt.Sprintf(
		"oi, %s! este pedido contém o brinde *%s*%s. Valor: %s. Previsão de entrega: %s.",
		clienteNome, produtoNome, relacao, views.FormatarReais(0), entrega.Format("02/01/2006"))
}

func construirItensPedido(form url.Values, produtos []modelos.Produto) ([]dtos.ItemPedidoInputDTO, float64, error) {
	ids := form["produto_id"]
	if len(ids) == 0 {
		return nil, 0, daos.ErrPedidoSemItens
	}
	produtosPorID := make(map[int]modelos.Produto, len(produtos))
	for _, produto := range produtos {
		produtosPorID[produto.ID] = produto
	}
	selecionados := make(map[int]bool, len(ids))
	itens := make([]dtos.ItemPedidoInputDTO, 0, len(ids))
	total := 0.0
	for _, valorID := range ids {
		produtoID, err := strconv.Atoi(valorID)
		if err != nil || produtoID < 1 || selecionados[produtoID] {
			return nil, 0, daos.ErrPedidoSemItens
		}
		produto, existe := produtosPorID[produtoID]
		if !existe {
			return nil, 0, fmt.Errorf("Um dos produtos selecionados não existe.")
		}
		quantidade := 1
		if valorQuantidade := form.Get("quantidade_" + strconv.Itoa(produtoID)); valorQuantidade != "" {
			quantidade, err = strconv.Atoi(valorQuantidade)
			if err != nil || quantidade < 1 || quantidade > 999 {
				return nil, 0, fmt.Errorf("Informe uma quantidade válida para %s.", produto.NomePeca)
			}
		}
		selecionados[produtoID] = true
		itens = append(itens, dtos.ItemPedidoInputDTO{
			ProdutoID: produto.ID, NomeProduto: produto.NomePeca, Categoria: produto.Categoria,
			Quantidade: quantidade, PrecoUnitario: produto.Preco,
		})
		total += float64(quantidade) * produto.Preco
	}
	return itens, math.Round(total*100) / 100, nil
}

func nomesItensPedidoDTO(itens []dtos.ItemPedidoInputDTO) string {
	nomes := make([]string, 0, len(itens))
	for _, item := range itens {
		if item.Quantidade > 1 {
			nomes = append(nomes, fmt.Sprintf("%dx %s", item.Quantidade, item.NomeProduto))
		} else {
			nomes = append(nomes, item.NomeProduto)
		}
	}
	return strings.Join(nomes, ", ")
}

func itensPedidoDTO(itens []modelos.ItemPedido) []dtos.ItemPedidoInputDTO {
	resultado := make([]dtos.ItemPedidoInputDTO, 0, len(itens))
	for _, item := range itens {
		resultado = append(resultado, dtos.ItemPedidoInputDTO{
			ProdutoID: item.ProdutoID, NomeProduto: item.NomeProduto, Categoria: item.Categoria,
			Quantidade: item.Quantidade, PrecoUnitario: item.PrecoUnitario,
		})
	}
	return resultado
}

func (c *PedidoControlador) Atualizar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.FormValue("pedido_id"))
	if err != nil || id < 1 {
		http.Error(w, "Pedido inválido.", http.StatusBadRequest)
		return
	}
	pedidos, err := c.dao.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar pedido.", http.StatusInternalServerError)
		return
	}
	var pedidoAtual modelos.Pedido
	for _, pedido := range pedidos {
		if pedido.ID == id {
			pedidoAtual = pedido
			break
		}
	}
	if pedidoAtual.ID == 0 {
		http.NotFound(w, r)
		return
	}
	temPagamentos := len(pedidoAtual.Pagamentos) > 0
	itens := itensPedidoDTO(pedidoAtual.Itens)
	produtoPreco := pedidoAtual.PrecoBase
	if !temPagamentos {
		itens = nil
	}
	clienteID, err := strconv.Atoi(r.FormValue("cliente_id"))
	if err != nil || clienteID < 1 {
		http.Error(w, "Selecione um cliente válido.", http.StatusBadRequest)
		return
	}
	diasPrazo, err := strconv.Atoi(r.FormValue("dias_prazo"))
	if err != nil || diasPrazo < 1 {
		http.Error(w, "Informe um prazo válido.", http.StatusBadRequest)
		return
	}
	clientes, err := c.clienteDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar clientes.", http.StatusInternalServerError)
		return
	}
	clienteNome := ""
	for _, cliente := range clientes {
		if cliente.ID == clienteID {
			clienteNome = cliente.Nome
			break
		}
	}
	if clienteNome == "" {
		http.Error(w, "Cliente não encontrado.", http.StatusBadRequest)
		return
	}
	produtos, err := c.produtoDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar produtos.", http.StatusInternalServerError)
		return
	}
	if temPagamentos {
		produtoPreco = 0
		for _, item := range itens {
			produtoPreco += float64(item.Quantidade) * item.PrecoUnitario
		}
	} else {
		itens, produtoPreco, err = construirItensPedido(r.Form, produtos)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if len(itens) == 0 {
		http.Error(w, "O pedido não possui produtos para edição.", http.StatusConflict)
		return
	}
	produtoID := itens[0].ProdutoID
	produtoNome := nomesItensPedidoDTO(itens)
	if produtoNome == "" {
		http.Error(w, "Selecione ao menos um produto.", http.StatusBadRequest)
		return
	}
	dataPedido := r.FormValue("data_pedido")
	if _, err := time.Parse("2006-01-02", dataPedido); err != nil {
		http.Error(w, "Informe uma data de pedido válida.", http.StatusBadRequest)
		return
	}
	dataEntrega := r.FormValue("data_entrega")
	if _, err := time.Parse("2006-01-02", dataEntrega); err != nil {
		http.Error(w, "Informe uma data de entrega válida.", http.StatusBadRequest)
		return
	}
	brinde := pedidoAtual.Brinde
	if !temPagamentos {
		brinde = r.FormValue("pedido_brinde") == "on"
	}
	pedidoPrincipalID := pedidoAtual.PedidoPrincipalID
	if !brinde {
		pedidoPrincipalID = 0
	} else if !temPagamentos {
		pedidoPrincipalID = 0
		if valor := r.FormValue("pedido_principal_id"); valor != "" {
			pedidoPrincipalID, err = strconv.Atoi(valor)
			if err != nil || pedidoPrincipalID < 1 {
				http.Error(w, "Selecione um pedido principal válido.", http.StatusBadRequest)
				return
			}
		}
	}

	descontoCartao := pedidoAtual.DescontoCartao
	descontoPix := pedidoAtual.DescontoPix
	descontoPixAvista := pedidoAtual.DescontoPixAvista
	pixParcelado := pedidoAtual.PixParcelado
	precoBase := pedidoAtual.PrecoBase
	incluirBrinde := pedidoAtual.IncluirBrinde
	if !temPagamentos {
		descontoCartao = 0
		descontoPix = 0
		descontoPixAvista = r.FormValue("desconto_pix_avista") == "on"
		pixParcelado = r.FormValue("pix_parcelado") == "on"
		incluirBrinde = r.FormValue("incluir_brinde") == "on"
		if !brinde {
			precoBase = produtoPreco
			if r.FormValue("tem_desconto") == "on" {
				descontoCartao, err = strconv.ParseFloat(r.FormValue("percentual_desconto"), 64)
				if err != nil || descontoCartao <= 0 || descontoCartao > 100 || math.IsNaN(descontoCartao) || math.IsInf(descontoCartao, 0) {
					http.Error(w, "Informe um percentual de desconto no cartão entre 0 e 100.", http.StatusBadRequest)
					return
				}
			}
			if pixParcelado && r.FormValue("tem_desconto_pix_parcelado") == "on" {
				descontoPix, err = strconv.ParseFloat(r.FormValue("percentual_desconto_pix_parcelado"), 64)
				if err != nil || descontoPix <= 0 || descontoPix > 100 || math.IsNaN(descontoPix) || math.IsInf(descontoPix, 0) {
					http.Error(w, "Informe um percentual de desconto no Pix parcelado entre 0 e 100.", http.StatusBadRequest)
					return
				}
			}
		} else {
			precoBase = 0
			descontoCartao = 0
			descontoPix = 0
			descontoPixAvista = false
			pixParcelado = false
		}
	}

	linhaCoresNome := r.FormValue("linha_cores_nome")
	linhaCoresLink := r.FormValue("linha_cores_link")
	dataPedidoParsed, _ := time.Parse("2006-01-02", dataPedido)
	dataEntregaParsed, _ := time.Parse("2006-01-02", dataEntrega)
	mensagemOrcamento := ""
	if brinde {
		mensagemOrcamento = montarMensagemBrinde(clienteNome, produtoNome, pedidoPrincipalID, dataEntregaParsed)
	} else {
		precoMensagem := precoBase
		if temPagamentos && precoBase == 0 {
			precoMensagem = produtoPreco
		}
		mensagemOrcamento = montarMensagemOrcamento(clienteNome, produtoNome, precoMensagem, diasPrazo, dataEntregaParsed,
			linhaCoresNome, linhaCoresLink, descontoCartao > 0, descontoCartao, pixParcelado, incluirBrinde,
			descontoPix > 0, descontoPix, descontoPixAvista)
	}
	pedidoEdicao := dtos.PedidoEdicaoDTO{
		ClienteID: clienteID, ProdutoID: produtoID, Itens: itens, DataPedido: dataPedidoParsed.Format("2006-01-02"),
		DiasPrazo: diasPrazo, DataEntrega: dataEntregaParsed.Format("2006-01-02"), Observacoes: r.FormValue("observacoes"),
		MensagemOrcamento: mensagemOrcamento, LinhaCoresNome: linhaCoresNome, LinhaCoresLink: linhaCoresLink,
		PrecoBase: precoBase, DescontoCartao: descontoCartao, DescontoPix: descontoPix,
		DescontoPixAvista: descontoPixAvista, PixParcelado: pixParcelado, Brinde: brinde,
		IncluirBrinde: incluirBrinde, PedidoPrincipalID: pedidoPrincipalID,
	}
	if err := c.dao.AtualizarDetalhes(id, pedidoEdicao); err != nil {
		if errors.Is(err, daos.ErrPedidoNaoEncontrado) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, daos.ErrPedidoPrincipalNaoEncontrado) || errors.Is(err, daos.ErrPedidoPrincipalInvalido) {
			http.Error(w, "Pedido principal inválido.", http.StatusBadRequest)
			return
		}
		if errors.Is(err, daos.ErrCondicoesPedidoBloqueadas) {
			http.Error(w, "Não é possível alterar as condições deste pedido depois que os pagamentos foram gerados.", http.StatusConflict)
			return
		}
		http.Error(w, "Erro ao atualizar o pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
}

func montarMensagemOrcamento(clienteNome string, produtoNome string, preco float64, diasUteis int, entrega time.Time, linhaNome string, linhaLink string, temDesconto bool, percentualDesconto float64, temPixParcelado bool, incluirBrinde bool, temDescontoPixParcelado bool, percentualDescontoPixParcelado float64, temDescontoPixAvista bool) string {
	valorCartao := preco
	descricaoCartao := views.FormatarReais(valorCartao)
	if temDesconto {
		valorDesconto := math.Round(preco*percentualDesconto) / 100
		valorCartao = preco - valorDesconto
		percentualFormatado := strings.ReplaceAll(strconv.FormatFloat(percentualDesconto, 'f', -1, 64), ".", ",")
		descricaoCartao = fmt.Sprintf("%s\n%s%% de desconto", views.FormatarReais(valorCartao), percentualFormatado)
	}

	valorPix := preco
	descricaoPix := views.FormatarReais(valorPix)
	if temDescontoPixAvista {
		valorPix = math.Round(preco*0.90*100) / 100
		descricaoPix = fmt.Sprintf("%s\n10%% de desconto", views.FormatarReais(valorPix))
	}
	sinalCartaoPix := math.Round(valorCartao*0.30*100) / 100
	restanteCartao := valorCartao - sinalCartaoPix
	valorPixParcelado := preco
	descricaoPixParcelado := views.FormatarReais(valorPixParcelado)
	if temDescontoPixParcelado {
		valorDescontoPixParcelado := math.Round(preco*percentualDescontoPixParcelado) / 100
		valorPixParcelado = preco - valorDescontoPixParcelado
		percentualFormatado := strings.ReplaceAll(strconv.FormatFloat(percentualDescontoPixParcelado, 'f', -1, 64), ".", ",")
		descricaoPixParcelado = fmt.Sprintf("%s\n%s%% de desconto", views.FormatarReais(valorPixParcelado), percentualFormatado)
	}
	sinalPixParcelado := math.Round(valorPixParcelado*0.30*100) / 100
	restantePixParcelado := valorPixParcelado - sinalPixParcelado

	mensagem := fmt.Sprintf("oi, %s! seguem os detalhes do orçamento para a sua peça:\n\n🌸 *peça:* %s\n💰 *valor:* %s\n\n💳 *formas de pagamento*\n\n*pix à vista*\n%s\n\n*cartão de crédito*\n%s\n→ sinal de 30%% via pix: %s\n→ restante: %s em até 3x sem juros.", clienteNome, produtoNome, views.FormatarReais(preco), descricaoPix, descricaoCartao, views.FormatarReais(sinalCartaoPix), views.FormatarReais(restanteCartao))
	if temPixParcelado {
		mensagem += fmt.Sprintf("\n\n*pix parcelado*\n%s\n→ sinal de 30%% via pix: %s\n→ restante: %s em até 3 parcelas via pix.", descricaoPixParcelado, views.FormatarReais(sinalPixParcelado), views.FormatarReais(restantePixParcelado))
	}
	if incluirBrinde {
		mensagem += "\n\n🎁 *brinde:* a compra inclui um pequeno brinde nas cores escolhidas para a sua peça."
	}
	mensagem += fmt.Sprintf("\n\n⏳ *prazo de produção:* %d dias úteis, podendo ser entregue antes.\n📅 *previsão de entrega:* %s\n\neste orçamento é válido por 15 dias.", diasUteis, entrega.Format("02/01/2006"))
	if strings.TrimSpace(linhaNome) != "" && strings.TrimSpace(linhaLink) != "" {
		mensagem += fmt.Sprintf("\n\n🎨 *cores disponíveis*\nescolha a cor da sua peça pela tabela oficial de cores da linha %s:\n%s", strings.TrimSpace(linhaNome), strings.TrimSpace(linhaLink))
	}
	return mensagem + "\n\nse tiver dúvidas na hora de escolher as cores, é só me chamar! 🌟\n\npodemos confirmar o seu pedido? 🧶✨"
}

func (c *PedidoControlador) AtualizarStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.FormValue("pedido_id"))
	if err != nil || id < 1 {
		http.Error(w, "Pedido inválido.", http.StatusBadRequest)
		return
	}
	status := r.FormValue("status")
	primeiraData := r.FormValue("primeira_data")
	if primeiraData == "" {
		primeiraData = time.Now().Format("2006-01-02")
	}
	dataRecebimento := r.FormValue("data_recebimento")
	if status == modelos.StatusSinalRecebido {
		if r.FormValue("confirmado") != "sim" {
			http.Error(w, "Confirme que recebeu o sinal antes de avançar o pedido.", http.StatusBadRequest)
			return
		}
		if _, err := time.Parse("2006-01-02", dataRecebimento); err != nil {
			http.Error(w, "Informe uma data de recebimento válida.", http.StatusBadRequest)
			return
		}
	}
	numeroParcelas := 1
	if valor := r.FormValue("numero_parcelas"); valor != "" {
		numeroParcelas, err = strconv.Atoi(valor)
		if err != nil {
			http.Error(w, "Informe um número de parcelas válido.", http.StatusBadRequest)
			return
		}
	}
	if status == modelos.StatusSinalRecebido {
		err = c.dao.AvancarStatusComRecebimentoSinal(id, r.FormValue("forma_pagamento"), numeroParcelas, primeiraData, dataRecebimento)
	} else {
		err = c.dao.AtualizarStatusComPagamento(id, status, "", numeroParcelas, primeiraData)
	}
	if err != nil {
		if errors.Is(err, daos.ErrTransicaoPedidoInvalida) {
			http.Error(w, "Essa etapa não pode ser selecionada agora.", http.StatusBadRequest)
			return
		}
		if errors.Is(err, daos.ErrPedidoNaoEncontrado) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, daos.ErrFormaPagamentoInvalida) ||
			errors.Is(err, daos.ErrPixParceladoNaoOferecido) ||
			errors.Is(err, daos.ErrNumeroParcelasInvalido) ||
			errors.Is(err, daos.ErrDataParcelaInvalida) ||
			errors.Is(err, daos.ErrValorTotalInvalido) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Erro ao atualizar o pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
}

func (c *PedidoControlador) ReceberPagamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.FormValue("pagamento_id"))
	if err != nil || id < 1 {
		http.Error(w, "Pagamento inválido.", http.StatusBadRequest)
		return
	}
	if r.FormValue("confirmado") != "sim" {
		http.Error(w, "Confirme que o pagamento foi recebido antes de lançar no caixa.", http.StatusBadRequest)
		return
	}
	dataRecebimento := r.FormValue("data_recebimento")
	if _, err := time.Parse("2006-01-02", dataRecebimento); err != nil {
		http.Error(w, "Informe uma data de recebimento válida.", http.StatusBadRequest)
		return
	}
	if err := c.dao.RegistrarRecebimento(id, dataRecebimento); err != nil {
		if errors.Is(err, daos.ErrPagamentoNaoEncontrado) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, daos.ErrPagamentoJaRecebido) {
			http.Error(w, "Este pagamento já foi registrado como recebido.", http.StatusConflict)
			return
		}
		http.Error(w, "Erro ao registrar recebimento.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
}

func (c *PedidoControlador) AdicionarEtiqueta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	pedidoID, err := strconv.Atoi(r.FormValue("pedido_id"))
	if err != nil || pedidoID < 1 {
		http.Error(w, "Pedido inválido.", http.StatusBadRequest)
		return
	}
	texto := strings.TrimSpace(r.FormValue("texto"))
	if texto == "" || len([]rune(texto)) > 100 {
		http.Error(w, "A etiqueta deve ter entre 1 e 100 caracteres.", http.StatusBadRequest)
		return
	}
	cor := r.FormValue("cor")
	if !corEtiquetaPedidoValida(cor) {
		http.Error(w, "Selecione uma cor válida para a etiqueta.", http.StatusBadRequest)
		return
	}
	if err := c.dao.AdicionarEtiquetaPedido(pedidoID, texto, cor); err != nil {
		if errors.Is(err, daos.ErrPedidoNaoEncontrado) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao adicionar etiqueta ao pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos?detalhes="+strconv.Itoa(pedidoID), http.StatusSeeOther)
}

func (c *PedidoControlador) ExcluirEtiqueta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	pedidoID, err := strconv.Atoi(r.FormValue("pedido_id"))
	if err != nil || pedidoID < 1 {
		http.Error(w, "Pedido inválido.", http.StatusBadRequest)
		return
	}
	etiquetaID, err := strconv.Atoi(r.FormValue("etiqueta_id"))
	if err != nil || etiquetaID < 1 {
		http.Error(w, "Etiqueta inválida.", http.StatusBadRequest)
		return
	}
	if err := c.dao.ExcluirEtiquetaPedido(pedidoID, etiquetaID); err != nil {
		if errors.Is(err, daos.ErrEtiquetaPedidoNaoEncontrada) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao remover etiqueta do pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos?detalhes="+strconv.Itoa(pedidoID), http.StatusSeeOther)
}

func corEtiquetaPedidoValida(cor string) bool {
	if len(cor) != 7 || cor[0] != '#' {
		return false
	}
	_, err := strconv.ParseUint(cor[1:], 16, 24)
	return err == nil
}

func (c *PedidoControlador) Listar(w http.ResponseWriter, r *http.Request) {
	pedidos, err := c.dao.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar pedidos.", http.StatusInternalServerError)
		return
	}

	clientes, _ := c.clienteDAO.BuscarTodos()
	produtos, _ := c.produtoDAO.BuscarTodos()
	for i := range pedidos {
		if pedidos[i].Status == modelos.StatusAguardandoPagamento && todosPagamentosRecebidos(pedidos[i].Pagamentos) {
			pedidos[i].Status = modelos.StatusPago
		}
		clienteNome := "Desconhecido"
		for _, cliente := range clientes {
			if cliente.ID == pedidos[i].ClienteID {
				clienteNome = cliente.Nome
				break
			}
		}
		itens := itensPedidoDTO(pedidos[i].Itens)
		produtoNome := nomesItensPedidoDTO(itens)
		produtoPreco := 0.0
		for _, item := range itens {
			produtoPreco += float64(item.Quantidade) * item.PrecoUnitario
		}
		if len(itens) == 0 {
			produtoNome = "Desconhecido"
			for _, produto := range produtos {
				if produto.ID == pedidos[i].ProdutoID {
					produtoNome = produto.NomePeca
					produtoPreco = produto.Preco
					break
				}
			}
		}
		for j := range pedidos[i].Pagamentos {
			if pedidos[i].Pagamentos[j].Status == "Recebido" {
				continue
			}
			pedidos[i].Pagamentos[j].MensagemLembrete = montarMensagemLembrete(clienteNome, produtoNome, pedidos[i].Pagamentos[j])
		}
		if pedidos[i].MensagemOrcamento != "" {
			continue
		}
		dataEntrega, err := time.Parse("2006-01-02", pedidos[i].DataEntrega)
		if err != nil {
			dataEntrega = time.Now()
		}
		diasPrazo := pedidos[i].DiasPrazo
		if diasPrazo <= 0 {
			diasPrazo = prazoPadraoDiasUteis
		}
		pedidos[i].MensagemOrcamento = montarMensagemOrcamento(clienteNome, produtoNome, produtoPreco, diasPrazo, dataEntrega, "", "", false, 0, false, false, false, 0, pedidos[i].DescontoPixAvista)
	}

	componente := views.AdminLayout("Gestão de Pedidos", views.PedidosLista(pedidos, clientes, produtos))
	componente.Render(r.Context(), w)
}

func (c *PedidoControlador) ImprimirOrcamento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	pedidoID, err := strconv.Atoi(r.URL.Query().Get("pedido_id"))
	if err != nil || pedidoID < 1 {
		http.Error(w, "Pedido inválido.", http.StatusBadRequest)
		return
	}

	pedidos, err := c.dao.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar pedido.", http.StatusInternalServerError)
		return
	}
	var pedido modelos.Pedido
	for _, item := range pedidos {
		if item.ID == pedidoID {
			pedido = item
			break
		}
	}
	if pedido.ID == 0 {
		http.NotFound(w, r)
		return
	}

	clientes, err := c.clienteDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar cliente.", http.StatusInternalServerError)
		return
	}
	produtos, err := c.produtoDAO.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar produto.", http.StatusInternalServerError)
		return
	}
	var cliente modelos.Cliente
	for _, item := range clientes {
		if item.ID == pedido.ClienteID {
			cliente = item
			break
		}
	}
	var produto modelos.Produto
	for _, item := range produtos {
		if item.ID == pedido.ProdutoID {
			produto = item
			break
		}
	}
	if cliente.ID == 0 || produto.ID == 0 {
		http.Error(w, "Cliente ou produto do pedido não encontrado.", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.OrcamentoImpressao(pedido, cliente, produto, os.Getenv("WHATSAPP_NUMERO")).Render(r.Context(), w); err != nil {
		http.Error(w, "Erro ao gerar orçamento.", http.StatusInternalServerError)
	}
}

func calcularDataEntrega(inicio time.Time, diasUteis int) time.Time {
	data := inicio
	for diasUteis > 0 {
		data = data.AddDate(0, 0, 1)
		if data.Weekday() != time.Saturday && data.Weekday() != time.Sunday {
			diasUteis--
		}
	}
	return data
}

func formatarDataParaMensagem(data string) string {
	dataConvertida, err := time.Parse("2006-01-02", data)
	if err != nil {
		return data
	}
	return dataConvertida.Format("02/01/2006")
}

func montarMensagemLembrete(clienteNome string, produtoNome string, pagamento modelos.Pagamento) string {
	return fmt.Sprintf(
		"oi, %s! passando para lembrar do pagamento de %s da peça *%s*, no valor de %s, com vencimento em %s. Se já realizou, pode desconsiderar. Obrigada! 🌸",
		clienteNome, pagamento.Descricao, produtoNome, views.FormatarReais(pagamento.Valor), formatarDataParaMensagem(pagamento.DataVencimento))
}

func todosPagamentosRecebidos(pagamentos []modelos.Pagamento) bool {
	if len(pagamentos) == 0 {
		return false
	}
	for _, pagamento := range pagamentos {
		if pagamento.Status != "Recebido" {
			return false
		}
	}
	return true
}
