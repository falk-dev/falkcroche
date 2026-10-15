const botaoNotificacoes = document.getElementById("abrir-notificacoes");
const modalNotificacoes = document.getElementById("modal-notificacoes");
const contadorNotificacoes = document.getElementById("contador-notificacoes");
const listaNotificacoes = document.getElementById("lista-notificacoes");
const mensagemNotificacoes = document.getElementById("mensagem-notificacoes");

function formatarDataNotificacao(data) {
    const [ano, mes, dia] = data.split("-").map(Number);
    return new Intl.DateTimeFormat("pt-BR").format(new Date(ano, mes - 1, dia));
}

function formatarValorNotificacao(valor) {
    return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" }).format(valor);
}

function criarNotificacao(notificacao) {
    const item = document.createElement("article");
    item.className = "rounded-xl border border-[#e8d5c4] p-4";

    const cabecalho = document.createElement("div");
    cabecalho.className = "flex items-start justify-between gap-3";
    const titulo = document.createElement("div");
    const cliente = document.createElement("p");
    cliente.className = "font-bold text-[#5a262c]";
    cliente.textContent = notificacao.cliente;
    const produto = document.createElement("p");
    produto.className = "mt-0.5 text-sm text-[#5a262c]/75";
    produto.textContent = notificacao.produto;
    titulo.append(cliente, produto);

    const valor = document.createElement("strong");
    valor.className = "whitespace-nowrap text-sm text-[#5a262c]";
    valor.textContent = formatarValorNotificacao(notificacao.valor);
    cabecalho.append(titulo, valor);

    const vencimento = document.createElement("p");
    vencimento.className = "mt-2 text-sm text-[#7d4438]";
    const hoje = new Date();
    const dataHoje = `${hoje.getFullYear()}-${String(hoje.getMonth() + 1).padStart(2, "0")}-${String(hoje.getDate()).padStart(2, "0")}`;
    const vencido = notificacao.data_vencimento < dataHoje;
    vencimento.textContent = `${notificacao.descricao} · ${vencido ? "Vencido em" : "Vence hoje"} ${formatarDataNotificacao(notificacao.data_vencimento)}`;

    const acoes = document.createElement("div");
    acoes.className = "mt-3 flex items-center justify-between gap-3";
    const verPedido = document.createElement("a");
    verPedido.href = `/pedidos#pedido-${notificacao.pedido_id}`;
    verPedido.className = "text-sm font-semibold text-[#5a262c] underline";
    verPedido.textContent = "Ver pedido";
    verPedido.addEventListener("click", () => modalNotificacoes.close());
    const marcarLida = document.createElement("button");
    marcarLida.type = "button";
    marcarLida.className = "rounded-lg bg-[#5a262c] px-3 py-2 text-xs font-bold text-[#ffeae8] hover:bg-[#472025]";
    marcarLida.textContent = "Marcar como lida";
    marcarLida.addEventListener("click", async () => {
        marcarLida.disabled = true;
        try {
            const response = await fetch("/notificacoes/ler", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: new URLSearchParams({ id: String(notificacao.id) }),
            });
            if (!response.ok) throw new Error(`Erro ${response.status} ao marcar lembrete como lido.`);
            item.remove();
            atualizarContador();
            if (listaNotificacoes.childElementCount === 0) {
                mensagemNotificacoes.textContent = "Você não tem lembretes de vencimento não lidos.";
            }
        } catch (error) {
            console.error(error);
            mensagemNotificacoes.textContent = "Não foi possível marcar o lembrete como lido. Tente novamente.";
            marcarLida.disabled = false;
        }
    });
    acoes.append(verPedido, marcarLida);
    item.append(cabecalho, vencimento, acoes);
    return item;
}

function atualizarContador() {
    const quantidade = listaNotificacoes.childElementCount;
    contadorNotificacoes.textContent = quantidade > 99 ? "99+" : String(quantidade);
    contadorNotificacoes.hidden = quantidade === 0;
}

async function carregarNotificacoes() {
    mensagemNotificacoes.textContent = "Carregando lembretes...";
    try {
        const response = await fetch("/notificacoes", { headers: { Accept: "application/json" } });
        if (!response.ok) throw new Error(`Erro ${response.status} ao carregar lembretes.`);
        const notificacoes = await response.json();
        listaNotificacoes.replaceChildren(...notificacoes.map(criarNotificacao));
        atualizarContador();
        mensagemNotificacoes.textContent = notificacoes.length
            ? `${notificacoes.length} lembrete(s) de vencimento não lido(s).`
            : "Você não tem lembretes de vencimento não lidos.";
    } catch (error) {
        console.error(error);
        mensagemNotificacoes.textContent = "Não foi possível carregar os lembretes. Tente novamente.";
    }
}

botaoNotificacoes.addEventListener("click", () => {
    modalNotificacoes.showModal();
    carregarNotificacoes();
});
document.getElementById("fechar-modal-notificacoes").addEventListener("click", () => modalNotificacoes.close());
modalNotificacoes.addEventListener("click", (evento) => {
    if (evento.target === modalNotificacoes) modalNotificacoes.close();
});
carregarNotificacoes();
window.setInterval(carregarNotificacoes, 60 * 1000);
