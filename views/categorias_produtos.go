package views

import (
	"fmt"
	"sort"
	"strings"

	"falkcroche/modelos"
)

type grupoProdutos struct {
	chave      string
	ID         string
	Nome       string
	Prioridade string
	Produtos   []modelos.Produto
}

func agruparProdutos(produtos []modelos.Produto) []grupoProdutos {
	return agruparProdutosComOrdenacao(produtos, false)
}

func agruparProdutosPorPrioridade(produtos []modelos.Produto) []grupoProdutos {
	return agruparProdutosComOrdenacao(produtos, true)
}

func agruparProdutosComOrdenacao(produtos []modelos.Produto, priorizarCategorias bool) []grupoProdutos {
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
		if priorizarCategorias {
			prioridadeI := ordemPrioridadeCategoria(ordenados[i].PrioridadeCategoria)
			prioridadeJ := ordemPrioridadeCategoria(ordenados[j].PrioridadeCategoria)
			if prioridadeI != prioridadeJ {
				return prioridadeI < prioridadeJ
			}
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
			prioridade := strings.ToLower(strings.TrimSpace(produto.PrioridadeCategoria))
			if ordemPrioridadeCategoria(prioridade) == 1 {
				prioridade = "media"
			}
			grupos = append(grupos, grupoProdutos{
				chave:      chave,
				ID:         fmt.Sprintf("categoria-%d", len(grupos)),
				Nome:       categoria,
				Prioridade: prioridade,
			})
		}
		indice := len(grupos) - 1
		grupos[indice].Produtos = append(grupos[indice].Produtos, produto)
	}
	return grupos
}

func ordemPrioridadeCategoria(prioridade string) int {
	switch strings.ToLower(strings.TrimSpace(prioridade)) {
	case "alta":
		return 0
	case "baixa":
		return 2
	default:
		return 1
	}
}

func classePrioridadeCategoria(prioridade string) string {
	switch strings.ToLower(strings.TrimSpace(prioridade)) {
	case "alta":
		return "bg-[#f6d7d4] text-[#87352d]"
	case "baixa":
		return "bg-[#dce9d9] text-[#42613f]"
	default:
		return "bg-[#f0dcb0] text-[#74551a]"
	}
}
