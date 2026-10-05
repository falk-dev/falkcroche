package controladores

import (
	"database/sql"
	"net/http"
	"strconv"

	"falkcroche/daos"
	"falkcroche/dtos"
	"falkcroche/views"
)

// ProdutoControlador lida com as requisições web (rotas) relacionadas aos produtos.
type ProdutoControlador struct {
	dao *daos.ProdutoDAO
}

func NovoProdutoControlador(dao *daos.ProdutoDAO) *ProdutoControlador {
	return &ProdutoControlador{dao: dao}
}

func (c *ProdutoControlador) Inserir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}

	preco, _ := strconv.ParseFloat(r.FormValue("preco"), 64)
	prontaEntrega := r.FormValue("pronta_entrega") == "on"
	publicado := r.FormValue("publicado") == "on"

	produtoDTO := dtos.ProdutoInputDTO{
		NomePeca:      r.FormValue("nome_peca"),
		Preco:         preco,
		FotoURL:       r.FormValue("foto_url"),
		ProntaEntrega: prontaEntrega,
		Publicado:     publicado,
	}

	c.dao.Inserir(produtoDTO)
	http.Redirect(w, r, "/produtos", http.StatusSeeOther)
}

func (c *ProdutoControlador) Atualizar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário.", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id < 1 {
		http.Error(w, "Produto inválido.", http.StatusBadRequest)
		return
	}
	preco, err := strconv.ParseFloat(r.FormValue("preco"), 64)
	if err != nil || preco < 0 {
		http.Error(w, "Preço inválido.", http.StatusBadRequest)
		return
	}
	dto := dtos.ProdutoInputDTO{
		NomePeca:      r.FormValue("nome_peca"),
		Preco:         preco,
		FotoURL:       r.FormValue("foto_url"),
		ProntaEntrega: r.FormValue("pronta_entrega") == "on",
		Publicado:     r.FormValue("publicado") == "on",
	}
	if err := c.dao.Atualizar(id, dto); err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao atualizar produto.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/produtos", http.StatusSeeOther)
}

// Listar busca todos os produtos do banco.
func (c *ProdutoControlador) Listar(w http.ResponseWriter, r *http.Request) {
	produtos, err := c.dao.BuscarTodos()
	if err != nil {
		http.Error(w, "Erro ao buscar a lista de produtos.", http.StatusInternalServerError)
		return
	}

	// Renderiza a Vitrine de Produtos (Templ)
	componente := views.AdminLayout("Catálogo de Produtos", views.Produtos(produtos))
	componente.Render(r.Context(), w)
}
