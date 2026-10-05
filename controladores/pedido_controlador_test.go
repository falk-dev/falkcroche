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
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "Linha Exemplo", "https://exemplo.com/cores")
	for _, trecho := range []string{
		"oi, Ana! seguem os detalhes do orçamento",
		"🌸 peça: Blusa Vitória-Régia",
		"💰 valor: R$ 170,00",
		"pix à vista: R$ 153,00 (10% de desconto)",
		"sinal de 30% via pix (R$ 51,00)",
		"o restante (R$ 119,00) pode ser parcelado em até 3x sem juros",
		"2x de R$ 39,67 e 1x de R$ 39,66",
		"prazo de produção: 20 dias úteis",
		"previsão de entrega: 06-11-2026",
		"este orçamento é válido por 15 dias",
		"Linha Exemplo:\nhttps://exemplo.com/cores",
	} {
		if !strings.Contains(mensagem, trecho) {
			t.Errorf("mensagem de orçamento não contém %q", trecho)
		}
	}
}

func TestMensagemSemLinhaDeCoresOmiteSecaoOpcional(t *testing.T) {
	entrega := time.Date(2026, time.November, 6, 12, 0, 0, 0, time.Local)
	mensagem := montarMensagemOrcamento("Ana", "Blusa Vitória-Régia", 170, prazoPadraoDiasUteis, entrega, "", "")
	if strings.Contains(mensagem, "cores disponíveis") || strings.Contains(mensagem, "[nome da linha]") {
		t.Fatalf("mensagem sem linha de cores incluiu conteúdo placeholder: %s", mensagem)
	}
}

func TestProximosStatusPedidoSegueFluxo(t *testing.T) {
	proximos := modelos.ProximosStatusPedido(modelos.StatusSinalRecebido)
	if len(proximos) != 2 || proximos[0] != modelos.StatusEmProducao || proximos[1] != modelos.StatusCancelado {
		t.Fatalf("próximos status inesperados: %v", proximos)
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
