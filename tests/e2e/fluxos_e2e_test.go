package e2e

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

func TestFluxosAdministrativos(t *testing.T) {
	seleniumURL := os.Getenv("SELENIUM_URL")
	if seleniumURL == "" {
		seleniumURL = "http://localhost:4444/wd/hub"
	}
	driver, err := selenium.NewRemote(
		selenium.Capabilities{"browserName": "chrome"},
		seleniumURL,
	)
	if err != nil {
		t.Skipf("Selenium indisponível em %s: %v", seleniumURL, err)
	}
	t.Cleanup(func() { _ = driver.Quit() })
	if err := driver.SetImplicitWaitTimeout(5 * time.Second); err != nil {
		t.Fatal(err)
	}
	if err := driver.SetPageLoadTimeout(20 * time.Second); err != nil {
		t.Fatal(err)
	}

	baseURL := iniciarAppTeste(t)

	if err := driver.Get(baseURL + "/clientes"); err != nil {
		t.Fatal(err)
	}
	verificarURLContem(t, driver, "/admin/login")
	assertTextNaPagina(t, driver, "Área administrativa")

	preencher(t, driver, selenium.ByID, "senha", "senha-errada")
	clicar(t, driver, selenium.ByCSSSelector, "form[action='/admin/login'] button[type='submit']")
	assertTextNaPagina(t, driver, "Senha incorreta.")
	preencher(t, driver, selenium.ByID, "senha", "senha-e2e-segura")
	clicar(t, driver, selenium.ByCSSSelector, "form[action='/admin/login'] button[type='submit']")
	verificarURLContem(t, driver, "/dashboard")

	criarEEditarCliente(t, driver, baseURL)
	criarEEditarProduto(t, driver, baseURL)
	verificarVitrinePublica(t, driver, baseURL)
	criarEAvancarPedido(t, driver, baseURL)
	registrarMovimentacoesFinanceiras(t, driver, baseURL)

	if os.Getenv("E2E_HOLD") == "1" {
		t.Log("Chrome aberto no noVNC; pressione Enter aqui para encerrar o teste.")
		_, _ = fmt.Scanln()
	}

	clicar(t, driver, selenium.ByCSSSelector, "form[action='/admin/sair'] button[type='submit']")
	verificarURLContem(t, driver, "/admin/login")
	if err := driver.Get(baseURL + "/financeiro"); err != nil {
		t.Fatal(err)
	}
	verificarURLContem(t, driver, "/admin/login")
}

