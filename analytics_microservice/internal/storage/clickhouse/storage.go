package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
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
	var total int64
	query := `SELECT sumMerge(counter) as counter FROM counters WHERE id = ? AND user_id = 0`
	err := c.db.QueryRow(ctx, query, url).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
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
