const modalCliente = document.getElementById("modal-cliente");
const formCliente = document.getElementById("form-cliente");
const tituloModalCliente = document.getElementById("titulo-modal-cliente");
const botaoSalvarCliente = formCliente.querySelector("button[type='submit']");

function abrirModalCliente(cliente) {
    formCliente.reset();
    formCliente.action = cliente ? "/clientes/atualizar" : "/clientes/inserir";
    tituloModalCliente.textContent = cliente ? "Editar cliente" : "Novo cliente";
    botaoSalvarCliente.textContent = cliente ? "Salvar alterações" : "Salvar cliente";
    if (cliente) {
        document.getElementById("cliente-id").value = cliente.id;
        document.getElementById("cliente-nome").value = cliente.nome;
        document.getElementById("cliente-whatsapp").value = cliente.whatsapp;
        document.getElementById("cliente-cidade").value = cliente.cidade;
        document.getElementById("cliente-estado").value = cliente.estado;
    }
    modalCliente.showModal();
}

document.getElementById("abrir-modal-cliente").addEventListener("click", () => abrirModalCliente(null));
document.querySelectorAll("[data-editar-cliente]").forEach((botao) => {
    botao.addEventListener("click", () => abrirModalCliente(botao.dataset));
});
document.getElementById("fechar-modal-cliente").addEventListener("click", () => modalCliente.close());
document.getElementById("cancelar-modal-cliente").addEventListener("click", () => modalCliente.close());