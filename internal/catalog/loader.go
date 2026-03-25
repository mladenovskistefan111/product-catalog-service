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
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{}
	log.Out = os.Stdout
}

// LoadCatalog connects to Postgres via DATABASE_URL and loads all products.
func LoadCatalog(catalog *pb.ListProductsResponse) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	rows, err := pool.Query(ctx, `
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