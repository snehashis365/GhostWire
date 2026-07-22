package janitor

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Janitor struct {
	db       *sql.DB
	interval time.Duration
}

func New(db *sql.DB, interval time.Duration) *Janitor { return &Janitor{db: db, interval: interval} }

func (j *Janitor) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	j.cleanup()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.cleanup()
		}
	}
}

func (j *Janitor) cleanup() {
	res, err := j.db.Exec(`DELETE FROM messages WHERE expires_at < CURRENT_TIMESTAMP OR (read_at IS NOT NULL AND read_at < datetime('now', '-5 minutes'))`)
	if err != nil {
		log.Println("janitor cleanup:", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("janitor removed %d expired messages", n)
	}
}
