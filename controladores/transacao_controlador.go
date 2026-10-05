package controladores

import (
	"falkcroche/daos"
	"falkcroche/dtos"
	"falkcroche/views"
	"net/http"
	"strconv"
)

type TransacaoControlador struct {
	dao *daos.TransacaoDAO
}

func NovoTransacaoControlador(dao *daos.TransacaoDAO) *TransacaoControlador {
	return &TransacaoControlador{dao: dao}
}

func (c *TransacaoControlador) Inserir(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		valor, _ := strconv.ParseFloat(r.FormValue("valor"), 64)
		dto := dtos.TransacaoInputDTO{
			Tipo:      r.FormValue("tipo"),
			Descricao: r.FormValue("descricao"),
			Valor:     valor,
			Data:      r.FormValue("data"),
		}
		c.dao.Inserir(dto)
	}
	http.Redirect(w, r, "/financeiro", http.StatusSeeOther)
}

func (c *TransacaoControlador) Listar(w http.ResponseWriter, r *http.Request) {
	transacoes, _ := c.dao.BuscarTodas()
	
	var entradas, saidas float64
	for _, t := range transacoes {
		if t.Tipo == "Entrada" {
			entradas += t.Valor
		} else {
			saidas += t.Valor
		}
	}
	saldo := entradas - saidas
	
	componente := views.AdminLayout("Gestão Financeira", views.Financeiro(transacoes, entradas, saidas, saldo))
	componente.Render(r.Context(), w)
}

