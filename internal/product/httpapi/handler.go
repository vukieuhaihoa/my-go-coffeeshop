// Package httpapi provides HTTP handlers for the product service.
package httpapi

import (
	"net/http"
	"strings"

	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/entity"
	"github.com/vukieuhaihoa/my-go-coffeeshop/pkg/logger"
)

// Handler is the HTTP handler for the product service.
type Handler struct {
	log      logger.Logger
	items    []entity.ItemType
	itemsMap map[string]entity.ItemType
}

// NewHandler creates a new Handler instance with the provided logger.
// It initializes the item types and creates a map for quick lookup by name.
// Parameters:
//   - log: The logger to be used for logging within the handler.
//
// Returns:
//   - A pointer to a Handler instance.
func NewHandler(log logger.Logger) *Handler {
	itemTypes := []entity.ItemType{
		{
			Name:  "CAPPUCCINO",
			Type:  0,
			Price: 4.5,
			Image: "img/CAPPUCCINO.png",
		},
		{
			Name:  "COFFEE_BLACK",
			Type:  1,
			Price: 3,
			Image: "img/COFFEE_BLACK.png",
		},
		{
			Name:  "COFFEE_WITH_ROOM",
			Type:  2,
			Price: 3,
			Image: "img/COFFEE_WITH_ROOM.png",
		},
		{
			Name:  "ESPRESSO",
			Type:  3,
			Price: 3.5,
			Image: "img/ESPRESSO.png",
		},
		{
			Name:  "ESPRESSO_DOUBLE",
			Type:  4,
			Price: 4.5,
			Image: "img/ESPRESSO_DOUBLE.png",
		},
		{
			Name:  "LATTE",
			Type:  5,
			Price: 4.5,
			Image: "img/LATTE.png",
		},
		{
			Name:  "CAKEPOP",
			Type:  6,
			Price: 2.5,
			Image: "img/CAKEPOP.png",
		},
		{
			Name:  "CROISSANT",
			Type:  7,
			Price: 3.25,
			Image: "img/CROISSANT.png",
		},
		{
			Name:  "MUFFIN",
			Type:  8,
			Price: 3,
			Image: "img/MUFFIN.png",
		},
		{
			Name:  "CROISSANT_CHOCOLATE",
			Type:  9,
			Price: 3.5,
			Image: "img/CROISSANT_CHOCOLATE.png",
		},
	}
	itemsMap := make(map[string]entity.ItemType)
	for _, item := range itemTypes {
		itemsMap[item.Name] = item
	}
	return &Handler{
		log:      log,
		items:    itemTypes,
		itemsMap: itemsMap,
	}
}

// Routes returns the HTTP routes handled by this Handler.
// It sets up the routes for getting item types, getting items by types, and health checks.
// Returns:
//   - An http.Handler that serves the defined routes.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/api/item-types", h.getItemTypes)
	mux.HandleFunc("GET /v1/api/items-by-types/{itemTypes}", h.getItemsByTypes)
	mux.HandleFunc("GET /healthz", h.healthz)

	return mux
}

func (h *Handler) getItemTypes(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, h.items)
}

func (h *Handler) getItemsByTypes(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("itemTypes")

	// Filter the items based on the provided types
	items := []entity.Item{}
	unknownTypes := []string{}

	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		item, exists := h.itemsMap[name]
		if exists {
			items = append(items, entity.Item{
				Type:  item.Type,
				Price: item.Price,
			})
		} else {
			unknownTypes = append(unknownTypes, name)
		}
	}

	if len(unknownTypes) > 0 {
		h.writeError(w, http.StatusBadRequest, "Unknown item types: "+strings.Join(unknownTypes, ", "))
		return
	}

	if len(items) == 0 {
		h.writeError(w, http.StatusBadRequest, "No valid item types provided")
		return
	}

	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
