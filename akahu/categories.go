package akahu

import "context"

// CategoryGroup is a single grouping classifier for a Category.
type CategoryGroup struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// Category is an Akahu transaction category.
type Category struct {
	ID     string                   `json:"_id"`
	Name   string                   `json:"name"`
	Groups map[string]CategoryGroup `json:"groups"`
}

// CategoriesService provides access to /categories.
type CategoriesService struct{ baseService }

// List returns all categories the app has access to.
//
// API: GET /categories
func (s *CategoriesService) List(ctx context.Context) ([]Category, error) {
	var out []Category
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/categories",
		auth:   basicAuth{},
	}, &out)
	return out, err
}

// Get returns a single category by id.
//
// API: GET /categories/{id}
func (s *CategoriesService) Get(ctx context.Context, categoryID string) (*Category, error) {
	var out Category
	err := s.c.doRequest(ctx, apiRequest{
		method: "GET",
		path:   "/categories/" + categoryID,
		auth:   basicAuth{},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
