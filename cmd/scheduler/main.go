package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eva-markov/internal/analyzer"
	"eva-markov/internal/config"
	"eva-markov/internal/database"
	"eva-markov/internal/optimizer"

	"github.com/robfig/cron/v3"
)

func main() {
	log.Println("🧠 EVA-Markov Meta-Agent iniciando...")

	// Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Erro ao carregar config: %v", err)
	}

	// Conectar ao banco
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao banco: %v", err)
	}
	defer db.Close()

	// Inicializar serviços
	analyzerSvc := analyzer.NewService(db, cfg)
	optimizerSvc := optimizer.NewService(db, cfg)

	// Configurar cronjob
	c := cron.New()

	// Job principal: Análise e Otimização Noturna
	_, err = c.AddFunc(cfg.CronSchedule, func() {
		log.Println("⏰ Iniciando ciclo de otimização...")
		ctx := context.Background()

		// Fase 1: Análise
		log.Println("📊 Fase 1: Analisando conversas do dia...")
		insights, err := analyzerSvc.AnalyzeDailyConversations(ctx)
		if err != nil {
			log.Printf("⚠️ Erro na análise: %v", err)
			return
		}
		log.Printf("✅ Análise completa: %d insights gerados", len(insights))

		// Fase 2: Otimização
		log.Println("🔧 Fase 2: Otimizando prompts...")
		optimizations, err := optimizerSvc.OptimizePrompts(ctx, insights)
		if err != nil {
			log.Printf("⚠️ Erro na otimização: %v", err)
			return
		}
		log.Printf("✅ Otimização completa: %d prompts atualizados", optimizations)

		log.Println("🎉 Ciclo de otimização concluído com sucesso!")
	})

	if err != nil {
		log.Fatalf("❌ Erro ao configurar cronjob: %v", err)
	}

	// Iniciar scheduler
	c.Start()
	log.Printf("✅ Scheduler ativo (próxima execução: %s)", cfg.CronSchedule)

	// Modo de teste: executar imediatamente se solicitado
	if os.Getenv("RUN_NOW") == "true" {
		log.Println("🧪 Modo de teste: executando análise imediatamente...")
		ctx := context.Background()
		insights, _ := analyzerSvc.AnalyzeDailyConversations(ctx)
		optimizerSvc.OptimizePrompts(ctx, insights)
	}

	// Aguardar sinal de término
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Encerrando EVA-Markov...")
	c.Stop()
	log.Println("👋 Até logo!")
}
