package optimizer

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"eva-markov/internal/analyzer"
	"eva-markov/internal/config"
	"eva-markov/internal/llm"
)

type Service struct {
	db  *sql.DB
	cfg *config.Config
	llm *llm.Client
}

func NewService(db *sql.DB, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		cfg: cfg,
		llm: llm.NewClient(cfg.GoogleAPIKey),
	}
}

// OptimizePrompts gera novos prompts personalizados baseado nos insights
func (s *Service) OptimizePrompts(ctx context.Context, insights []analyzer.Insight) (int, error) {
	log.Println("🔧 Iniciando otimização de prompts...")

	// Agrupar insights por idoso
	insightsByIdoso := s.groupInsightsByIdoso(insights)

	optimizedCount := 0

	for idosoID, idosoInsights := range insightsByIdoso {
		// Verificar se há insights suficientes
		if len(idosoInsights) < 2 {
			log.Printf("⏭️ Pulando idoso %d (insights insuficientes)", idosoID)
			continue
		}

		// Buscar prompt atual
		currentPrompt, err := s.getCurrentPrompt(ctx, idosoID)
		if err != nil {
			log.Printf("⚠️ Erro ao buscar prompt do idoso %d: %v", idosoID, err)
			continue
		}

		// Gerar novo prompt otimizado
		newPrompt, err := s.generateOptimizedPrompt(ctx, idosoID, currentPrompt, idosoInsights)
		if err != nil {
			log.Printf("⚠️ Erro ao gerar prompt para idoso %d: %v", idosoID, err)
			continue
		}

		// Salvar novo prompt
		if err := s.savePrompt(ctx, idosoID, newPrompt); err != nil {
			log.Printf("⚠️ Erro ao salvar prompt do idoso %d: %v", idosoID, err)
			continue
		}

		log.Printf("✅ Prompt otimizado para idoso %d", idosoID)
		optimizedCount++
	}

	return optimizedCount, nil
}

func (s *Service) groupInsightsByIdoso(insights []analyzer.Insight) map[int64][]analyzer.Insight {
	grouped := make(map[int64][]analyzer.Insight)
	for _, insight := range insights {
		grouped[insight.IdosoID] = append(grouped[insight.IdosoID], insight)
	}
	return grouped
}

func (s *Service) getCurrentPrompt(ctx context.Context, idosoID int64) (string, error) {
	query := `
		SELECT content 
		FROM prompt_templates_personalized 
		WHERE idoso_id = $1 AND active = true 
		ORDER BY version DESC 
		LIMIT 1
	`

	var prompt string
	err := s.db.QueryRowContext(ctx, query, idosoID).Scan(&prompt)
	if err == sql.ErrNoRows {
		// Retornar prompt base se não houver personalizado
		return s.getBasePrompt(ctx)
	}
	if err != nil {
		return "", err
	}

	return prompt, nil
}

func (s *Service) getBasePrompt(ctx context.Context) (string, error) {
	query := `SELECT template FROM prompt_templates WHERE nome = 'eva_base_v2' AND ativo = true LIMIT 1`
	var prompt string
	err := s.db.QueryRowContext(ctx, query).Scan(&prompt)
	return prompt, err
}

func (s *Service) generateOptimizedPrompt(ctx context.Context, idosoID int64, currentPrompt string, insights []analyzer.Insight) (string, error) {
	// Construir contexto de insights
	insightsText := ""
	for _, insight := range insights {
		insightsText += fmt.Sprintf("- [%s] %s (confiança: %.0f%%)\n",
			insight.Category, insight.Observation, insight.Confidence*100)
	}

	// Prompt para o Meta-Agente
	metaPrompt := fmt.Sprintf(`Você é um Meta-Agente especializado em otimização de prompts para assistentes de saúde.

Seu objetivo é melhorar o prompt de sistema da EVA para um idoso específico, baseado em insights comportamentais.

PROMPT ATUAL:
%s

INSIGHTS COMPORTAMENTAIS:
%s

TAREFA:
Reescreva o prompt de sistema incorporando os insights acima. O novo prompt deve:
1. Manter a estrutura base da EVA
2. Adicionar instruções específicas baseadas nos insights
3. Ser conciso e objetivo
4. Focar em melhorar a adesão e satisfação

Retorne APENAS o novo prompt, sem explicações adicionais.`, currentPrompt, insightsText)

	// Gerar novo prompt
	newPrompt, err := s.llm.Generate(ctx, metaPrompt)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar prompt: %w", err)
	}

	return newPrompt, nil
}

func (s *Service) savePrompt(ctx context.Context, idosoID int64, prompt string) error {
	// Desativar prompts anteriores
	_, err := s.db.ExecContext(ctx, `
		UPDATE prompt_templates_personalized 
		SET active = false 
		WHERE idoso_id = $1
	`, idosoID)
	if err != nil {
		return err
	}

	// Inserir novo prompt
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO prompt_templates_personalized 
		(idoso_id, template_type, content, version, active, created_at)
		VALUES ($1, 'optimized', $2, 
			COALESCE((SELECT MAX(version) + 1 FROM prompt_templates_personalized WHERE idoso_id = $1), 1),
			true, NOW())
	`, idosoID, prompt)

	return err
}
