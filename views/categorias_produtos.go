package views

import (
	"fmt"
	"sort"
	"strings"

	"falkcroche/modelos"
)

type grupoProdutos struct {
	chave    string
	ID       string
	Nome     string
	Produtos []modelos.Produto
}

func agruparProdutos(produtos []modelos.Produto) []grupoProdutos {
	ordenados := append([]modelos.Produto(nil), produtos...)
	sort.SliceStable(ordenados, func(i int, j int) bool {
		categoriaI := strings.TrimSpace(ordenados[i].Categoria)
		categoriaJ := strings.TrimSpace(ordenados[j].Categoria)
		if categoriaI == "" && categoriaJ != "" {
			return false
		}
		if categoriaI != "" && categoriaJ == "" {
			return true
		}
		if !strings.EqualFold(categoriaI, categoriaJ) {
			return strings.ToLower(categoriaI) < strings.ToLower(categoriaJ)
		}
		return strings.ToLower(ordenados[i].NomePeca) < strings.ToLower(ordenados[j].NomePeca)
	})

	grupos := make([]grupoProdutos, 0)
	for _, produto := range ordenados {
		chave := strings.ToLower(strings.TrimSpace(produto.Categoria))
		categoria := strings.TrimSpace(produto.Categoria)
		if categoria == "" {
			categoria = "Sem categoria"
		}
		if len(grupos) == 0 || grupos[len(grupos)-1].chave != chave {
			grupos = append(grupos, grupoProdutos{
				chave: chave,
				ID:    fmt.Sprintf("categoria-%d", len(grupos)),
				Nome:  categoria,
			})
		}
		indice := len(grupos) - 1
		grupos[indice].Produtos = append(grupos[indice].Produtos, produto)
	}
	return grupos
}
