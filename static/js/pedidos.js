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
const checkboxEditarBrinde = document.getElementById("editar-pedido-brinde");
const containerEditarPedidoPrincipal = document.getElementById("editar-pedido-principal-container");
const selectEditarPedidoPrincipal = document.getElementById("editar-pedido-principal");
const opcoesFinanceirasEditar = document.getElementById("editar-opcoes-financeiras");
const avisoOpcoesBloqueadas = document.getElementById("editar-opcoes-bloqueadas");
const atualizarOpcoesEdicao = (temPagamentos) => {
    const brinde = checkboxEditarBrinde.checked;
    containerEditarPedidoPrincipal.hidden = !brinde;
    selectEditarPedidoPrincipal.disabled = !brinde;
    document.getElementById("editar-opcoes-orcamento").hidden = brinde;
    checkboxEditarBrinde.disabled = temPagamentos;
    opcoesFinanceirasEditar.querySelectorAll("input").forEach((campo) => {
        campo.disabled = temPagamentos;
    });
    avisoOpcoesBloqueadas.hidden = !temPagamentos;
};
document.querySelectorAll("[data-editar-pedido]").forEach((botao) => {
    botao.addEventListener("click", () => {
        botao.closest("details").open = false;
        const temPagamentos = Number(botao.dataset.pagamentos) > 0;
        document.getElementById("editar-pedido-id").value = botao.dataset.id;
        document.getElementById("editar-pedido-cliente").value = botao.dataset.clienteId;
        document.getElementById("editar-pedido-produto").value = botao.dataset.produtoId;
        document.getElementById("editar-pedido-data").value = botao.dataset.dataPedido;
        document.getElementById("editar-pedido-entrega").value = botao.dataset.dataEntrega;
        document.getElementById("editar-pedido-prazo").value = botao.dataset.diasPrazo;
        document.getElementById("editar-pedido-observacoes").value = botao.dataset.observacoes;
        checkboxEditarBrinde.checked = botao.dataset.pedidoBrinde === "true";
        selectEditarPedidoPrincipal.querySelectorAll("option[data-pedido-id]").forEach((opcao) => {
            opcao.hidden = opcao.dataset.pedidoId === botao.dataset.id;
        });
        selectEditarPedidoPrincipal.value = botao.dataset.pedidoPrincipalId === "0" ? "" : botao.dataset.pedidoPrincipalId;
        document.getElementById("editar-desconto-pix-avista").checked = botao.dataset.descontoPixAvista === "true";
        document.getElementById("editar-tem-desconto-cartao").checked = Number(botao.dataset.descontoCartao) > 0;
        document.getElementById("editar-percentual-desconto").value = botao.dataset.descontoCartao;
        document.getElementById("editar-pix-parcelado").checked = botao.dataset.pixParcelado === "true";
        document.getElementById("editar-tem-desconto-pix-parcelado").checked = Number(botao.dataset.descontoPix) > 0;
        document.getElementById("editar-percentual-desconto-pix-parcelado").value = botao.dataset.descontoPix;
        document.getElementById("editar-incluir-brinde").checked = botao.dataset.incluirBrinde === "true";
        document.getElementById("editar-linha-cores-nome").value = botao.dataset.linhaCoresNome;
        document.getElementById("editar-linha-cores-link").value = botao.dataset.linhaCoresLink;
        atualizarOpcoesEdicao(temPagamentos);
        modalEditarPedido.showModal();
    });
});
checkboxEditarBrinde.addEventListener("change", () => atualizarOpcoesEdicao(false));
document.getElementById("fechar-modal-editar-pedido").addEventListener("click", () => modalEditarPedido.close());
document.getElementById("cancelar-modal-editar-pedido").addEventListener("click", () => modalEditarPedido.close());
document.getElementById("form-editar-pedido").addEventListener("submit", (evento) => {
    if (!window.confirm("Salvar as alterações do pedido e atualizar o orçamento conforme as opções escolhidas?")) {
        evento.preventDefault();
    }
});

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

document.querySelectorAll("[data-abrir-detalhes]").forEach((botao) => {
    const modal = document.getElementById(botao.dataset.modalId);
    botao.addEventListener("click", () => {
        botao.closest("details").open = false;
        modal.showModal();
    });
    modal.querySelector("[data-fechar-detalhes]").addEventListener("click", () => modal.close());
    modal.addEventListener("click", (evento) => {
        if (evento.target === modal) modal.close();
    });
});

const pedidoDetalhesId = new URLSearchParams(window.location.search).get("detalhes");
if (pedidoDetalhesId) {
    document.getElementById("modal-detalhes-" + pedidoDetalhesId)?.showModal();
}

document.querySelectorAll("[data-copiar-orcamento]").forEach((botao) => {
    botao.addEventListener("click", async () => {
        botao.closest("details").open = false;
        const pedidoId = botao.dataset.pedidoId;
        const mensagem = document.getElementById("mensagem-orcamento-" + pedidoId);
        const notificacao = document.getElementById("notificacao-copia-orcamento");
        const copiarComFallback = () => {
            const temporaria = document.createElement("textarea");
            temporaria.value = mensagem.value;
            temporaria.setAttribute("readonly", "");
            temporaria.style.position = "fixed";
            temporaria.style.opacity = "0";
            document.body.appendChild(temporaria);
            temporaria.select();
            const copiado = document.execCommand("copy");
            temporaria.remove();
            if (!copiado) throw new Error("Não foi possível copiar o orçamento.");
        };
        try {
            if (navigator.clipboard?.writeText) {
                try {
                    await navigator.clipboard.writeText(mensagem.value);
                } catch {
                    copiarComFallback();
                }
            } else {
                copiarComFallback();
            }
            notificacao.hidden = false;
            clearTimeout(notificacao.timeoutId);
            notificacao.timeoutId = setTimeout(() => {
                notificacao.hidden = true;
            }, 3500);
            document.getElementById("status-copia-orcamento-" + pedidoId).textContent = "Orçamento copiado.";
        } catch (erro) {
            notificacao.textContent = "Não foi possível copiar o orçamento.";
            notificacao.classList.replace("bg-[#5a262c]", "bg-[#9b3b36]");
            notificacao.hidden = false;
            clearTimeout(notificacao.timeoutId);
            notificacao.timeoutId = setTimeout(() => {
                notificacao.hidden = true;
                notificacao.textContent = "Orçamento copiado.";
                notificacao.classList.replace("bg-[#9b3b36]", "bg-[#5a262c]");
            }, 3500);
            console.error(erro);
        }
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