func iniciarAppTeste(t *testing.T) string {
	t.Helper()

	raiz, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	diretorioTemp := t.TempDir()
	executavel := filepath.Join(diretorioTemp, "falkcroche-e2e")
	compilar := exec.Command("go", "build", "-o", executavel, ".")
	compilar.Dir = raiz
	if saida, err := compilar.CombinedOutput(); err != nil {
		t.Fatalf("compilação do servidor E2E falhou: %v\n%s", err, saida)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	porta := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	endereco := net.JoinHostPort("0.0.0.0", strconv.Itoa(porta))
	urlLocal := "http://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(porta))
	urlSelenium := "http://host.docker.internal:" + strconv.Itoa(porta)
	comando := exec.Command(executavel)
	comando.Dir = raiz
	comando.Env = ambienteComSobrescritas(map[string]string{
		"ADMIN_PASSWORD":    "senha-e2e-segura",
		"ADMIN_SESSION_KEY": strings.Repeat("e", 64),
		"DATABASE_PATH":     filepath.Join(diretorioTemp, "e2e.db"),
		"SERVER_ADDRESS":    endereco,
	})
	comando.Stdout = os.Stdout
	comando.Stderr = os.Stderr
	if err := comando.Start(); err != nil {
		t.Fatalf("não foi possível iniciar o servidor E2E: %v", err)
	}
	t.Cleanup(func() {
		_ = comando.Process.Kill()
		_ = comando.Wait()
	})

	clienteHTTP := &http.Client{Timeout: time.Second}
	limite := time.After(30 * time.Second)
	tentativas := time.NewTicker(150 * time.Millisecond)
	defer tentativas.Stop()
	for {
		resposta, err := clienteHTTP.Get(urlLocal + "/admin/login")
		if err == nil {
			_ = resposta.Body.Close()
			if resposta.StatusCode == http.StatusOK {
				return urlSelenium
			}
		}
		select {
		case <-limite:
			t.Fatal("servidor temporário E2E não iniciou em 30 segundos")
		case <-tentativas.C:
		}
	}
}

func ambienteComSobrescritas(sobrescritas map[string]string) []string {
	ambiente := make([]string, 0, len(os.Environ())+len(sobrescritas))
	for _, variavel := range os.Environ() {
		nome, _, _ := strings.Cut(variavel, "=")
		if _, substituir := sobrescritas[nome]; !substituir {
			ambiente = append(ambiente, variavel)
		}
	}
	for nome, valor := range sobrescritas {
		ambiente = append(ambiente, nome+"="+valor)
	}
	return ambiente
}

func criarEEditarCliente(t *testing.T, driver selenium.WebDriver, baseURL string) {
	t.Helper()
	if err := driver.Get(baseURL + "/clientes"); err != nil {
		t.Fatal(err)
	}
	abrirModal(t, driver, "abrir-modal-cliente", "modal-cliente")
	preencher(t, driver, selenium.ByID, "cliente-nome", "Cliente Selenium")
	preencher(t, driver, selenium.ByID, "cliente-whatsapp", "11987654321")
	preencher(t, driver, selenium.ByID, "cliente-cidade", "São Paulo")
	preencher(t, driver, selenium.ByID, "cliente-estado", "SP")
	clicar(t, driver, selenium.ByCSSSelector, "#form-cliente button[type='submit']")
	assertTextNaPagina(t, driver, "Cliente Selenium")

	clicar(t, driver, selenium.ByCSSSelector, "[data-editar-cliente]")
	preencher(t, driver, selenium.ByID, "cliente-nome", "Cliente Selenium Atualizado")
	clicar(t, driver, selenium.ByCSSSelector, "#form-cliente button[type='submit']")
	assertTextNaPagina(t, driver, "Cliente Selenium Atualizado")
}

func criarEEditarProduto(t *testing.T, driver selenium.WebDriver, baseURL string) {
	t.Helper()
	if err := driver.Get(baseURL + "/produtos"); err != nil {
		t.Fatal(err)
	}
	abrirModal(t, driver, "abrir-modal-produto", "modal-produto")
	preencher(t, driver, selenium.ByID, "produto-nome", "Peça Selenium")
	preencher(t, driver, selenium.ByID, "produto-preco", "120.00")
	marcarCampo(t, driver, "pronta_entrega", true)
	marcarCampo(t, driver, "publicado", true)
	clicar(t, driver, selenium.ByCSSSelector, "#form-produto button[type='submit']")
	assertTextNaPagina(t, driver, "Peça Selenium")

	clicar(t, driver, selenium.ByCSSSelector, "[data-editar-produto]")
	preencher(t, driver, selenium.ByID, "produto-nome", "Peça Selenium Atualizada")
	preencher(t, driver, selenium.ByID, "produto-preco", "125.00")
	clicar(t, driver, selenium.ByCSSSelector, "#form-produto button[type='submit']")
	assertTextNaPagina(t, driver, "Peça Selenium Atualizada")
}

func verificarVitrinePublica(t *testing.T, driver selenium.WebDriver, baseURL string) {
	t.Helper()
	if err := driver.Get(baseURL + "/"); err != nil {
		t.Fatal(err)
	}
	assertTextNaPagina(t, driver, "Peça Selenium Atualizada")
}

func criarEAvancarPedido(t *testing.T, driver selenium.WebDriver, baseURL string) {
	t.Helper()
	if err := driver.Get(baseURL + "/pedidos"); err != nil {
		t.Fatal(err)
	}
	abrirModal(t, driver, "abrir-modal-pedido", "modal-pedido")
	preencher(t, driver, selenium.ByID, "cliente-busca", "Cliente Selenium Atualizado")
	clicar(t, driver, selenium.ByCSSSelector, "button[data-opcao-cliente]")
	preencher(t, driver, selenium.ByID, "produto-busca", "Peça Selenium Atualizada")
	clicar(t, driver, selenium.ByCSSSelector, "button[data-opcao-produto]")
	marcar(t, driver, "tem-desconto-cartao", true)
	preencher(t, driver, selenium.ByID, "percentual-desconto", "5")
	marcar(t, driver, "pix-parcelado", true)
	marcar(t, driver, "tem-desconto-pix-parcelado", true)
	preencher(t, driver, selenium.ByID, "percentual-desconto-pix-parcelado", "8")
	marcar(t, driver, "incluir-brinde", true)
	clicar(t, driver, selenium.ByCSSSelector, "form[action='/pedidos/inserir'] button[type='submit']")
	assertTextNaPagina(t, driver, "Cliente Selenium Atualizado")
	assertTextNaPagina(t, driver, "Peça Selenium Atualizada")

	orcamento := buscar(t, driver, selenium.ByCSSSelector, "textarea[id^='mensagem-orcamento-']")
	valorMensagem, err := driver.ExecuteScript("return arguments[0].value", []interface{}{orcamento})
	if err != nil {
		t.Fatal(err)
	}
	mensagem, ok := valorMensagem.(string)
	if !ok {
		t.Fatalf("valor do orçamento retornou tipo inesperado: %T", valorMensagem)
	}
	for _, trecho := range []string{"R$ 112,50", "R$ 118,75", "R$ 115,00", "R$ 34,50", "brinde"} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("orçamento E2E não contém %q", trecho)
		}
	}

	for _, status := range []string{"Aguardando sinal", "Sinal recebido", "Em produção", "Aguardando pagamento final", "Pago", "Entregue"} {
		clicar(t, driver, selenium.ByCSSSelector, "form[action='/pedidos/status'] button[type='submit']")
		esperarStatusPedido(t, driver, status)
	}
}

