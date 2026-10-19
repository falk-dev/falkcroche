package controladores

import (
	"errors"
	"falkcroche/daos"
	"falkcroche/dtos"
	"falkcroche/views"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TransacaoControlador struct {
	dao *daos.TransacaoDAO
}

func NovoTransacaoControlador(dao *daos.TransacaoDAO) *TransacaoControlador {
	return &TransacaoControlador{dao: dao}
}

func (c *TransacaoControlador) Inserir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	dto, err := transacaoInput(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := c.dao.Inserir(dto); err != nil {
		http.Error(w, "Erro ao registrar movimentação financeira.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/financeiro", http.StatusSeeOther)
}

func (c *TransacaoControlador) Atualizar(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Movimentação inválida.", http.StatusBadRequest)
		return
	}
	dto, err := transacaoInput(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.dao.Atualizar(id, dto); err != nil {
		if errors.Is(err, daos.ErrTransacaoNaoEncontrada) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, daos.ErrTipoTransacaoInvalido) {
			http.Error(w, "Recebimentos de pedidos devem permanecer como entrada.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Erro ao atualizar movimentação financeira.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/financeiro", http.StatusSeeOther)
}

func transacaoInput(r *http.Request) (dtos.TransacaoInputDTO, error) {
	if err := r.ParseForm(); err != nil {
		return dtos.TransacaoInputDTO{}, errors.New("Erro ao processar formulário.")
	}
	tipo := r.FormValue("tipo")
	if tipo != "Entrada" && tipo != "Saída" {
		return dtos.TransacaoInputDTO{}, errors.New("Selecione um tipo de movimentação válido.")
	}
	descricao := strings.TrimSpace(r.FormValue("descricao"))
	if descricao == "" {
		return dtos.TransacaoInputDTO{}, errors.New("Informe uma descrição.")
	}
	valor, err := strconv.ParseFloat(r.FormValue("valor"), 64)
	if err != nil || math.IsNaN(valor) || math.IsInf(valor, 0) || valor <= 0 {
		return dtos.TransacaoInputDTO{}, errors.New("Informe um valor maior que zero.")
	}
	data := r.FormValue("data")
	if _, err := time.Parse("2006-01-02", data); err != nil {
		return dtos.TransacaoInputDTO{}, errors.New("Informe uma data válida.")
	}
	return dtos.TransacaoInputDTO{Tipo: tipo, Descricao: descricao, Valor: valor, Data: data}, nil
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
