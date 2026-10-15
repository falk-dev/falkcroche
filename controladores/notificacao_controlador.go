package controladores

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"falkcroche/daos"
)

type NotificacaoControlador struct {
	dao *daos.NotificacaoDAO
}

func NovoNotificacaoControlador(dao *daos.NotificacaoDAO) *NotificacaoControlador {
	return &NotificacaoControlador{dao: dao}
}

func (c *NotificacaoControlador) Listar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}
	notificacoes, err := c.dao.ListarVencidas(time.Now().Format("2006-01-02"))
	if err != nil {
		http.Error(w, "Erro ao buscar notificações.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(notificacoes); err != nil {
		http.Error(w, "Erro ao enviar notificações.", http.StatusInternalServerError)
	}
}

func (c *NotificacaoControlador) MarcarComoLida(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Notificação inválida.", http.StatusBadRequest)
		return
	}
	if err := c.dao.MarcarComoLida(id); err != nil {
		if errors.Is(err, daos.ErrNotificacaoNaoEncontrada) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Erro ao marcar notificação como lida.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
