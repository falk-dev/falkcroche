const modalPedido = document.getElementById("modal-pedido");
document.getElementById("abrir-modal-pedido").addEventListener("click", () => modalPedido.showModal());
document.getElementById("fechar-modal-pedido").addEventListener("click", () => modalPedido.close());
document.getElementById("cancelar-modal-pedido").addEventListener("click", () => modalPedido.close());

function configurarBusca(tipo) {
    const campo = document.getElementById(tipo + "-campo");
    const busca = document.getElementById(tipo + "-busca");
    const idSelecionado = document.getElementById(tipo + "-id");
    const opcoes = document.getElementById(tipo + "-opcoes");
    const botoes = opcoes.querySelectorAll("[data-opcao-" + tipo + "]");
    const semResultado = document.getElementById(tipo + "-sem-resultado");

    busca.addEventListener("input", () => {
        idSelecionado.value = "";
        const termo = busca.value.trim().toLocaleLowerCase("pt-BR");
        let encontrou = false;
        botoes.forEach((botao) => {
            const corresponde = botao.dataset.busca.toLocaleLowerCase("pt-BR").includes(termo);
            botao.hidden = !corresponde;
            encontrou = encontrou || corresponde;
        });
        semResultado.hidden = encontrou;
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