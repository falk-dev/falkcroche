package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/joho/godotenv"

	"falkcroche/controladores"
	"falkcroche/daos"
	"falkcroche/database"
	"falkcroche/views"
)

func main() {
	fmt.Println("Iniciando o sistema Falk Crochê...")
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Aviso: não foi possível carregar .env: %v", err)
	}

	db := database.Conectar()

	// agenda o fechamento do banco quando main() terminar
	defer db.Close()
	fmt.Println("Banco de dados pronto para uso.")

	clienteDAO := daos.NovoClienteDAO(db)
	produtoDAO := daos.NovoProdutoDAO(db)
	pedidoDAO := daos.NovoPedidoDAO(db)
	transacaoDAO := daos.NovoTransacaoDAO(db)

	clienteCtrl := controladores.NovoClienteControlador(clienteDAO)
	produtoCtrl := controladores.NovoProdutoControlador(produtoDAO)
	pedidoCtrl := controladores.NovoPedidoControlador(pedidoDAO, clienteDAO, produtoDAO)
	transacaoCtrl := controladores.NovoTransacaoControlador(transacaoDAO)
	acessoCtrl := controladores.NovoAcessoControlador(
		os.Getenv("ADMIN_PASSWORD"),
		os.Getenv("ADMIN_SESSION_KEY"),
	)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/admin/login", acessoCtrl.Login)
	mux.Handle("/admin/sair", acessoCtrl.Proteger(http.HandlerFunc(acessoCtrl.Sair)))

	// ROTAS PRIVADAS
	mux.Handle("/clientes", acessoCtrl.Proteger(http.HandlerFunc(clienteCtrl.Listar)))
	mux.Handle("/clientes/inserir", acessoCtrl.Proteger(http.HandlerFunc(clienteCtrl.Inserir)))
	mux.Handle("/clientes/atualizar", acessoCtrl.Proteger(http.HandlerFunc(clienteCtrl.Atualizar)))

	mux.Handle("/produtos", acessoCtrl.Proteger(http.HandlerFunc(produtoCtrl.Listar)))
	mux.Handle("/produtos/inserir", acessoCtrl.Proteger(http.HandlerFunc(produtoCtrl.Inserir)))
	mux.Handle("/produtos/atualizar", acessoCtrl.Proteger(http.HandlerFunc(produtoCtrl.Atualizar)))

	mux.Handle("/pedidos", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.Listar)))
	mux.Handle("/pedidos/inserir", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.Inserir)))
	mux.Handle("/pedidos/status", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.AtualizarStatus)))

	mux.Handle("/financeiro", acessoCtrl.Proteger(http.HandlerFunc(transacaoCtrl.Listar)))
	mux.Handle("/financeiro/inserir", acessoCtrl.Proteger(http.HandlerFunc(transacaoCtrl.Inserir)))

	mux.Handle("/dashboard", acessoCtrl.Proteger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pedidos, _ := pedidoDAO.BuscarTodos()
		produtos, _ := produtoDAO.BuscarTodos()
		clientes, _ := clienteDAO.BuscarTodos()
		transacoes, _ := transacaoDAO.BuscarTodas()

		var saldo float64
		for _, t := range transacoes {
			if t.Tipo == "Entrada" {
				saldo += t.Valor
			} else {
				saldo -= t.Valor
			}
		}

		sort.Slice(pedidos, func(i int, j int) bool {
			return pedidos[i].ID > pedidos[j].ID
		})
		componente := views.AdminLayout("Visão Geral", views.Dashboard(len(pedidos), len(produtos), len(clientes), saldo, produtos, pedidos, clientes))
		componente.Render(r.Context(), w)
	})))

	// ROTA PÚBLICA
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		produtos, _ := produtoDAO.BuscarPublicados()
		componente := views.PublicLayout("Vitrine", views.VitrinePublica(produtos, os.Getenv("WHATSAPP_NUMERO")), "/static/logo.png")
		componente.Render(r.Context(), w)
	})

	endereco := os.Getenv("SERVER_ADDRESS")
	if endereco == "" {
		endereco = "127.0.0.1:8080"
	}
	fmt.Printf("Servidor rodando em http://%s\n", endereco)
	if err := http.ListenAndServe(endereco, mux); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