func registrarMovimentacoesFinanceiras(t *testing.T, driver selenium.WebDriver, baseURL string) {
	t.Helper()
	if err := driver.Get(baseURL + "/financeiro"); err != nil {
		t.Fatal(err)
	}
	assertTextNaPagina(t, driver, "Movimentações financeiras")
	registrarTransacao(t, driver, "Saída", "Despesa Selenium", "20.25")
	registrarTransacao(t, driver, "Entrada", "Receita Selenium", "75.50")
	assertTextNaPagina(t, driver, "Despesa Selenium")
	assertTextNaPagina(t, driver, "Receita Selenium")
	assertTextNaPagina(t, driver, "R$ 180,25")
}

func registrarTransacao(t *testing.T, driver selenium.WebDriver, tipo string, descricao string, valor string) {
	t.Helper()
	abrirModal(t, driver, "abrir-modal-transacao", "modal-transacao")
	seletorTipo := buscar(t, driver, selenium.ByID, "transacao-tipo")
	if tipo == "Entrada" {
		_, err := driver.ExecuteScript("arguments[0].value = arguments[1]; arguments[0].dispatchEvent(new Event('change', { bubbles: true }));", []interface{}{seletorTipo, tipo})
		if err != nil {
			t.Fatal(err)
		}
	}
	preencher(t, driver, selenium.ByID, "transacao-descricao", descricao)
	preencher(t, driver, selenium.ByID, "transacao-valor", valor)
	dataTransacao := time.Now().Format("2006-01-02")
	campoData := buscar(t, driver, selenium.ByID, "transacao-data")
	valorData, err := driver.ExecuteScript("arguments[0].value = arguments[1]; arguments[0].dispatchEvent(new Event('input', { bubbles: true })); arguments[0].dispatchEvent(new Event('change', { bubbles: true })); return arguments[0].value", []interface{}{campoData, dataTransacao})
	if err != nil {
		t.Fatal(err)
	}
	if valorData != dataTransacao {
		t.Fatalf("data no formulário = %v; esperava %s", valorData, dataTransacao)
	}
	clicar(t, driver, selenium.ByCSSSelector, "form[action='/financeiro/inserir'] button[type='submit']")
	assertTextNaPagina(t, driver, descricao)
	assertTextNaPagina(t, driver, dataTransacao)
}

func buscar(t *testing.T, driver selenium.WebDriver, por string, seletor string) selenium.WebElement {
	t.Helper()
	elemento, err := driver.FindElement(por, seletor)
	if err != nil {
		t.Fatalf("não encontrou elemento %q: %v", seletor, err)
	}
	return elemento
}

func clicar(t *testing.T, driver selenium.WebDriver, por string, seletor string) {
	t.Helper()
	if err := buscar(t, driver, por, seletor).Click(); err != nil {
		t.Fatalf("não foi possível clicar em %q: %v", seletor, err)
	}
	pausaEntreAcoes(t)
}

