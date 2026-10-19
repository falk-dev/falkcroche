const modalTransacao = document.getElementById("modal-transacao");
const formularioTransacao = document.getElementById("form-transacao");
const tituloTransacao = document.getElementById("transacao-modal-titulo");
const idTransacao = document.getElementById("transacao-id");
const tipoTransacao = document.getElementById("transacao-tipo");
const tipoContainerTransacao = document.getElementById("transacao-tipo-container");
const tipoVinculadoTransacao = document.getElementById("transacao-tipo-vinculado");
const descricaoTransacao = document.getElementById("transacao-descricao");
const valorTransacao = document.getElementById("transacao-valor");
const dataTransacao = document.getElementById("transacao-data");
const botaoTransacao = document.getElementById("transacao-submit");

document.getElementById("abrir-modal-transacao").addEventListener("click", () => {
    formularioTransacao.action = "/financeiro/inserir";
    idTransacao.disabled = true;
    idTransacao.value = "";
    tipoTransacao.disabled = false;
    tipoTransacao.value = "Saída";
    tipoContainerTransacao.hidden = false;
    tipoVinculadoTransacao.hidden = true;
    descricaoTransacao.value = "";
    valorTransacao.value = "";
    dataTransacao.value = "";
    tituloTransacao.textContent = "Novo lançamento";
    botaoTransacao.textContent = "Registrar";
    modalTransacao.showModal();
});

document.querySelectorAll("[data-editar-transacao]").forEach((botao) => {
    botao.addEventListener("click", () => {
        formularioTransacao.action = "/financeiro/atualizar";
        idTransacao.disabled = false;
        idTransacao.value = botao.dataset.id;
        tipoTransacao.value = botao.dataset.tipo;
        const vinculada = botao.dataset.vinculada === "true";
        tipoTransacao.disabled = vinculada;
        tipoContainerTransacao.hidden = vinculada;
        tipoVinculadoTransacao.hidden = !vinculada;
        descricaoTransacao.value = botao.dataset.descricao;
        valorTransacao.value = botao.dataset.valor;
        dataTransacao.value = botao.dataset.data;
        tituloTransacao.textContent = "Editar lançamento";
        botaoTransacao.textContent = "Salvar alterações";
        modalTransacao.showModal();
    });
});

document.getElementById("fechar-modal-transacao").addEventListener("click", () => modalTransacao.close());
document.getElementById("cancelar-modal-transacao").addEventListener("click", () => modalTransacao.close());