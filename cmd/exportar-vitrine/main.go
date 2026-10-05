package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"falkcroche/daos"
	"falkcroche/database"
	"falkcroche/views"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "Aviso: não foi possível carregar .env: %v\n", err)
	}
	if err := exportarVitrine(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exportarVitrine() error {
	diretorio := flag.String("out", "docs", "pasta de saída dos arquivos estáticos")
	flag.Parse()

	if _, err := os.Stat("falkcroche.db"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("banco falkcroche.db não encontrado; execute o comando na raiz do projeto")
		}
		return fmt.Errorf("erro ao acessar o banco de dados: %w", err)
	}
	numeroWhatsApp := strings.TrimSpace(os.Getenv("WHATSAPP_NUMERO"))
	if numeroWhatsApp == "" {
		return errors.New("configure WHATSAPP_NUMERO antes de exportar a vitrine")
	}

	db := database.Conectar()
	defer db.Close()

	produtos, err := daos.NovoProdutoDAO(db).BuscarPublicados()
	if err != nil {
		return fmt.Errorf("erro ao buscar produtos publicados: %w", err)
	}

	diretorioStatic := filepath.Join(*diretorio, "static")
	if err := os.MkdirAll(diretorioStatic, 0755); err != nil {
		return fmt.Errorf("erro ao criar pasta de saída: %w", err)
	}
	logo, err := os.ReadFile("logo.png")
	if err != nil {
		return fmt.Errorf("erro ao ler logo.png: %w", err)
	}
	if err := os.WriteFile(filepath.Join(diretorioStatic, "logo.png"), logo, 0644); err != nil {
		return fmt.Errorf("erro ao copiar logo para a exportação: %w", err)
	}

	pagina := views.PublicLayout("Vitrine", views.VitrinePublica(produtos, numeroWhatsApp), "static/logo.png")
	arquivo, err := os.Create(filepath.Join(*diretorio, "index.html"))
	if err != nil {
		return fmt.Errorf("erro ao criar index.html: %w", err)
	}
	if err := pagina.Render(context.Background(), arquivo); err != nil {
		arquivo.Close()
		return fmt.Errorf("erro ao renderizar a vitrine: %w", err)
	}
	if err := arquivo.Close(); err != nil {
		return fmt.Errorf("erro ao finalizar index.html: %w", err)
	}

	fmt.Printf("Vitrine exportada em %s\n", *diretorio)
	return nil
}
