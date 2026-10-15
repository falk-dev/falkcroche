package modelos

type Notificacao struct {
	ID             int     `json:"id"`
	PagamentoID    int     `json:"pagamento_id"`
	PedidoID       int     `json:"pedido_id"`
	Cliente        string  `json:"cliente"`
	Produto        string  `json:"produto"`
	Descricao      string  `json:"descricao"`
	Valor          float64 `json:"valor"`
	DataVencimento string  `json:"data_vencimento"`
}
