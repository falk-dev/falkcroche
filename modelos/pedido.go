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
		return []string{StatusAguardandoPagamento, StatusCancelado}
	case StatusAguardandoPagamento:
		return []string{StatusPago, StatusCancelado}
	case StatusPago:
		return []string{StatusEntregue}
	default:
		return nil
	}
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
}
