package controladores

import (
	"errors"
	"fmt"
	"math"
	"net/http"
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

	clienteID, erroCliente := strconv.Atoi(r.FormValue("cliente_id"))
	produtoID, erroProduto := strconv.Atoi(r.FormValue("produto_id"))
	if erroCliente != nil || erroProduto != nil {
		http.Error(w, "Selecione um cliente e um produto.", http.StatusBadRequest)
		return
	}

	dataAtual := time.Now()
	dataEntregaCalculada := calcularDataEntrega(dataAtual, prazoPadraoDiasUteis)
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
	var produtoNome string
	var produtoPreco float64
	for _, produto := range produtos {
		if produto.ID == produtoID {
			produtoNome = produto.NomePeca
			produtoPreco = produto.Preco
			break
		}
	}
	if clienteNome == "" || produtoNome == "" {
		http.Error(w, "Cliente ou produto não encontrado.", http.StatusBadRequest)
		return
	}
	mensagem := montarMensagemOrcamento(clienteNome, produtoNome, produtoPreco, prazoPadraoDiasUteis, dataEntregaCalculada, r.FormValue("linha_cores_nome"), r.FormValue("linha_cores_link"))

	pedidoDTO := dtos.PedidoInputDTO{
		ClienteID:         clienteID,
		ProdutoID:         produtoID,
		Status:            modelos.StatusOrcamento,
		DataPedido:        dataAtual.Format("2006-01-02"),
		DiasPrazo:         prazoPadraoDiasUteis,
		DataEntrega:       dataEntregaCalculada.Format("2006-01-02"),
		Observacoes:       r.FormValue("observacoes"),
		MensagemOrcamento: mensagem,
	}

	if _, err := c.dao.Inserir(pedidoDTO); err != nil {
		http.Error(w, "Erro ao registrar pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
}

func montarMensagemOrcamento(clienteNome string, produtoNome string, preco float64, diasUteis int, entrega time.Time, linhaNome string, linhaLink string) string {
	valorPix := math.Round(preco*0.90*100) / 100
	sinalCartaoPix := math.Round(preco*0.30*100) / 100
	restanteCartao := preco - sinalCartaoPix
	parcelaCartao := math.Round(restanteCartao/3*100) / 100
	ultimaParcela := math.Round((restanteCartao-2*parcelaCartao)*100) / 100

	mensagem := fmt.Sprintf("oi, %s! seguem os detalhes do orçamento para a sua peça:\n\n🌸 peça: %s\n💰 valor: %s\n\n💳 formas de pagamento:\n• pix à vista: %s (10%% de desconto)\n• cartão de crédito: sinal de 30%% via pix (%s), e o restante (%s) pode ser parcelado em até 3x sem juros: 2x de %s e 1x de %s.\n\n⏳ prazo de produção: %d dias úteis (podendo ser entregue antes).\n📅 previsão de entrega: %s\n\neste orçamento é válido por 15 dias.", clienteNome, produtoNome, views.FormatarReais(preco), views.FormatarReais(valorPix), views.FormatarReais(sinalCartaoPix), views.FormatarReais(restanteCartao), views.FormatarReais(parcelaCartao), views.FormatarReais(ultimaParcela), diasUteis, entrega.Format("02-01-2006"))
	if strings.TrimSpace(linhaNome) != "" && strings.TrimSpace(linhaLink) != "" {
		mensagem += fmt.Sprintf("\n\n🎨 cores disponíveis:\nescolha a cor da sua peça pela tabela oficial de cores da linha %s:\n%s", strings.TrimSpace(linhaNome), strings.TrimSpace(linhaLink))
	}
	return mensagem + "\n\nse tiver dúvidas na hora de tirar as medidas ou escolher a cor, é só me chamar! 🌟\n\npodemos confirmar o seu pedido? 🧶✨"
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
	if err := c.dao.AtualizarStatus(id, r.FormValue("status")); err != nil {
		if errors.Is(err, daos.ErrTransicaoPedidoInvalida) {
			http.Error(w, "Essa etapa não pode ser selecionada agora.", http.StatusBadRequest)
			return
		}
		if errors.Is(err, daos.ErrPedidoNaoEncontrado) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao atualizar o pedido.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pedidos", http.StatusSeeOther)
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
		if pedidos[i].MensagemOrcamento != "" {
			continue
		}
		clienteNome := "Desconhecido"
		for _, cliente := range clientes {
			if cliente.ID == pedidos[i].ClienteID {
				clienteNome = cliente.Nome
				break
			}
		}
		produtoNome := "Desconhecido"
		var produtoPreco float64
		for _, produto := range produtos {
			if produto.ID == pedidos[i].ProdutoID {
				produtoNome = produto.NomePeca
				produtoPreco = produto.Preco
				break
			}
		}
		dataEntrega, err := time.Parse("2006-01-02", pedidos[i].DataEntrega)
		if err != nil {
			dataEntrega = time.Now()
		}
		diasPrazo := pedidos[i].DiasPrazo
		if diasPrazo <= 0 {
			diasPrazo = prazoPadraoDiasUteis
		}
		pedidos[i].MensagemOrcamento = montarMensagemOrcamento(clienteNome, produtoNome, produtoPreco, diasPrazo, dataEntrega, "", "")
	}

	componente := views.AdminLayout("Gestão de Pedidos", views.PedidosLista(pedidos, clientes, produtos))
	componente.Render(r.Context(), w)
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
