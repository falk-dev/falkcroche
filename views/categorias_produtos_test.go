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

func TestAgruparProdutosPriorizaAltaMediaBaixa(t *testing.T) {
	grupos := agruparProdutosPorPrioridade([]modelos.Produto{
		{NomePeca: "Bolsa", Categoria: "Bolsas", PrioridadeCategoria: "baixa"},
		{NomePeca: "Boina", Categoria: "Boina", PrioridadeCategoria: "alta"},
		{NomePeca: "Amigurumi", Categoria: "Amigurumi", PrioridadeCategoria: "media"},
	})
	for indice, esperado := range []string{"Boina", "Amigurumi", "Bolsas"} {
		if grupos[indice].Nome != esperado {
			t.Fatalf("categoria na posição %d = %q; esperado %q", indice, grupos[indice].Nome, esperado)
		}
	}
}

func TestAgruparProdutosMantemCategoriasEmOrdemAlfabetica(t *testing.T) {
	grupos := agruparProdutos([]modelos.Produto{
		{NomePeca: "Bolsa", Categoria: "Bolsas", PrioridadeCategoria: "alta"},
		{NomePeca: "Boina", Categoria: "Boina", PrioridadeCategoria: "baixa"},
		{NomePeca: "Amigurumi", Categoria: "Amigurumi", PrioridadeCategoria: "media"},
	})
	for indice, esperado := range []string{"Amigurumi", "Boina", "Bolsas"} {
		if grupos[indice].Nome != esperado {
			t.Fatalf("categoria na posição %d = %q; esperado %q", indice, grupos[indice].Nome, esperado)
		}
	}
}

func TestClassePrioridadeCategoriaDefineCoresDistintas(t *testing.T) {
	coresEsperadas := map[string]string{
		"alta":  "bg-[#f6d7d4] text-[#87352d]",
		"media": "bg-[#f0dcb0] text-[#74551a]",
		"baixa": "bg-[#dce9d9] text-[#42613f]",
	}
	for prioridade, classe := range coresEsperadas {
		if obtida := classePrioridadeCategoria(prioridade); obtida != classe {
			t.Errorf("classe da prioridade %q = %q; esperada %q", prioridade, obtida, classe)
		}
	}
}
