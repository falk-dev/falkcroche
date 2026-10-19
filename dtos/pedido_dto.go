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
	LinhaCoresNome    string
	LinhaCoresLink    string
	PrecoBase         float64
	DescontoCartao    float64
	DescontoPix       float64
	DescontoPixAvista bool
	PixParcelado      bool
	Brinde            bool
	IncluirBrinde     bool
	PedidoPrincipalID int
}

type PedidoEdicaoDTO struct {
	ClienteID         int
	ProdutoID         int
	DataPedido        string
	DiasPrazo         int
	DataEntrega       string
	Observacoes       string
	MensagemOrcamento string
	LinhaCoresNome    string
	LinhaCoresLink    string
	PrecoBase         float64
	DescontoCartao    float64
	DescontoPix       float64
	DescontoPixAvista bool
	PixParcelado      bool
	Brinde            bool
	IncluirBrinde     bool
	PedidoPrincipalID int
}
