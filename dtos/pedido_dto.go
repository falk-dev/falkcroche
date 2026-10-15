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
	PrecoBase         float64
	DescontoCartao    float64
	DescontoPix       float64
	DescontoPixAvista bool
	PixParcelado      bool
}
