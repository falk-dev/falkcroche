package controladores

import (
	"strings"
	"testing"
	"time"

	"falkcroche/modelos"
	"falkcroche/views"
)

func TestCalcularDataEntregaContaSomenteDiasUteis(t *testing.T) {
	inicio := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.Local)
	resultado := calcularDataEntrega(inicio, 1)
	esperado := time.Date(2026, time.October, 12, 12, 0, 0, 0, time.Local)
	if !resultado.Equal(esperado) {
		t.Fatalf("data calculada = %s; esperado %s", resultado.Format("2006-01-02"), esperado.Format("2006-01-02"))
	}
}

func TestPrazoPadraoDeVinteDiasUteis(t *testing.T) {
	inicio := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.Local)
	resultado := calcularDataEntrega(inicio, prazoPadraoDiasUteis)
	esperado := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	if !resultado.Equal(esperado) {
		t.Fatalf("data calculada = %s; esperado %s", resultado.Format("02-01-2006"), esperado.Format("02-01-2006"))
	}
}

func TestMensagemOrcamentoComDescontoEParcelas(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "Linha Exemplo", "https://exemplo.com/cores", false, 0, false, true, false, 0, true)
	for _, trecho := range []string{
		"oi, Ana! seguem os detalhes do orçamento",
		"🌸 *peça:* Blusa Vitória-Régia",
		"💰 *valor:* R$ 170,00",
		"*pix à vista*\nR$ 153,00\n10% de desconto",
		"*cartão de crédito*\nR$ 170,00\n→ sinal de 30% via pix: R$ 51,00\n→ restante: R$ 119,00 em até 3x sem juros.",
		"🎁 *brinde:* a compra inclui um pequeno brinde nas cores escolhidas para a sua peça.",
		"prazo de produção:* 20 dias úteis, podendo ser entregue antes.",
		"previsão de entrega:* 06/11/2026",
		"este orçamento é válido por 15 dias",
		"Linha Exemplo:\nhttps://exemplo.com/cores",
	} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("mensagem de orçamento não contém %q", trecho)
		}
	}
}

func TestMensagemOrcamentoAplicaDescontoPromocionalSomenteNoCartao(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", true, 5, false, false, false, 0, true)
	for _, trecho := range []string{
		"*pix à vista*\nR$ 153,00\n10% de desconto",
		"*cartão de crédito*\nR$ 161,50\n5% de desconto",
		"→ sinal de 30% via pix: R$ 48,45",
		"→ restante: R$ 113,05 em até 3x sem juros.",
	} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("mensagem de orçamento não contém %q", trecho)
		}
	}
}

func TestMensagemIncluiPixParceladoSomenteQuandoSelecionado(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	semOpcao := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", false, 0, false, false, false, 0, true)
	comOpcao := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", false, 0, true, true, false, 0, true)
	if strings.Contains(semOpcao, "pix parcelado") || strings.Contains(semOpcao, "brinde") {
		t.Fatalf("mensagem incluiu opcionais não selecionados: %s", semOpcao)
	}
	for _, trecho := range []string{
		"*pix parcelado*\nR$ 170,00\n→ sinal de 30% via pix: R$ 51,00",
		"→ restante: R$ 119,00 em até 3 parcelas via pix.",
		"🎁 *brinde:* a compra inclui um pequeno brinde nas cores escolhidas para a sua peça.",
	} {
		if !strings.Contains(comOpcao, trecho) {
			t.Errorf("mensagem de orçamento não contém %q", trecho)
		}
	}
}

func TestMensagemAplicaDescontoAoPixParceladoERecalculaSinal(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", false, 0, true, false, true, 10, true)
	for _, trecho := range []string{
		"*pix parcelado*\nR$ 153,00\n10% de desconto",
		"→ sinal de 30% via pix: R$ 45,90",
		"→ restante: R$ 107,10 em até 3 parcelas via pix.",
	} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("mensagem de orçamento não contém %q", trecho)
		}
	}
}

