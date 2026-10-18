package products

import (
	"context"

	shopqua "shop-qua"
	"shop-qua/db"
)

// The catalogue.

var products = db.Model[Product](modelProducts)

// listProducts gives active products to everyone; a manager may see the inactive ones too.
func listProducts(ctx context.Context, search string, limit, offset int, everything bool) ([]Product, error) {
	where := shopqua.Where{}
	if !everything {
		where["active"] = true
	}
	if search != "" {
		where["name"] = map[string]any{"like": "%" + search + "%"}
	}
	if limit == 0 {
		limit = 50
	}
	return products.Find(ctx, db.Query{Where: where, OrderBy: map[string]string{"name": "asc"}, Limit: limit, Offset: offset})
}

func createProduct(ctx context.Context, in newProduct) (Product, error) {
	if existing, err := products.FindOne(ctx, shopqua.Where{"name": in.Name}); err != nil || existing != nil {
		if err != nil {
			return Product{}, err
		}
		return Product{}, &appError{409, "name_taken", "A product with this name exists"}
	}
	stock := 0
	if in.Stock != nil {
		stock = *in.Stock
	}
	return products.Create(ctx, map[string]any{"name": in.Name, "description": in.Description, "price": in.Price, "stock": stock, "active": true})
}

func updateProduct(ctx context.Context, id string, changes map[string]any) (Product, error) {
	p, err := products.UpdateByID(ctx, id, changes)
	if err != nil {
		return Product{}, err
	}
	if p == nil {
		return Product{}, &appError{404, "product_not_found", "No such product"}
	}
	return *p, nil
}

func restock(ctx context.Context, id string, quantity int) (Product, error) {
	return updateProduct(ctx, id, map[string]any{"stock": db.Increment(float64(quantity))})
}
