package modelos

type Produto struct {
	ID                  int
	NomePeca            string
	Categoria           string
	PrioridadeCategoria string
	Preco               float64
	FotoURL             string
	ProntaEntrega       bool
	Publicado           bool
}
