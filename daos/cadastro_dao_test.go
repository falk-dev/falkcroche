package daos

import (
	"database/sql"
	"testing"

	"falkcroche/dtos"

	_ "github.com/mattn/go-sqlite3"
)

func TestAtualizarCliente(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE clientes (id INTEGER PRIMARY KEY, nome TEXT, whatsapp TEXT, cidade TEXT, estado TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO clientes (id, nome, whatsapp, cidade, estado) VALUES (1, 'Ana', '', '', '')`); err != nil {
		t.Fatal(err)
	}

	dao := NovoClienteDAO(db)
	if err := dao.Atualizar(1, dtos.ClienteInputDTO{Nome: "Ana Lima", Whatsapp: "55119999", Cidade: "Campinas", Estado: "SP"}); err != nil {
		t.Fatal(err)
	}
	clientes, err := dao.BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	if len(clientes) != 1 || clientes[0].Nome != "Ana Lima" || clientes[0].Cidade != "Campinas" || clientes[0].Estado != "SP" {
		t.Fatalf("cliente não atualizado: %+v", clientes)
	}
	if err := dao.Atualizar(2, dtos.ClienteInputDTO{}); err != sql.ErrNoRows {
		t.Fatalf("atualização de cliente inexistente = %v; esperado sql.ErrNoRows", err)
	}
}

func TestAtualizarProduto(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE categorias_produtos (nome TEXT PRIMARY KEY COLLATE NOCASE, prioridade TEXT NOT NULL DEFAULT 'media')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE produtos (id INTEGER PRIMARY KEY, nome_peca TEXT, categoria TEXT NOT NULL DEFAULT '', preco REAL, foto_url TEXT, pronta_entrega BOOLEAN, publicado BOOLEAN, imagem_ilustrativa BOOLEAN NOT NULL DEFAULT 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO produtos (id, nome_peca, preco, foto_url, pronta_entrega, publicado) VALUES (1, 'Bolsa', 50, '', 0, 0)`); err != nil {
		t.Fatal(err)
	}

	dao := NovoProdutoDAO(db)
	if err := dao.CriarCategoria("Bolsas", "alta"); err != nil {
		t.Fatal(err)
	}
	categorias, err := dao.BuscarCategorias()
	if err != nil {
		t.Fatal(err)
	}
	if len(categorias) != 1 || categorias[0].Nome != "Bolsas" || categorias[0].Prioridade != "alta" {
		t.Fatalf("categoria cadastrada antes dos produtos = %+v", categorias)
	}
	if err := dao.Atualizar(1, dtos.ProdutoInputDTO{NomePeca: "Bolsa floral", Categoria: "Bolsas", Preco: 75.5, FotoURL: "https://exemplo.com/bolsa.jpg", ProntaEntrega: true, Publicado: true}); err != nil {
		t.Fatal(err)
	}
	produtos, err := dao.BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	if len(produtos) != 1 || produtos[0].NomePeca != "Bolsa floral" || produtos[0].Categoria != "Bolsas" || produtos[0].PrioridadeCategoria != "alta" || produtos[0].Preco != 75.5 || !produtos[0].ProntaEntrega || !produtos[0].Publicado {
		t.Fatalf("produto não atualizado: %+v", produtos)
	}
	if err := dao.Atualizar(2, dtos.ProdutoInputDTO{}); err != sql.ErrNoRows {
		t.Fatalf("atualização de produto inexistente = %v; esperado sql.ErrNoRows", err)
	}
}

func TestPrioridadeDaCategoriaEhCompartilhadaEntreProdutos(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		`CREATE TABLE categorias_produtos (nome TEXT PRIMARY KEY COLLATE NOCASE, prioridade TEXT NOT NULL DEFAULT 'media')`,
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY AUTOINCREMENT, nome_peca TEXT, categoria TEXT NOT NULL DEFAULT '', preco REAL, foto_url TEXT, pronta_entrega BOOLEAN, publicado BOOLEAN, imagem_ilustrativa BOOLEAN NOT NULL DEFAULT 0)`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	dao := NovoProdutoDAO(db)
	for _, produto := range []dtos.ProdutoInputDTO{
		{NomePeca: "Bolsa A", Categoria: "Bolsas"},
		{NomePeca: "Bolsa B", Categoria: "Bolsas"},
	} {
		if _, err := dao.Inserir(produto); err != nil {
			t.Fatal(err)
		}
	}
	if err := dao.AtualizarPrioridadeCategoria("Bolsas", "baixa"); err != nil {
		t.Fatal(err)
	}
	if err := dao.Atualizar(2, dtos.ProdutoInputDTO{NomePeca: "Bolsa B", Categoria: "Bolsas"}); err != nil {
		t.Fatal(err)
	}
	produtos, err := dao.BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	for _, produto := range produtos {
		if produto.PrioridadeCategoria != "baixa" {
			t.Fatalf("prioridade do produto %q = %q; esperada baixa após atualizar o produto", produto.NomePeca, produto.PrioridadeCategoria)
		}
	}
}

func TestBuscarTodosOrdenaProdutosPorCategoria(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		`CREATE TABLE categorias_produtos (nome TEXT PRIMARY KEY COLLATE NOCASE, prioridade TEXT NOT NULL DEFAULT 'media')`,
		`CREATE TABLE produtos (id INTEGER PRIMARY KEY, nome_peca TEXT, categoria TEXT NOT NULL DEFAULT '', preco REAL, foto_url TEXT, pronta_entrega BOOLEAN, publicado BOOLEAN, imagem_ilustrativa BOOLEAN NOT NULL DEFAULT 0)`,
		`INSERT INTO produtos (id, nome_peca, categoria, preco, foto_url, pronta_entrega, publicado) VALUES (1, 'Bolsa', 'Bolsas', 10, '', 0, 0), (2, 'Chapéu', 'Chapéus', 10, '', 0, 0), (3, 'Boina vermelha', 'Boina', 10, '', 0, 0), (4, 'Boina azul', 'Boina', 10, '', 0, 0), (5, 'Case', '', 10, '', 0, 0)`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}

	produtos, err := NovoProdutoDAO(db).BuscarTodos()
	if err != nil {
		t.Fatal(err)
	}
	nomesEsperados := []string{"Boina azul", "Boina vermelha", "Bolsa", "Chapéu", "Case"}
	if len(produtos) != len(nomesEsperados) {
		t.Fatalf("quantidade de produtos = %d; esperada %d", len(produtos), len(nomesEsperados))
	}
	for indice, nome := range nomesEsperados {
		if produtos[indice].NomePeca != nome {
			t.Fatalf("produto na posição %d = %q; esperado %q", indice, produtos[indice].NomePeca, nome)
		}
	}
}
