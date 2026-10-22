package modelos

const (
	StatusOrcamento           = "Orçamento"
	StatusAguardandoSinal     = "Aguardando sinal"
	StatusSinalRecebido       = "Sinal recebido"
	StatusEmProducao          = "Em produção"
	StatusAguardandoPagamento = "Aguardando pagamento final"
	StatusPago                = "Pago"
	StatusEntregue            = "Entregue"
	StatusCancelado           = "Cancelado"
)

func ProximosStatusPedido(status string) []string {
	switch status {
	case StatusOrcamento:
		return []string{StatusAguardandoSinal, StatusCancelado}
	case StatusAguardandoSinal:
		return []string{StatusSinalRecebido, StatusCancelado}
	case StatusSinalRecebido:
		return []string{StatusEmProducao, StatusCancelado}
	case StatusEmProducao:
		return []string{StatusAguardandoPagamento, StatusEntregue, StatusCancelado}
	case StatusAguardandoPagamento:
		return []string{StatusEntregue, StatusCancelado}
	case StatusPago:
		return []string{StatusEntregue}
	default:
		return nil
	}
}

func ProximosStatusPedidoComBrinde(status string, brinde bool) []string {
	if !brinde {
		return ProximosStatusPedido(status)
	}
	if status == StatusEmProducao {
		return []string{StatusEntregue, StatusCancelado}
	}
	return nil
}

type Pedido struct {
	ID                int
	ClienteID         int
	ProdutoID         int
	Status            string
	DataPedido        string
	DiasPrazo         int
	DataEntrega       string
	Observacoes       string
	MensagemOrcamento string
	LinhaCoresNome    string
	LinhaCoresLink    string
	IncluirBrinde     bool
	PrecoBase         float64
	DescontoCartao    float64
	DescontoPix       float64
	DescontoPixAvista bool
	PixParcelado      bool
	FormaPagamento    string
	TotalAcordado     float64
	Brinde            bool
	PedidoPrincipalID int
	Pagamentos        []Pagamento
	Etiquetas         []EtiquetaPedido
	Itens             []ItemPedido
}

type ItemPedido struct {
	PedidoID      int
	ProdutoID     int
	NomeProduto   string
	Categoria     string
	Quantidade    int
	PrecoUnitario float64
}

type EtiquetaPedido struct {
	ID       int
	PedidoID int
	Texto    string
	Cor      string
}

type Pagamento struct {
	ID               int
	PedidoID         int
	Etapa            string
	Descricao        string
	Valor            float64
	DataVencimento   string
	DataRecebimento  string
	Status           string
	MensagemLembrete string
}
