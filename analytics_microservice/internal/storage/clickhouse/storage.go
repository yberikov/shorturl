package clickhouse

import (
	"context"
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"time"

	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickhouseStorage struct {
	db driver.Conn
}

func New(addr string) (*ClickhouseStorage, error) {
	conn, err := connect(addr)
	if err != nil {
		return nil, err
	}

	storage := &ClickhouseStorage{db: conn}
	return storage, storage.init(context.TODO())
}

func WaitForStorage(addr string, timeout time.Duration) (*ClickhouseStorage, error) {
	end := time.Now().Add(timeout)

	for time.Now().Before(end) {
		storage, err := New(addr)
		if err == nil {
			return storage, nil
		}
		time.Sleep(1 * time.Second) // Wait before retrying
	}
	return nil, fmt.Errorf("failed to connect to ClickHouse at %s after %v", addr, timeout)
}

func connect(addr string) (driver.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{addr},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "user",
				Password: "12345",
			},
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "an-example-go-client", Version: "0.1"},
				},
			},

			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}

func (c *ClickhouseStorage) init(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS source (
			value String
		) ENGINE = Memory`,
		`CREATE TABLE IF NOT EXISTS counters (
			id String,
			user_id Int64,
			counter AggregateFunction(sum, Int64)
		) ENGINE = AggregatingMergeTree()
		ORDER BY (id, user_id)`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS counters_mv TO counters
		AS SELECT
			JSONExtractString(value, 'url') AS id,
			JSONExtractInt(value, 'user_id') AS user_id,
			sumState(toInt64(1)) AS counter
		FROM source
		GROUP BY id, user_id`,
	}

	for _, query := range queries {
		if err := c.db.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (c *ClickhouseStorage) SaveStats(ctx context.Context, userID int64, urlText string) error {
	value := fmt.Sprintf(`{"url":"%s", "user_id": %d}`, urlText, userID)
	query := `INSERT INTO source (value) VALUES (?)`
	err := c.db.Exec(ctx, query, value)
	return err
}

func (c *ClickhouseStorage) GetURLStats(ctx context.Context, url string) (int64, error) {
	var total uint64
	query := `SELECT sumMerge(counter) as counter FROM counters WHERE id = ? AND user_id = 0`
	err := c.db.QueryRow(ctx, query, url).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return int64(total), nil
}

func (c *ClickhouseStorage) LogURLAccess(ctx context.Context, url string, userId int64) (bool, error) {
	var count uint64
	query := `SELECT count(*) FROM counters WHERE id = ? AND user_id = ?`
	err := c.db.QueryRow(ctx, query, url, userId).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
