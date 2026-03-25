package catalog

import (
	"context"
	"testing"

	pb "product-catalog-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newMockCatalog builds a ProductCatalog with hardcoded products — no file I/O.
func newMockCatalog() *ProductCatalog {
	return &ProductCatalog{
		catalog: pb.ListProductsResponse{
			Products: []*pb.Product{
				{Id: "abc001", Name: "Product Alpha One"},
				{Id: "abc002", Name: "Product Delta"},
				{Id: "abc003", Name: "Product Alpha Two"},
				{Id: "abc004", Name: "Product Gamma"},
			},
		},
	}
}

func TestGetProductExists(t *testing.T) {
	svc := newMockCatalog()
	product, err := svc.GetProduct(context.Background(), &pb.GetProductRequest{Id: "abc003"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := product.Name, "Product Alpha Two"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetProductNotFound(t *testing.T) {
	svc := newMockCatalog()
	_, err := svc.GetProduct(context.Background(), &pb.GetProductRequest{Id: "doesnotexist"})
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestListProducts(t *testing.T) {
	svc := newMockCatalog()
	resp, err := svc.ListProducts(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(resp.Products), 4; got != want {
		t.Errorf("got %d products, want %d", got, want)
	}
}

func TestSearchProducts(t *testing.T) {
	svc := newMockCatalog()
	resp, err := svc.SearchProducts(context.Background(), &pb.SearchProductsRequest{Query: "alpha"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(resp.Results), 2; got != want {
		t.Errorf("got %d results, want %d", got, want)
	}
}