package dtos

type PedidoInputDTO struct {
	ClienteID         int
	ProdutoID         int
	Status            string
	DataPedido        string
	DiasPrazo         int
	DataEntrega       string
	Observacoes       string
	MensagemOrcamento string
}
