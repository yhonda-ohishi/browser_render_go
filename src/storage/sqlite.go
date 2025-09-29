package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

type Session struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	UserID    string
	CompanyID string
}

type Cookie struct {
	Name      string
	Value     string
	Domain    string
	Path      string
	ExpiresAt time.Time
	HTTPOnly  bool
	Secure    bool
}

type VehicleCache struct {
	VehicleCD string
	Data      string
	CachedAt  time.Time
	ExpiresAt time.Time
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Optimize SQLite
	if _, err := db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA cache_size=10000;
		PRAGMA temp_store=MEMORY;
	`); err != nil {
		return nil, fmt.Errorf("failed to optimize database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return s, nil
}

func (s *Storage) initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			user_id TEXT,
			company_id TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS cookies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			domain TEXT,
			path TEXT,
			expires_at TIMESTAMP,
			http_only BOOLEAN DEFAULT 0,
			secure BOOLEAN DEFAULT 0,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS vehicle_cache (
			vehicle_cd TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS kv_store (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ttl INTEGER
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// Session methods
func (s *Storage) CreateSession(session *Session) error {
	query := `
		INSERT INTO sessions (id, created_at, updated_at, expires_at, user_id, company_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		session.ID,
		session.CreatedAt,
		session.UpdatedAt,
		session.ExpiresAt,
		session.UserID,
		session.CompanyID,
	)
	return err
}

func (s *Storage) GetSession(sessionID string) (*Session, error) {
	query := `
		SELECT id, created_at, updated_at, expires_at, user_id, company_id
		FROM sessions
		WHERE id = ? AND expires_at > ?
	`
	var session Session
	err := s.db.QueryRow(query, sessionID, time.Now()).Scan(
		&session.ID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.ExpiresAt,
		&session.UserID,
		&session.CompanyID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *Storage) DeleteSession(sessionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete cookies first
	if _, err := tx.Exec("DELETE FROM cookies WHERE session_id = ?", sessionID); err != nil {
		return err
	}

	// Delete session
	if _, err := tx.Exec("DELETE FROM sessions WHERE id = ?", sessionID); err != nil {
		return err
	}

	return tx.Commit()
}

// Cookie methods
func (s *Storage) SaveCookies(sessionID string, cookies []Cookie) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing cookies
	if _, err := tx.Exec("DELETE FROM cookies WHERE session_id = ?", sessionID); err != nil {
		return err
	}

	// Insert new cookies
	query := `
		INSERT INTO cookies (session_id, name, value, domain, path, expires_at, http_only, secure)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, cookie := range cookies {
		_, err := stmt.Exec(
			sessionID,
			cookie.Name,
			cookie.Value,
			cookie.Domain,
			cookie.Path,
			cookie.ExpiresAt,
			cookie.HTTPOnly,
			cookie.Secure,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetCookies(sessionID string) ([]Cookie, error) {
	query := `
		SELECT name, value, domain, path, expires_at, http_only, secure
		FROM cookies
		WHERE session_id = ?
	`
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cookies []Cookie
	for rows.Next() {
		var cookie Cookie
		err := rows.Scan(
			&cookie.Name,
			&cookie.Value,
			&cookie.Domain,
			&cookie.Path,
			&cookie.ExpiresAt,
			&cookie.HTTPOnly,
			&cookie.Secure,
		)
		if err != nil {
			return nil, err
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

// KV Store methods
func (s *Storage) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO kv_store (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`
	_, err = s.db.Exec(query, key, string(data), time.Now())
	return err
}

func (s *Storage) Get(key string, value interface{}) error {
	query := `SELECT value FROM kv_store WHERE key = ?`
	var data string
	err := s.db.QueryRow(query, key).Scan(&data)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), value)
}

func (s *Storage) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM kv_store WHERE key = ?", key)
	return err
}

// Vehicle cache methods
func (s *Storage) CacheVehicleData(vehicleCD string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO vehicle_cache (vehicle_cd, data, cached_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(vehicle_cd) DO UPDATE SET
			data = excluded.data,
			cached_at = excluded.cached_at,
			expires_at = excluded.expires_at
	`
	_, err = s.db.Exec(query, vehicleCD, string(jsonData), time.Now(), time.Now().Add(ttl))
	return err
}

func (s *Storage) GetCachedVehicleData(vehicleCD string) (string, error) {
	query := `
		SELECT data FROM vehicle_cache
		WHERE vehicle_cd = ? AND expires_at > ?
	`
	var data string
	err := s.db.QueryRow(query, vehicleCD, time.Now()).Scan(&data)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return data, err
}

// Cleanup expired data
func (s *Storage) CleanupExpired() error {
	queries := []string{
		"DELETE FROM sessions WHERE expires_at < ?",
		"DELETE FROM vehicle_cache WHERE expires_at < ?",
		"DELETE FROM cookies WHERE expires_at < ?",
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query, time.Now()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}