package analyzer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"eva-markov/internal/config"
	"eva-markov/internal/database"
	"eva-markov/internal/llm"
)

type Service struct {
	db  *sql.DB
	cfg *config.Config
	llm *llm.Client
}

type Insight struct {
	IdosoID     int64
	Category    string
	Observation string
	Confidence  float64
	Evidence    []string
}

func NewService(db *sql.DB, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		cfg: cfg,
		llm: llm.NewClient(cfg.GoogleAPIKey),
	}
}

// AnalyzeDailyConversations analisa todas as conversas do dia anterior
func (s *Service) AnalyzeDailyConversations(ctx context.Context) ([]Insight, error) {
	log.Println("📊 Buscando conversas das últimas 24h...")

	// Buscar conversas recentes
	conversations, err := s.fetchRecentConversations(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conversas: %w", err)
	}

	log.Printf("📝 Encontradas %d conversas para análise", len(conversations))

	// Agrupar por idoso
	conversationsByIdoso := s.groupByIdoso(conversations)

	var allInsights []Insight

	// Analisar cada idoso
	for idosoID, convs := range conversationsByIdoso {
		log.Printf("🔍 Analisando %d conversas do idoso %d...", len(convs), idosoID)

		insights, err := s.analyzeIdosoConversations(ctx, idosoID, convs)
		if err != nil {
			log.Printf("⚠️ Erro ao analisar idoso %d: %v", idosoID, err)
			continue
		}

		allInsights = append(allInsights, insights...)

		// Salvar insights no banco
		if err := s.saveInsights(ctx, insights); err != nil {
			log.Printf("⚠️ Erro ao salvar insights do idoso %d: %v", idosoID, err)
		}
	}

	return allInsights, nil
}

func (s *Service) fetchRecentConversations(ctx context.Context) ([]database.ConversationLog, error) {
	query := `
		SELECT 
			em.id,
			em.idoso_id,
			i.nome as idoso_nome,
			em.timestamp,
			em.speaker,
			em.content,
			em.emotion,
			em.importance,
			em.session_id,
			em.call_history_id
		FROM episodic_memories em
		JOIN idosos i ON i.id = em.idoso_id
		WHERE em.timestamp >= NOW() - INTERVAL '1 day'
		ORDER BY em.idoso_id, em.timestamp
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []database.ConversationLog
	for rows.Next() {
		var c database.ConversationLog
		err := rows.Scan(
			&c.ID,
			&c.IdosoID,
			&c.IdosoNome,
			&c.Timestamp,
			&c.Speaker,
			&c.Content,
			&c.Emotion,
			&c.Importance,
			&c.SessionID,
			&c.CallHistoryID,
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}

	return conversations, nil
}

func (s *Service) groupByIdoso(conversations []database.ConversationLog) map[int64][]database.ConversationLog {
	grouped := make(map[int64][]database.ConversationLog)
	for _, conv := range conversations {
		grouped[conv.IdosoID] = append(grouped[conv.IdosoID], conv)
	}
	return grouped
}

func (s *Service) analyzeIdosoConversations(ctx context.Context, idosoID int64, conversations []database.ConversationLog) ([]Insight, error) {
	// Construir transcrição completa
	transcript := s.buildTranscript(conversations)

	// Prompt para análise comportamental
	prompt := fmt.Sprintf(`Você é um psicólogo especialista em gerontologia e análise comportamental.

Analise a seguinte transcrição de conversas com um idoso e identifique:

1. Padrões de comunicação que funcionaram bem
2. Padrões que não funcionaram (ex: ignorou lembretes, ficou irritado)
3. Preferências de tom de voz e estilo
4. Gatilhos emocionais (positivos e negativos)
5. Tópicos de interesse

Transcrição:
%s

Retorne um JSON com o seguinte formato:
{
  "insights": [
    {
      "category": "communication_style|preferences|triggers|failures",
      "observation": "descrição curta e objetiva",
      "confidence": 0.0-1.0,
      "evidence": ["trecho da conversa que suporta isso"]
    }
  ]
}`, transcript)

	// Chamar LLM
	response, err := s.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar análise: %w", err)
	}

	// Parse JSON
	var result struct {
		Insights []struct {
			Category    string   `json:"category"`
			Observation string   `json:"observation"`
			Confidence  float64  `json:"confidence"`
			Evidence    []string `json:"evidence"`
		} `json:"insights"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	// Converter para Insights
	var insights []Insight
	for _, i := range result.Insights {
		insights = append(insights, Insight{
			IdosoID:     idosoID,
			Category:    i.Category,
			Observation: i.Observation,
			Confidence:  i.Confidence,
			Evidence:    i.Evidence,
		})
	}

	return insights, nil
}

func (s *Service) buildTranscript(conversations []database.ConversationLog) string {
	var transcript string
	for _, c := range conversations {
		timestamp := c.Timestamp.Format("15:04")
		transcript += fmt.Sprintf("[%s] %s: %s\n", timestamp, c.Speaker, c.Content)
	}
	return transcript
}

func (s *Service) saveInsights(ctx context.Context, insights []Insight) error {
	query := `
		INSERT INTO behavioral_notes (idoso_id, category, observation, confidence, created_at, active)
		VALUES ($1, $2, $3, $4, NOW(), true)
	`

	for _, insight := range insights {
		_, err := s.db.ExecContext(ctx, query,
			insight.IdosoID,
			insight.Category,
			insight.Observation,
			insight.Confidence,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
