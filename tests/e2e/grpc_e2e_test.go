//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	pb "product-catalog-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// grpcHost returns the service address from env or defaults to localhost.
// In the pipeline this will be set to the kind service address.
func grpcHost() string {
	if h := os.Getenv("GRPC_HOST"); h != "" {
		return h
	}
	return "localhost:3550"
}

// newClient creates a gRPC client connection for the test suite.
func newClient(t *testing.T) (pb.ProductCatalogServiceClient, func()) {
	t.Helper()

	conn, err := grpc.NewClient(
		grpcHost(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to connect to %s: %v", grpcHost(), err)
	}

	client := pb.NewProductCatalogServiceClient(conn)
	return client, func() { conn.Close() }
}

// ── ListProducts ──────────────────────────────────────────────────────────────

func TestListProducts_ReturnsAllProducts(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	resp, err := client.ListProducts(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}
	if len(resp.Products) == 0 {
		t.Fatal("expected products, got empty list")
	}
	t.Logf("ListProducts returned %d products", len(resp.Products))
}

func TestListProducts_EachProductHasRequiredFields(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	resp, err := client.ListProducts(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	for _, p := range resp.Products {
		if p.Id == "" {
			t.Errorf("product has empty ID: %+v", p)
		}
		if p.Name == "" {
			t.Errorf("product %s has empty name", p.Id)
		}
		if p.PriceUsd == nil {
			t.Errorf("product %s has nil price", p.Id)
		}
	}
}

// ── GetProduct ────────────────────────────────────────────────────────────────

func TestGetProduct_ExistingID_ReturnsProduct(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	product, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: "OLJCESPC7Z"})
	if err != nil {
		t.Fatalf("GetProduct failed: %v", err)
	}
	if product.Id != "OLJCESPC7Z" {
		t.Errorf("got ID %s, want OLJCESPC7Z", product.Id)
	}
	if product.Name != "Sunglasses" {
		t.Errorf("got name %s, want Sunglasses", product.Name)
	}
}

func TestGetProduct_UnknownID_ReturnsNotFound(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: "doesnotexist"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got status %s, want %s", got, want)
	}
}

func TestGetProduct_EmptyID_ReturnsNotFound(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: ""})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got status %s, want %s", got, want)
	}
}

// ── SearchProducts ────────────────────────────────────────────────────────────

func TestSearchProducts_MatchingQuery_ReturnsResults(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	resp, err := client.SearchProducts(context.Background(), &pb.SearchProductsRequest{Query: "mug"})
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected results for 'mug', got empty")
	}
	t.Logf("SearchProducts('mug') returned %d results", len(resp.Results))
}

func TestSearchProducts_CaseInsensitive(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	lower, err := client.SearchProducts(context.Background(), &pb.SearchProductsRequest{Query: "mug"})
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}

	upper, err := client.SearchProducts(context.Background(), &pb.SearchProductsRequest{Query: "MUG"})
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}

	if len(lower.Results) != len(upper.Results) {
		t.Errorf("case sensitivity issue: 'mug' returned %d, 'MUG' returned %d",
			len(lower.Results), len(upper.Results))
	}
}

func TestSearchProducts_NoMatch_ReturnsEmpty(t *testing.T) {
	client, cleanup := newClient(t)
	defer cleanup()

	resp, err := client.SearchProducts(context.Background(), &pb.SearchProductsRequest{Query: "zzznomatch"})
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}
