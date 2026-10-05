const modalTransacao = document.getElementById("modal-transacao");
document.getElementById("abrir-modal-transacao").addEventListener("click", () => modalTransacao.showModal());
document.getElementById("fechar-modal-transacao").addEventListener("click", () => modalTransacao.close());
document.getElementById("cancelar-modal-transacao").addEventListener("click", () => modalTransacao.close());