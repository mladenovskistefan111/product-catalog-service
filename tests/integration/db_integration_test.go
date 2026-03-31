//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"product-catalog-service/internal/catalog"
	pb "product-catalog-service/proto"
)

func TestInitDB_ConnectsSuccessfully(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := catalog.InitDB(ctx); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
}

func TestLoadCatalog_ReturnsProducts(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := catalog.InitDB(ctx); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	var resp pb.ListProductsResponse
	if err := catalog.LoadCatalog(&resp); err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}
	if len(resp.Products) == 0 {
		t.Fatal("expected products, got empty list")
	}
	for _, p := range resp.Products {
		if p.Id == "" || p.Name == "" || p.PriceUsd == nil {
			t.Errorf("product has missing required fields: %+v", p)
		}
	}
}
