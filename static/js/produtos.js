const modalProduto = document.getElementById("modal-produto");
const formProduto = document.getElementById("form-produto");
const tituloModalProduto = document.getElementById("titulo-modal-produto");
const botaoSalvarProduto = formProduto.querySelector("button[type='submit']");
const campoCategoriaProduto = document.getElementById("produto-categoria");
const prioridadeCategoriaProduto = document.getElementById("produto-categoria-prioridade");
const normalizarCategoria = (categoria) => categoria.trim().toLocaleLowerCase("pt-BR");
const prioridadesCategorias = new Map(Array.from(document.querySelectorAll("#categorias-produto option"), (opcao) => [
    normalizarCategoria(opcao.value),
    opcao.dataset.prioridade,
]));

campoCategoriaProduto.addEventListener("change", () => {
    prioridadeCategoriaProduto.value = prioridadesCategorias.get(normalizarCategoria(campoCategoriaProduto.value)) || "media";
});

function abrirModalProduto(produto) {
    formProduto.reset();
    formProduto.action = produto ? "/produtos/atualizar" : "/produtos/inserir";
    tituloModalProduto.textContent = produto ? "Editar produto" : "Novo produto";
    botaoSalvarProduto.textContent = produto ? "Salvar alterações" : "Salvar produto";
    if (produto) {
        document.getElementById("produto-id").value = produto.id;
        document.getElementById("produto-nome").value = produto.nome;
        document.getElementById("produto-categoria").value = produto.categoria;
        prioridadeCategoriaProduto.value = produto.prioridadeCategoria || prioridadesCategorias.get(normalizarCategoria(produto.categoria)) || "media";
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