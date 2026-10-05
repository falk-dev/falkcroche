const modalCliente = document.getElementById("modal-cliente");
const formCliente = document.getElementById("form-cliente");
const tituloModalCliente = document.getElementById("titulo-modal-cliente");
const botaoSalvarCliente = formCliente.querySelector("button[type='submit']");
const campoWhatsappCliente = document.getElementById("cliente-whatsapp");

function formatarTelefone(valor) {
    let digitos = valor.replace(/\D/g, "");
    let codigoPais = "";

    if (digitos.startsWith("55") && digitos.length > 11) {
        codigoPais = "+55 ";
        digitos = digitos.slice(2, 13);
    } else {
        digitos = digitos.slice(0, 11);
    }

    if (digitos.length < 2) return codigoPais + (digitos ? `(${digitos}` : "");

    const ddd = digitos.slice(0, 2);
    const numero = digitos.slice(2);
    const tamanhoPrefixo = digitos.length > 10 ? 5 : 4;
    const prefixo = `${codigoPais}(${ddd})`;

    if (!numero) return prefixo;
    if (numero.length <= tamanhoPrefixo) return `${prefixo} ${numero}`;
    return `${prefixo} ${numero.slice(0, tamanhoPrefixo)}-${numero.slice(tamanhoPrefixo)}`;
}

campoWhatsappCliente.addEventListener("input", () => {
    const cursor = campoWhatsappCliente.selectionStart;
    const digitosAntesDoCursor = campoWhatsappCliente.value.slice(0, cursor).replace(/\D/g, "").length;
    campoWhatsappCliente.value = formatarTelefone(campoWhatsappCliente.value);

    let digitosEncontrados = 0;
    let novaPosicao = digitosAntesDoCursor === 0 ? 0 : campoWhatsappCliente.value.length;
    for (let indice = 0; indice < campoWhatsappCliente.value.length; indice++) {
        if (/\d/.test(campoWhatsappCliente.value[indice])) digitosEncontrados++;
        if (digitosEncontrados === digitosAntesDoCursor) {
            novaPosicao = indice + 1;
            break;
        }
    }
    campoWhatsappCliente.setSelectionRange(novaPosicao, novaPosicao);
});

function abrirModalCliente(cliente) {
    formCliente.reset();
    formCliente.action = cliente ? "/clientes/atualizar" : "/clientes/inserir";
    tituloModalCliente.textContent = cliente ? "Editar cliente" : "Novo cliente";
    botaoSalvarCliente.textContent = cliente ? "Salvar alterações" : "Salvar cliente";
    if (cliente) {
        document.getElementById("cliente-id").value = cliente.id;
        document.getElementById("cliente-nome").value = cliente.nome;
        campoWhatsappCliente.value = formatarTelefone(cliente.whatsapp);
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