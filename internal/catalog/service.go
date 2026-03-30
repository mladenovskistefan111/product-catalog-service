package catalog

import (
	"context"
	"strings"
	"time"

	pb "product-catalog-service/proto"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// ProductCatalog implements the gRPC ProductCatalogService.
type ProductCatalog struct {
	pb.UnimplementedProductCatalogServiceServer
	catalog      pb.ListProductsResponse
	extraLatency time.Duration
	reloadFlag   *bool
}

// NewProductCatalog initialises the DB pool and loads the catalog into memory.
func NewProductCatalog(ctx context.Context, extraLatency time.Duration, reloadFlag *bool) (*ProductCatalog, error) {
	if err := InitDB(ctx); err != nil {
		return nil, err
	}

	svc := &ProductCatalog{
		extraLatency: extraLatency,
		reloadFlag:   reloadFlag,
	}
	if err := LoadCatalog(&svc.catalog); err != nil {
		return nil, err
	}
	return svc, nil
}

// --- Health ---

func (p *ProductCatalog) Check(_ context.Context, _ *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *ProductCatalog) Watch(_ *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

// --- RPC methods ---

func (p *ProductCatalog) ListProducts(_ context.Context, _ *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(p.extraLatency)
	products, err := p.products()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load catalog: %v", err)
	}
	return &pb.ListProductsResponse{Products: products}, nil
}

func (p *ProductCatalog) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(p.extraLatency)
	products, err := p.products()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load catalog: %v", err)
	}
	for _, product := range products {
		if product.Id == req.Id {
			return product, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
}

func (p *ProductCatalog) SearchProducts(_ context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(p.extraLatency)
	products, err := p.products()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load catalog: %v", err)
	}
	query := strings.ToLower(req.Query)
	var results []*pb.Product
	for _, product := range products {
		if strings.Contains(strings.ToLower(product.Name), query) ||
			strings.Contains(strings.ToLower(product.Description), query) {
			results = append(results, product)
		}
	}
	return &pb.SearchProductsResponse{Results: results}, nil
}

// --- Internal ---

// products returns the in-memory catalog, reloading from Postgres only when
// the SIGUSR1 reload flag is set (triggered externally) or the cache is empty.
func (p *ProductCatalog) products() ([]*pb.Product, error) {
	shouldReload := p.reloadFlag != nil && *p.reloadFlag
	if shouldReload || len(p.catalog.Products) == 0 {
		if err := LoadCatalog(&p.catalog); err != nil {
			return nil, err
		}
		if p.reloadFlag != nil {
			*p.reloadFlag = false
		}
	}
	return p.catalog.Products, nil
}
