# API

Package `hipstershop`, service `ProductCatalogService` ([`proto/productcatalog.proto`](https://github.com/platform-boutique/app-product-catalog-service/blob/main/proto/productcatalog.proto)).

| Method | Request | Response | Behaviour |
|---|---|---|---|
| `ListProducts` | `Empty` | `{ repeated Product products }` | Returns all products |
| `GetProduct` | `{ id }` | `Product` | Exact ID match; gRPC `NOT_FOUND` ("no product with ID …") otherwise |
| `SearchProducts` | `{ query }` | `{ repeated Product results }` | Case-insensitive substring match on **name or description** |

`Product` = `{ id, name, description, picture, Money price_usd, repeated string categories }`. All prices are USD (`units` + `nanos`). `picture` is a relative path such as `/static/img/products/sunglasses.jpg` that is **served by `frontend`**, not by this service.

### Data source

Query run against Postgres on load:

```sql
SELECT id, name, description, picture,
       price_usd_currency_code, price_usd_units, price_usd_nanos, categories
FROM products ORDER BY id
```

`categories` is a comma-separated text column; the service lowercases and splits it. The schema and 9 seed products come from the [`product-catalog-db`](https://github.com/platform-boutique/app-product-catalog-db) image.

### Caching & reload

- The catalog is loaded once at startup and cached in memory (per pod).
- **SIGUSR1** arms a one-shot reload: the *next* request re-reads Postgres, then the flag clears. **SIGUSR2** disarms it. In Kubernetes: `kubectl exec deploy/product-catalog-service -- kill -USR1 1`.
- If the cache is ever empty, every request re-queries Postgres.
- Database changes are therefore **not** visible until a reload or a pod restart.
