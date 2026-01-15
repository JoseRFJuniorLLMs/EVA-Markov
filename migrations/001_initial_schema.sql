-- EVA-Markov: Schema de Tabelas para Meta-Agent

-- Tabela de Notas Comportamentais
CREATE TABLE IF NOT EXISTS behavioral_notes (
    id BIGSERIAL PRIMARY KEY,
    idoso_id BIGINT NOT NULL REFERENCES idosos(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- communication_style, preferences, triggers, failures
    observation TEXT NOT NULL,
    confidence DECIMAL(3,2) CHECK (confidence >= 0 AND confidence <= 1),
    evidence JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    active BOOLEAN DEFAULT true
);

CREATE INDEX idx_behavioral_notes_idoso ON behavioral_notes(idoso_id);
CREATE INDEX idx_behavioral_notes_category ON behavioral_notes(category);
CREATE INDEX idx_behavioral_notes_active ON behavioral_notes(active);

-- Tabela de Prompts Personalizados
CREATE TABLE IF NOT EXISTS prompt_templates_personalized (
    id BIGSERIAL PRIMARY KEY,
    idoso_id BIGINT NOT NULL REFERENCES idosos(id) ON DELETE CASCADE,
    template_type VARCHAR(50) DEFAULT 'optimized',
    content TEXT NOT NULL,
    version INT DEFAULT 1,
    score DECIMAL(3,2), -- Score de qualidade (0-10)
    created_at TIMESTAMP DEFAULT NOW(),
    active BOOLEAN DEFAULT true
);

CREATE INDEX idx_prompt_templates_idoso ON prompt_templates_personalized(idoso_id);
CREATE INDEX idx_prompt_templates_active ON prompt_templates_personalized(active);
CREATE UNIQUE INDEX idx_prompt_templates_active_idoso ON prompt_templates_personalized(idoso_id, active) WHERE active = true;

-- Tabela de Scores de Interação
CREATE TABLE IF NOT EXISTS interaction_scores (
    id BIGSERIAL PRIMARY KEY,
    idoso_id BIGINT NOT NULL REFERENCES idosos(id) ON DELETE CASCADE,
    call_history_id BIGINT REFERENCES historico_ligacoes(id) ON DELETE SET NULL,
    overall_score DECIMAL(3,2) CHECK (overall_score >= 0 AND overall_score <= 10),
    adherence_score DECIMAL(3,2), -- Adesão a medicamentos
    satisfaction_score DECIMAL(3,2), -- Satisfação do idoso
    engagement_score DECIMAL(3,2), -- Engajamento na conversa
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_interaction_scores_idoso ON interaction_scores(idoso_id);
CREATE INDEX idx_interaction_scores_call ON interaction_scores(call_history_id);

-- Tabela de Histórico de Otimizações
CREATE TABLE IF NOT EXISTS optimization_history (
    id BIGSERIAL PRIMARY KEY,
    idoso_id BIGINT NOT NULL REFERENCES idosos(id) ON DELETE CASCADE,
    optimization_type VARCHAR(50), -- prompt_update, behavior_note, etc
    before_state JSONB,
    after_state JSONB,
    insights_used JSONB,
    success BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_optimization_history_idoso ON optimization_history(idoso_id);
CREATE INDEX idx_optimization_history_type ON optimization_history(optimization_type);

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger para behavioral_notes
CREATE TRIGGER update_behavioral_notes_updated_at 
    BEFORE UPDATE ON behavioral_notes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- View para análise rápida
CREATE OR REPLACE VIEW v_idoso_optimization_status AS
SELECT 
    i.id as idoso_id,
    i.nome,
    COUNT(DISTINCT bn.id) as total_behavioral_notes,
    COUNT(DISTINCT ptp.id) as total_personalized_prompts,
    AVG(iscr.overall_score) as avg_interaction_score,
    MAX(ptp.created_at) as last_optimization_date
FROM idosos i
LEFT JOIN behavioral_notes bn ON bn.idoso_id = i.id AND bn.active = true
LEFT JOIN prompt_templates_personalized ptp ON ptp.idoso_id = i.id AND ptp.active = true
LEFT JOIN interaction_scores iscr ON iscr.idoso_id = i.id AND iscr.created_at >= NOW() - INTERVAL '7 days'
GROUP BY i.id, i.nome;

COMMENT ON TABLE behavioral_notes IS 'Notas de comportamento geradas pelo Meta-Agente';
COMMENT ON TABLE prompt_templates_personalized IS 'Prompts otimizados por idoso';
COMMENT ON TABLE interaction_scores IS 'Scores de qualidade das interações';
COMMENT ON TABLE optimization_history IS 'Histórico de otimizações realizadas';
