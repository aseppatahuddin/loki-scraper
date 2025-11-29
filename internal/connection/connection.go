package connection

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect() (driver.Conn, error) {
	var host = os.Getenv("CLICKHOUSE_HOST")
	var db = os.Getenv("CLICKHOUSE_DATABASE")
	var user = os.Getenv("CLICKHOUSE_USER")
	var pass = os.Getenv("CLICKHOUSE_PASSWORD")

	log.Printf("Connect to Clickhouse with detail Host: %s, DB: %s, User: %s", host, db, user)

	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{host},
			Auth: clickhouse.Auth{
				Database: db,
				Username: user,
				Password: pass,
			},
			Protocol: clickhouse.HTTP,
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "loki-exporter", Version: "0.1"},
				},
			},
		})
	)

	// defer conn.Close()

	if err != nil {
		return conn, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return conn, err
	}
	return conn, nil
}