func TestMensagemOrcamentoPermiteDesativarDescontoNoPixAvista(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", false, 0, false, false, false, 0, false)
	if !strings.Contains(mensagem, "*pix à vista*\nR$ 170,00") || strings.Contains(mensagem, "10% de desconto") {
		t.Fatalf("mensagem Pix à vista não respeitou a opção sem desconto: %s", mensagem)
	}
}

func TestMensagemSemLinhaDeCoresOmiteSecaoOpcional(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "", false, 0, false, false, false, 0, true)
	if strings.Contains(mensagem, "cores disponíveis") || strings.Contains(mensagem, "[nome da linha]") {
		t.Fatalf("mensagem sem linha de cores incluiu conteúdo placeholder: %s", mensagem)
	}
}

func TestProximosStatusPedidoSegueFluxo(t *testing.T) {
	proximos := modelos.ProximosStatusPedido(modelos.StatusSinalRecebido)
	if len(proximos) != 2 || proximos[0] != modelos.StatusEmProducao || proximos[1] != modelos.StatusCancelado {
		t.Fatalf("próximos status inesperados: %v", proximos)
	}
	proximosBrinde := modelos.ProximosStatusPedidoComBrinde(modelos.StatusEmProducao, true)
	if len(proximosBrinde) != 2 || proximosBrinde[0] != modelos.StatusEntregue || proximosBrinde[1] != modelos.StatusCancelado {
		t.Fatalf("próximos status de brinde inesperados: %v", proximosBrinde)
	}
}

func TestMensagemDeBrindeIndicaValorZeroEPedidoRelacionado(t *testing.T) {
	mensagem := montarMensagemBrinde("Ana", "Mini flor", 12, time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC))
	for _, trecho := range []string{"brinde *Mini flor*", "pedido #12", "R$ 0,00", "06/11/2026"} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("mensagem de brinde não contém %q: %s", trecho, mensagem)
		}
	}
}

func TestTodosPagamentosRecebidosExigeTodosOsItensConfirmados(t *testing.T) {
	if todosPagamentosRecebidos(nil) {
		t.Fatal("nenhum pagamento não deve ser considerado quitação")
	}
	if todosPagamentosRecebidos([]modelos.Pagamento{{Status: "Recebido"}, {Status: "Pendente"}}) {
		t.Fatal("parcelas pendentes não devem ser consideradas quitadas")
	}
	if !todosPagamentosRecebidos([]modelos.Pagamento{{Status: "Recebido"}, {Status: "Recebido"}}) {
		t.Fatal("todos os pagamentos recebidos deveriam indicar quitação")
	}
}

func TestMensagemLembreteIdentificaPecaPeloNome(t *testing.T) {
	mensagem := montarMensagemLembrete("Camila", "Amigurumi Coelho", modelos.Pagamento{
		Descricao:      "Parcela 1 de 2",
		Valor:          29.90,
		DataVencimento: "2026-11-15",
	})
	if !strings.Contains(mensagem, "da peça *Amigurumi Coelho*") {
		t.Fatalf("lembrete não identifica a peça: %s", mensagem)
	}
	if strings.Contains(mensagem, "pedido #") {
		t.Fatalf("lembrete ainda menciona número do pedido: %s", mensagem)
	}
}

func TestFormatarReaisUsaPadraoBrasileiro(t *testing.T) {
	testes := []struct {
		valor       float64
		expectativa string
	}{
		{valor: 1100, expectativa: "R$ 1.100,00"},
		{valor: 1234.5, expectativa: "R$ 1.234,50"},
		{valor: -12.5, expectativa: "-R$ 12,50"},
	}
	for _, teste := range testes {
		if resultado := views.FormatarReais(teste.valor); resultado != teste.expectativa {
			t.Errorf("FormatarReais(%v) = %q; esperado %q", teste.valor, resultado, teste.expectativa)
		}
	}
}
