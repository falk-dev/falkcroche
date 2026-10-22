const filtrosVitrine = Array.from(document.querySelectorAll("[data-filtro-vitrine]"));
const produtosVitrine = Array.from(document.querySelectorAll("[data-produto-vitrine]"));

filtrosVitrine.forEach((filtro) => {
    filtro.addEventListener("click", () => {
        const categoria = filtro.dataset.filtroVitrine;
        filtrosVitrine.forEach((outroFiltro) => {
            outroFiltro.setAttribute("aria-pressed", String(outroFiltro === filtro));
        });
        produtosVitrine.forEach((produto) => {
            produto.hidden = categoria !== "*" && produto.dataset.categoria !== categoria;
        });
    });
});