func abrirModal(t *testing.T, driver selenium.WebDriver, botaoID string, dialogoID string) {
	t.Helper()
	err := driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		if _, err := driver.FindElement(selenium.ByCSSSelector, "#"+dialogoID+"[open]"); err == nil {
			pausaEntreAcoes(t)
			return true, nil
		}
		botao, err := driver.FindElement(selenium.ByID, botaoID)
		if err != nil {
			return false, nil
		}
		_ = botao.Click()
		return false, nil
	}, 8*time.Second)
	if err != nil {
		t.Fatalf("modal %q não abriu após clicar em %q: %v", dialogoID, botaoID, err)
	}
}

func preencher(t *testing.T, driver selenium.WebDriver, por string, seletor string, valor string) {
	t.Helper()
	elemento := buscar(t, driver, por, seletor)
	if err := elemento.Clear(); err != nil {
		t.Fatalf("não foi possível limpar %q: %v", seletor, err)
	}
	if err := elemento.SendKeys(valor); err != nil {
		t.Fatalf("não foi possível preencher %q: %v", seletor, err)
	}
	pausaEntreAcoes(t)
}

func marcar(t *testing.T, driver selenium.WebDriver, id string, selecionado bool) {
	t.Helper()
	elemento := buscar(t, driver, selenium.ByID, id)
	marcarElemento(t, elemento, id, selecionado)
}

func marcarCampo(t *testing.T, driver selenium.WebDriver, nome string, selecionado bool) {
	t.Helper()
	elemento := buscar(t, driver, selenium.ByCSSSelector, "input[name='"+nome+"']")
	marcarElemento(t, elemento, nome, selecionado)
}

func marcarElemento(t *testing.T, elemento selenium.WebElement, nome string, selecionado bool) {
	t.Helper()
	atual, err := elemento.IsSelected()
	if err != nil {
		t.Fatalf("não foi possível ler estado de %q: %v", nome, err)
	}
	if atual != selecionado {
		if err := elemento.Click(); err != nil {
			t.Fatalf("não foi possível alterar %q: %v", nome, err)
		}
		pausaEntreAcoes(t)
	}
}

func pausaEntreAcoes(t *testing.T) {
	t.Helper()
	valor := os.Getenv("E2E_STEP_DELAY")
	if valor == "" {
		return
	}
	atraso, err := time.ParseDuration(valor)
	if err != nil {
		t.Fatalf("E2E_STEP_DELAY inválido %q: %v", valor, err)
	}
	if atraso > 0 {
		time.Sleep(atraso)
	}
}

func assertTextNaPagina(t *testing.T, driver selenium.WebDriver, esperado string) {
	t.Helper()
	texto := ""
	err := driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		body, err := driver.FindElement(selenium.ByTagName, "body")
		if err != nil {
			return false, nil
		}
		texto, err = body.Text()
		if err != nil {
			return false, nil
		}
		return strings.Contains(texto, esperado), nil
	}, 8*time.Second)
	if err != nil {
		urlAtual, _ := driver.CurrentURL()
		t.Fatalf("página %q não contém o texto visível %q; texto atual=%q", urlAtual, esperado, texto)
	}
}

func esperarStatusPedido(t *testing.T, driver selenium.WebDriver, esperado string) {
	t.Helper()
	statusAtual := ""
	err := driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		etiqueta, err := driver.FindElement(selenium.ByCSSSelector, "tbody tr td:nth-child(5) div span")
		if err != nil {
			return false, nil
		}
		statusAtual, err = etiqueta.Text()
		if err != nil {
			return false, nil
		}
		return statusAtual == esperado, nil
	}, 8*time.Second)
	if err != nil {
		t.Fatalf("status atual do pedido = %q; esperava %q", statusAtual, esperado)
	}
}

func verificarURLContem(t *testing.T, driver selenium.WebDriver, esperado string) {
	t.Helper()
	urlAtual := ""
	err := driver.WaitWithTimeout(func(driver selenium.WebDriver) (bool, error) {
		urlAtual, err := driver.CurrentURL()
		if err != nil {
			return false, nil
		}
		return strings.Contains(urlAtual, esperado), nil
	}, 8*time.Second)
	if err != nil {
		texto := ""
		if body, erroBody := driver.FindElement(selenium.ByTagName, "body"); erroBody == nil {
			texto, _ = body.Text()
		}
		t.Fatalf("URL atual = %q; esperava conter %q; texto da página=%q", urlAtual, esperado, texto)
	}
}
