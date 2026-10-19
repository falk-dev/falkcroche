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
	notificacaoDAO := daos.NovoNotificacaoDAO(db)

	clienteCtrl := controladores.NovoClienteControlador(clienteDAO)
	produtoCtrl := controladores.NovoProdutoControlador(produtoDAO)
	pedidoCtrl := controladores.NovoPedidoControlador(pedidoDAO, clienteDAO, produtoDAO)
	transacaoCtrl := controladores.NovoTransacaoControlador(transacaoDAO)
	notificacaoCtrl := controladores.NovoNotificacaoControlador(notificacaoDAO)
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
	mux.Handle("/pedidos/atualizar", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.Atualizar)))
	mux.Handle("/pedidos/status", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.AtualizarStatus)))
	mux.Handle("/pedidos/pagamentos/receber", acessoCtrl.Proteger(http.HandlerFunc(pedidoCtrl.ReceberPagamento)))

	mux.Handle("/financeiro", acessoCtrl.Proteger(http.HandlerFunc(transacaoCtrl.Listar)))
	mux.Handle("/financeiro/inserir", acessoCtrl.Proteger(http.HandlerFunc(transacaoCtrl.Inserir)))
	mux.Handle("/financeiro/atualizar", acessoCtrl.Proteger(http.HandlerFunc(transacaoCtrl.Atualizar)))
	mux.Handle("/notificacoes", acessoCtrl.Proteger(http.HandlerFunc(notificacaoCtrl.Listar)))
	mux.Handle("/notificacoes/ler", acessoCtrl.Proteger(http.HandlerFunc(notificacaoCtrl.MarcarComoLida)))

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
		componente := views.PublicLayout("Vitrine", views.VitrinePublica(produtos, os.Getenv("WHATSAPP_NUMERO")), "/static/logo.png", "/static")
		componente.Render(r.Context(), w)
	})

	host := os.Getenv("APP_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	endereco := host + ":" + port

	if endereco == "" {
		endereco = "0.0.0.0:8080"
	}
	fmt.Printf("Servidor rodando em http://%s\n", endereco)
	if err := http.ListenAndServe(endereco, mux); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
