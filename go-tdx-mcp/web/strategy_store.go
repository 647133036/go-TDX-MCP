package web

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SavedStrategy struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Kind          string                 `json:"kind"`
	Strategy      string                 `json:"strategy"`
	StrategyLabel string                 `json:"strategy_label"`
	Params        map[string]interface{} `json:"params"`
	Context       map[string]interface{} `json:"context"`
	TradeConfig   map[string]interface{} `json:"trade_config"`
	Snapshot      map[string]interface{} `json:"snapshot"`
	Tags          []string               `json:"tags"`
	Notes         string                 `json:"notes"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	AppVersion    string                 `json:"app_version"`
}

type StrategyStore struct {
	db *sql.DB
	mu sync.RWMutex
}

const strategySchema = `
CREATE TABLE IF NOT EXISTS strategies (
	id             TEXT PRIMARY KEY,
	name           TEXT NOT NULL,
	kind           TEXT NOT NULL,
	strategy       TEXT NOT NULL,
	strategy_label TEXT NOT NULL DEFAULT '',
	params         TEXT NOT NULL DEFAULT '{}',
	context        TEXT NOT NULL DEFAULT '{}',
	trade_config   TEXT NOT NULL DEFAULT '{}',
	snapshot       TEXT NOT NULL DEFAULT '{}',
	tags           TEXT NOT NULL DEFAULT '[]',
	notes          TEXT NOT NULL DEFAULT '',
	created_at     TEXT NOT NULL DEFAULT '',
	updated_at     TEXT NOT NULL DEFAULT '',
	app_version    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_strategies_kind ON strategies(kind);
CREATE INDEX IF NOT EXISTS idx_strategies_strategy ON strategies(strategy);
CREATE INDEX IF NOT EXISTS idx_strategies_created ON strategies(created_at);
`

var globalStore *StrategyStore
var storeOnce sync.Once

func NewStrategyStore(dbPath string) (*StrategyStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite3", dbPath+"?_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.Exec(strategySchema); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &StrategyStore{db: db}, nil
}

func GetStrategyStore(dbPath string) (*StrategyStore, error) {
	var err error
	storeOnce.Do(func() {
		globalStore, err = NewStrategyStore(dbPath)
	})
	return globalStore, err
}

func CloseStrategyStore() {
	if globalStore != nil && globalStore.db != nil {
		globalStore.db.Close()
	}
}

func (s *StrategyStore) Add(rec *SavedStrategy) (*SavedStrategy, error) {
	now := nowISO()
	if rec.ID == "" {
		rec.ID = newShortID()
	}
	if rec.CreatedAt == "" {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now
	paramsJSON, _ := json.Marshal(rec.Params)
	ctxJSON, _ := json.Marshal(rec.Context)
	tradeJSON, _ := json.Marshal(rec.TradeConfig)
	snapJSON, _ := json.Marshal(rec.Snapshot)
	tagsJSON, _ := json.Marshal(rec.Tags)

	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO strategies
		(id, name, kind, strategy, strategy_label, params, context,
			trade_config, snapshot, tags, notes, created_at, updated_at, app_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rec.ID, rec.Name, rec.Kind, rec.Strategy, rec.StrategyLabel,
		string(paramsJSON), string(ctxJSON), string(tradeJSON), string(snapJSON), string(tagsJSON),
		rec.Notes, rec.CreatedAt, rec.UpdatedAt, rec.AppVersion)
	return rec, err
}

func (s *StrategyStore) List() ([]SavedStrategy, error) {
	rows, err := s.db.Query("SELECT * FROM strategies ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStrategyRows(rows)
}

func (s *StrategyStore) Get(id string) (*SavedStrategy, error) {
	row := s.db.QueryRow("SELECT * FROM strategies WHERE id = ?", id)
	return scanStrategyOne(row)
}

func (s *StrategyStore) Delete(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, err := s.db.Exec("DELETE FROM strategies WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func scanStrategyRows(rows *sql.Rows) ([]SavedStrategy, error) {
	var results []SavedStrategy
	for rows.Next() {
		rec, err := scanStrategyRow(rows)
		if err != nil {
			continue
		}
		results = append(results, rec)
	}
	if results == nil {
		results = []SavedStrategy{}
	}
	return results, nil
}

func scanStrategyRow(rows *sql.Rows) (SavedStrategy, error) {
	var id, name, kind, strategy, label, params, ctx, tradeCfg, snapshot, tags, notes, created, updated, appVer string
	err := rows.Scan(&id, &name, &kind, &strategy, &label, &params, &ctx, &tradeCfg, &snapshot, &tags, &notes, &created, &updated, &appVer)
	if err != nil {
		return SavedStrategy{}, err
	}
	rec := newEmptyStrategy()
	rec.ID = id
	rec.Name = name
	rec.Kind = kind
	rec.Strategy = strategy
	rec.StrategyLabel = label
	rec.Notes = notes
	rec.CreatedAt = created
	rec.UpdatedAt = updated
	rec.AppVersion = appVer
	json.Unmarshal([]byte(params), &rec.Params)
	json.Unmarshal([]byte(ctx), &rec.Context)
	json.Unmarshal([]byte(tradeCfg), &rec.TradeConfig)
	json.Unmarshal([]byte(snapshot), &rec.Snapshot)
	json.Unmarshal([]byte(tags), &rec.Tags)
	if rec.Snapshot == nil {
		rec.Snapshot = map[string]interface{}{}
	}
	if rec.Tags == nil {
		rec.Tags = []string{}
	}
	return rec, nil
}

func scanStrategyOne(row *sql.Row) (*SavedStrategy, error) {
	var id, name, kind, strategy, label, params, ctx, tradeCfg, snapshot, tags, notes, created, updated, appVer string
	err := row.Scan(&id, &name, &kind, &strategy, &label, &params, &ctx, &tradeCfg, &snapshot, &tags, &notes, &created, &updated, &appVer)
	if err != nil {
		return nil, err
	}
	rec := newEmptyStrategy()
	rec.ID = id
	rec.Name = name
	rec.Kind = kind
	rec.Strategy = strategy
	rec.StrategyLabel = label
	rec.Notes = notes
	rec.CreatedAt = created
	rec.UpdatedAt = updated
	rec.AppVersion = appVer
	json.Unmarshal([]byte(params), &rec.Params)
	json.Unmarshal([]byte(ctx), &rec.Context)
	json.Unmarshal([]byte(tradeCfg), &rec.TradeConfig)
	json.Unmarshal([]byte(snapshot), &rec.Snapshot)
	json.Unmarshal([]byte(tags), &rec.Tags)
	if rec.Snapshot == nil {
		rec.Snapshot = map[string]interface{}{}
	}
	if rec.Tags == nil {
		rec.Tags = []string{}
	}
	return &rec, nil
}

func newEmptyStrategy() SavedStrategy {
	return SavedStrategy{
		Params:      map[string]interface{}{},
		Context:     map[string]interface{}{},
		TradeConfig: map[string]interface{}{},
		Snapshot:    map[string]interface{}{},
		Tags:        []string{},
	}
}

func newShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
