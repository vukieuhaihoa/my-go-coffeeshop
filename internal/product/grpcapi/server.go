// Package grpcapi implements the gRPC API for the product service.
package grpcapi

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	productv1 "github.com/vukieuhaihoa/my-go-coffeeshop/gen/go/coffeeshop/product/v1"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/catalog"
	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/entity"
)

// Server implements the gRPC server for the product service.
// It provides methods to retrieve item types and items by type from the catalog.
type Server struct {
	productv1.UnimplementedProductServiceServer
	catalog *catalog.Catalog
}

var _ productv1.ProductServiceServer = (*Server)(nil)

// NewServer creates a new instance of the gRPC server for the product service.
// It initializes the server with the provided catalog.
// Parameters:
//   - catalog: The catalog instance used to retrieve item types and items.
//
// Returns:
//   - A pointer to a Server instance.
func NewServer(catalog *catalog.Catalog) *Server {
	return &Server{
		catalog: catalog,
	}
}

// GetItemTypes retrieves all item types from the catalog and returns them in a GetItemTypesResponse.
// Parameters:
//   - ctx: The context for managing the request's lifecycle.
//   - req: The GetItemTypesRequest containing the request parameters.
//
// Returns:
//   - A pointer to a GetItemTypesResponse containing the item types.
//   - An error if there was an issue retrieving the item types.
func (s *Server) GetItemTypes(_ context.Context, _ *productv1.GetItemTypesRequest) (*productv1.GetItemTypesResponse, error) {
	itemTypes := s.catalog.All()

	res := &productv1.GetItemTypesResponse{
		ItemTypes: make([]*productv1.ItemType, len(itemTypes)),
	}

	for i, item := range itemTypes {
		res.ItemTypes[i] = toProtoItemType(item)
	}

	return res, nil
}

// GetItemsByType retrieves items from the catalog based on the provided item types in the request.
// It returns a GetItemsByTypeResponse containing the items and handles any unknown item types.
// Parameters:
//   - ctx: The context for managing the request's lifecycle.
//   - req: The GetItemsByTypeRequest containing the item types to retrieve.
//
// Returns:
//   - A pointer to a GetItemsByTypeResponse containing the items.
//   - An error if there was an issue retrieving the items or if unknown item types were provided.
func (s *Server) GetItemsByType(_ context.Context, req *productv1.GetItemsByTypeRequest) (*productv1.GetItemsByTypeResponse, error) {
	raw := req.GetItemTypes()

	names := []string{}

	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		names = append(names, name)
	}

	found, unknown := s.catalog.GetByNames(names)
	if len(unknown) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "unknown item types: [%s]", strings.Join(unknown, ", "))
	}

	if len(found) == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_types is required")
	}

	res := &productv1.GetItemsByTypeResponse{
		Items: make([]*productv1.Item, len(found)),
	}

	for i, item := range found {
		res.Items[i] = toProtoItem(item)
	}

	return res, nil
}

func toProtoItemType(item entity.ItemType) *productv1.ItemType {
	return &productv1.ItemType{
		Name:  item.Name,
		Type:  item.Type,
		Price: item.Price,
		Image: item.Image,
	}
}

func toProtoItem(item entity.ItemType) *productv1.Item {
	return &productv1.Item{
		Type:  item.Type,
		Price: item.Price,
	}
}
