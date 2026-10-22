package modelos

type Produto struct {
	ID                  int
	NomePeca            string
	Categoria           string
	PrioridadeCategoria string
	Preco               float64
	FotoURL             string
	ImagemIlustrativa   bool
	ProntaEntrega       bool
	Publicado           bool
}

type CategoriaProduto struct {
	Nome       string
	Prioridade string
}
