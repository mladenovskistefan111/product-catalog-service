package catalog

import (
	"context"
	"os"
	"strings"
	"sync"

	pb "product-catalog-service/proto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var (
	log          *logrus.Logger
	catalogMutex sync.Mutex
	pool         *pgxpool.Pool
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{}
	log.Out = os.Stdout
}

// InitDB creates the shared connection pool. Call once at startup.
func InitDB(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return err
	}

	pool = p
	log.Info("connected to postgres")
	return nil
}

// LoadCatalog queries Postgres using the shared pool and populates the catalog.
// It is safe to call from multiple goroutines.
func LoadCatalog(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	rows, err := pool.Query(context.Background(), `
		SELECT id, name, description, picture,
		       price_usd_currency_code, price_usd_units, price_usd_nanos,
		       categories
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	catalog.Products = catalog.Products[:0]
	for rows.Next() {
		p := &pb.Product{PriceUsd: &pb.Money{}}
		var categories string
		if err := rows.Scan(
			&p.Id, &p.Name, &p.Description, &p.Picture,
			&p.PriceUsd.CurrencyCode, &p.PriceUsd.Units, &p.PriceUsd.Nanos,
			&categories,
		); err != nil {
			return err
		}
		p.Categories = strings.Split(strings.ToLower(categories), ",")
		catalog.Products = append(catalog.Products, p)
	}

	log.Infof("loaded %d products from postgres", len(catalog.Products))
	return rows.Err()
}