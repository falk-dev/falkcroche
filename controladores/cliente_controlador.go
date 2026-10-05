package controladores

import (
	"database/sql"
	"net/http"
	"strconv"

	"falkcroche/daos"
	"falkcroche/dtos"
	"falkcroche/views"
)

type ClienteControlador struct {
	dao *daos.ClienteDAO
}

func NovoClienteControlador(dao *daos.ClienteDAO) *ClienteControlador {
	return &ClienteControlador{dao: dao}
}

func (c *ClienteControlador) Inserir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/clientes", http.StatusSeeOther)
		return
	}
	r.ParseForm()

	dto := dtos.ClienteInputDTO{
		Nome:     r.FormValue("nome"),
		Whatsapp: r.FormValue("whatsapp"),
		Cidade:   r.FormValue("cidade"),
		Estado:   r.FormValue("estado"),
	}

	c.dao.Inserir(dto)
	http.Redirect(w, r, "/clientes", http.StatusSeeOther)
}

func (c *ClienteControlador) Atualizar(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Cliente inválido.", http.StatusBadRequest)
		return
	}
	dto := dtos.ClienteInputDTO{
		Nome:     r.FormValue("nome"),
		Whatsapp: r.FormValue("whatsapp"),
		Cidade:   r.FormValue("cidade"),
		Estado:   r.FormValue("estado"),
	}
	if err := c.dao.Atualizar(id, dto); err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao atualizar cliente.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/clientes", http.StatusSeeOther)
}

func (c *ClienteControlador) Listar(w http.ResponseWriter, r *http.Request) {
	clientes, _ := c.dao.BuscarTodos()
	// Renderiza a página dentro do Layout base
	componente := views.AdminLayout("Meus Clientes", views.Clientes(clientes))
	componente.Render(r.Context(), w)
}
