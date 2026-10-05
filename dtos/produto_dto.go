package dtos

type ProdutoInputDTO struct {
	NomePeca      string
	Preco         float64
	FotoURL       string
	ProntaEntrega bool
	Publicado     bool
}
