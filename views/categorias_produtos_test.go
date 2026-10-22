package views

import (
	"testing"

	"falkcroche/modelos"
)

func TestAgruparProdutosOrdenaCategoriasESemCategoria(t *testing.T) {
	grupos := agruparProdutos([]modelos.Produto{
		{NomePeca: "Bolsa", Categoria: "Bolsa"},
		{NomePeca: "Sem categoria real", Categoria: "Sem categoria"},
		{NomePeca: "Boina", Categoria: "Boina"},
		{NomePeca: "Sem categoria", Categoria: ""},
	})
	nomesEsperados := []string{"Boina", "Bolsa", "Sem categoria", "Sem categoria"}
	for indice, nome := range nomesEsperados {
		if grupos[indice].Nome != nome {
			t.Fatalf("grupo na posição %d = %q; esperado %q", indice, grupos[indice].Nome, nome)
		}
	}
	if len(grupos[2].Produtos) != 1 || grupos[2].Produtos[0].NomePeca != "Sem categoria real" {
		t.Fatalf("categoria literal foi misturada com itens sem categoria: %+v", grupos[2])
	}
	if grupos[3].Produtos[0].NomePeca != "Sem categoria" {
		t.Fatalf("item sem categoria ficou no grupo errado: %+v", grupos[3])
	}
}
