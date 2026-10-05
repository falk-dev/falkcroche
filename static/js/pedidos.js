const modalPedido = document.getElementById("modal-pedido");
document.getElementById("abrir-modal-pedido").addEventListener("click", () => modalPedido.showModal());
document.getElementById("fechar-modal-pedido").addEventListener("click", () => modalPedido.close());
document.getElementById("cancelar-modal-pedido").addEventListener("click", () => modalPedido.close());

function configurarBusca(tipo) {
    const campo = document.getElementById(tipo + "-campo");
    const busca = document.getElementById(tipo + "-busca");
    const idSelecionado = document.getElementById(tipo + "-id");
    const opcoes = document.getElementById(tipo + "-opcoes");
    const semResultado = document.getElementById(tipo + "-sem-resultado");
    const botoes = Array.from(opcoes.querySelectorAll("[data-opcao-" + tipo + "]"));

    botoes.sort((a, b) => a.dataset.label.localeCompare(b.dataset.label, "pt-BR", { sensitivity: "base" }));
    botoes.forEach((botao) => opcoes.insertBefore(botao, semResultado));

    const normalizar = (texto) => texto
        .normalize("NFD")
        .replace(/[\u0300-\u036f]/g, "")
        .toLocaleLowerCase("pt-BR");

    busca.addEventListener("input", () => {
        idSelecionado.value = "";
        const termo = normalizar(busca.value.trim());
        let encontrou = false;
        botoes.forEach((botao) => {
            const corresponde = normalizar(botao.dataset.busca).includes(termo);
            botao.style.display = corresponde ? "" : "none";
            encontrou = encontrou || corresponde;
        });
        semResultado.classList.toggle("hidden", encontrou);
        opcoes.classList.remove("hidden");
    });
    busca.addEventListener("focus", () => opcoes.classList.remove("hidden"));
    botoes.forEach((botao) => botao.addEventListener("click", () => {
        idSelecionado.value = botao.dataset.id;
        busca.value = botao.dataset.label;
        opcoes.classList.add("hidden");
    }));
    document.addEventListener("click", (evento) => {
        if (!campo.contains(evento.target)) opcoes.classList.add("hidden");
    });
}

configurarBusca("cliente");
configurarBusca("produto");
document.querySelectorAll("[data-copiar-orcamento]").forEach((botao) => {
    botao.addEventListener("click", async () => {
        const pedidoId = botao.dataset.pedidoId;
        const mensagem = document.getElementById("mensagem-orcamento-" + pedidoId);
        try {
            await navigator.clipboard.writeText(mensagem.value);
        } catch {
            const temporaria = document.createElement("textarea");
            temporaria.value = mensagem.value;
            document.body.appendChild(temporaria);
            temporaria.select();
            document.execCommand("copy");
            temporaria.remove();
        }
        botao.textContent = "✅";
        botao.title = "Orçamento copiado";
        botao.setAttribute("aria-label", "Orçamento copiado");
        document.getElementById("status-copia-orcamento-" + pedidoId).textContent = "Orçamento copiado.";
    });
});