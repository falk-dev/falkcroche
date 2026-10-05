const modalProduto = document.getElementById("modal-produto");
const formProduto = document.getElementById("form-produto");
const tituloModalProduto = document.getElementById("titulo-modal-produto");
const botaoSalvarProduto = formProduto.querySelector("button[type='submit']");

function abrirModalProduto(produto) {
    formProduto.reset();
    formProduto.action = produto ? "/produtos/atualizar" : "/produtos/inserir";
    tituloModalProduto.textContent = produto ? "Editar produto" : "Novo produto";
    botaoSalvarProduto.textContent = produto ? "Salvar alterações" : "Salvar produto";
    if (produto) {
        document.getElementById("produto-id").value = produto.id;
        document.getElementById("produto-nome").value = produto.nome;
        document.getElementById("produto-preco").value = produto.preco;
        document.getElementById("produto-foto").value = produto.foto;
        formProduto.querySelector("[name='pronta_entrega']").checked = produto.prontaEntrega === "true";
        formProduto.querySelector("[name='publicado']").checked = produto.publicado === "true";
    }
    modalProduto.showModal();
}

document.getElementById("abrir-modal-produto").addEventListener("click", () => abrirModalProduto(null));
document.querySelectorAll("[data-editar-produto]").forEach((botao) => {
    botao.addEventListener("click", () => abrirModalProduto(botao.dataset));
});
document.getElementById("fechar-modal-produto").addEventListener("click", () => modalProduto.close());
document.getElementById("cancelar-modal-produto").addEventListener("click", () => modalProduto.close());