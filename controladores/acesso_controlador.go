package controladores

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"falkcroche/views"
)

const nomeCookieSessao = "falkcroche_sessao"
const duracaoSessao = 12 * time.Hour

type AcessoControlador struct {
	senha string
	chave []byte
}

func NovoAcessoControlador(senha string, chave string) *AcessoControlador {
	return &AcessoControlador{senha: senha, chave: []byte(chave)}
}

func (c *AcessoControlador) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	if c.senha == "" || len(c.chave) < 32 {
		c.renderizarLogin(w, r, http.StatusServiceUnavailable, "Configure ADMIN_PASSWORD e ADMIN_SESSION_KEY com pelo menos 32 caracteres.")
		return
	}

	if r.Method == http.MethodGet {
		c.renderizarLogin(w, r, http.StatusOK, "")
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Não foi possível processar o formulário.", http.StatusBadRequest)
		return
	}

	senhaRecebida := []byte(r.FormValue("senha"))
	if subtle.ConstantTimeCompare(senhaRecebida, []byte(c.senha)) != 1 {
		c.renderizarLogin(w, r, http.StatusUnauthorized, "Senha incorreta.")
		return
	}

	agora := time.Now()
	valorCookie := c.assinar(strconv.FormatInt(agora.Unix(), 10))
	http.SetCookie(w, &http.Cookie{
		Name:     nomeCookieSessao,
		Value:    valorCookie,
		Path:     "/",
		Expires:  agora.Add(duracaoSessao),
		MaxAge:   int(duracaoSessao.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (c *AcessoControlador) Sair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     nomeCookieSessao,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (c *AcessoControlador) Proteger(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(nomeCookieSessao)
		if err != nil || !c.sessaoValida(cookie.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		proximo.ServeHTTP(w, r)
	})
}

func (c *AcessoControlador) sessaoValida(valor string) bool {
	partes := strings.Split(valor, ".")
	if len(partes) != 2 || len(c.chave) < 32 {
		return false
	}

	criadoEm, err := strconv.ParseInt(partes[0], 10, 64)
	if err != nil {
		return false
	}

	idade := time.Since(time.Unix(criadoEm, 0))
	if idade < 0 || idade > duracaoSessao {
		return false
	}

	assinaturaRecebida, err := hex.DecodeString(partes[1])
	if err != nil {
		return false
	}

	assinaturaEsperada := c.assinatura(partes[0])
	return hmac.Equal(assinaturaRecebida, assinaturaEsperada)
}

func (c *AcessoControlador) assinar(valor string) string {
	assinatura := c.assinatura(valor)
	return valor + "." + hex.EncodeToString(assinatura)
}

func (c *AcessoControlador) assinatura(valor string) []byte {
	h := hmac.New(sha256.New, c.chave)
	h.Write([]byte(valor))
	return h.Sum(nil)
}

func (c *AcessoControlador) renderizarLogin(w http.ResponseWriter, r *http.Request, status int, mensagem string) {
	w.WriteHeader(status)
	views.Login(mensagem).Render(r.Context(), w)
}
