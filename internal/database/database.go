package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Connect estabelece conexão com PostgreSQL
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	// Configurar pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao pingar banco: %w", err)
	}

	return db, nil
}

// ConversationLog representa um log de conversa
type ConversationLog struct {
	ID            int64
	IdosoID       int64
	IdosoNome     string
	Timestamp     time.Time
	Speaker       string
	Content       string
	Emotion       string
	Importance    float64
	SessionID     string
	CallHistoryID int64
}

// BehavioralNote representa uma nota de comportamento
type BehavioralNote struct {
	ID          int64
	IdosoID     int64
	Category    string
	Observation string
	Confidence  float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Active      bool
}

// PromptTemplate representa um template de prompt personalizado
type PromptTemplate struct {
	ID           int64
	IdosoID      int64
	TemplateType string
	Content      string
	Version      int
	Score        float64
	CreatedAt    time.Time
	Active       bool
}

// InteractionScore representa a avaliação de uma interação
type InteractionScore struct {
	ID                int64
	IdosoID           int64
	CallHistoryID     int64
	OverallScore      float64
	AdherenceScore    float64
	SatisfactionScore float64
	EngagementScore   float64
	Notes             string
	CreatedAt         time.Time
}
