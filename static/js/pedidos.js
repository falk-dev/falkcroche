const modalPedido = document.getElementById("modal-pedido");
const checkboxPedidoBrinde = document.querySelector("[data-pedido-brinde]");
const pedidoPrincipalContainer = document.getElementById("pedido-principal-container");
const pedidoPrincipal = document.getElementById("pedido-principal");
const opcoesOrcamentoPedido = document.getElementById("opcoes-orcamento-pedido");
const avisoPedidoBrinde = document.querySelector("[data-aviso-pedido-brinde]");
const criarPedidoSubmit = document.getElementById("criar-pedido-submit");
const atualizarTipoPedido = () => {
    const brinde = checkboxPedidoBrinde.checked;
    pedidoPrincipalContainer.hidden = !brinde;
    pedidoPrincipal.disabled = !brinde;
    opcoesOrcamentoPedido.hidden = brinde;
    opcoesOrcamentoPedido.querySelectorAll("input, select").forEach((campo) => {
        campo.disabled = brinde;
    });
    avisoPedidoBrinde.hidden = !brinde;
    criarPedidoSubmit.textContent = brinde ? "Criar pedido de brinde" : "Criar orçamento";
};
checkboxPedidoBrinde.addEventListener("change", atualizarTipoPedido);
document.getElementById("abrir-modal-pedido").addEventListener("click", () => {
    checkboxPedidoBrinde.checked = false;
    pedidoPrincipal.value = "";
    atualizarTipoPedido();
    modalPedido.showModal();
});
document.getElementById("fechar-modal-pedido").addEventListener("click", () => modalPedido.close());
document.getElementById("cancelar-modal-pedido").addEventListener("click", () => modalPedido.close());

const modalEditarPedido = document.getElementById("modal-editar-pedido");
document.querySelectorAll("[data-editar-pedido]").forEach((botao) => {
    botao.addEventListener("click", () => {
        document.getElementById("editar-pedido-id").value = botao.dataset.id;
        document.getElementById("editar-pedido-data").value = botao.dataset.dataPedido;
        document.getElementById("editar-pedido-entrega").value = botao.dataset.dataEntrega;
        document.getElementById("editar-pedido-observacoes").value = botao.dataset.observacoes;
        modalEditarPedido.showModal();
    });
});
document.getElementById("fechar-modal-editar-pedido").addEventListener("click", () => modalEditarPedido.close());
document.getElementById("cancelar-modal-editar-pedido").addEventListener("click", () => modalEditarPedido.close());

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
document.querySelectorAll("[data-forma-pagamento]").forEach((seletor) => {
    const detalhes = seletor.closest("form").querySelector("[data-detalhes-parcelamento]");
    const atualizarDetalhes = () => {
        const parcelado = seletor.value !== "Pix à vista";
        detalhes.hidden = !parcelado;
        detalhes.querySelectorAll("select, input").forEach((campo) => {
            campo.disabled = !parcelado;
            campo.required = parcelado && campo.type === "date";
        });
    };
    seletor.addEventListener("change", atualizarDetalhes);
    atualizarDetalhes();
});

document.querySelectorAll("form[action='/pedidos/status']").forEach((formulario) => {
    const seletorStatus = formulario.querySelector("[data-pedido-status]");
    const pedidoId = formulario.querySelector("input[name='pedido_id']").value;
    const modal = document.getElementById("modal-sinal-" + pedidoId);
    if (!modal) return;

    formulario.addEventListener("submit", (evento) => {
        if (seletorStatus.value !== "Sinal recebido") return;
        evento.preventDefault();
        modal.showModal();
    });
    modal.querySelectorAll("[data-fechar-modal-sinal]").forEach((botao) => {
        botao.addEventListener("click", () => modal.close());
    });
    modal.addEventListener("click", (evento) => {
        if (evento.target === modal) modal.close();
    });
});

document.querySelectorAll("[data-abrir-pagamentos]").forEach((botao) => {
    const modal = document.getElementById(botao.dataset.modalId);
    botao.addEventListener("click", () => modal.showModal());
    modal.querySelector("[data-fechar-pagamentos]").addEventListener("click", () => modal.close());
    modal.addEventListener("click", (evento) => {
        if (evento.target === modal) modal.close();
    });
});

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

document.querySelectorAll("[data-copiar-lembrete]").forEach((botao) => {
    botao.addEventListener("click", async () => {
        const pagamentoId = botao.dataset.pagamentoId;
        const mensagem = document.getElementById("mensagem-lembrete-" + pagamentoId);
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
        botao.textContent = "Lembrete copiado";
        document.getElementById("status-copia-lembrete-" + pagamentoId).textContent = "Lembrete copiado.";
    });
});