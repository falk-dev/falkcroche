package controladores

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProtegerRedirecionaSemSessao(t *testing.T) {
	controlador := NovoAcessoControlador("senha", strings.Repeat("k", 32))
	handler := controlador.Proteger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	resposta := httptest.NewRecorder()
	requisicao := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

	handler.ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusSeeOther {
		t.Fatalf("status = %d; esperado %d", resposta.Code, http.StatusSeeOther)
	}
	if resposta.Header().Get("Location") != "/admin/login" {
		t.Fatalf("destino = %q; esperado /admin/login", resposta.Header().Get("Location"))
	}
}

func TestLoginCriaSessaoAceitaPeloMiddleware(t *testing.T) {
	controlador := NovoAcessoControlador("senha", strings.Repeat("k", 32))
	formulario := url.Values{"senha": {"senha"}}
	requisicao := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(formulario.Encode()))
	requisicao.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resposta := httptest.NewRecorder()

	controlador.Login(resposta, requisicao)

	if resposta.Code != http.StatusSeeOther {
		t.Fatalf("status = %d; esperado %d", resposta.Code, http.StatusSeeOther)
	}
	cookie := resposta.Result().Cookies()[0]
	if !cookie.HttpOnly {
		t.Fatal("cookie de sessão não está marcado como HttpOnly")
	}

	protegido := controlador.Proteger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	requisicaoPainel := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	requisicaoPainel.AddCookie(cookie)
	respostaPainel := httptest.NewRecorder()
	protegido.ServeHTTP(respostaPainel, requisicaoPainel)

	if respostaPainel.Code != http.StatusNoContent {
		t.Fatalf("status com sessão válida = %d; esperado %d", respostaPainel.Code, http.StatusNoContent)
	}
}
