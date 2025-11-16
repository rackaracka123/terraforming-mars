package http

import (
	"net/http"
	"strconv"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// CardHandler handles card-related HTTP requests
type CardHandler struct {
	*BaseHandler
	cardRepo card.CardRepository
}

// NewCardHandler creates a new CardHandler
func NewCardHandler(cardRepo card.CardRepository) *CardHandler {
	return &CardHandler{
		BaseHandler: &BaseHandler{},
		cardRepo:    cardRepo,
	}
}

// ListCards handles GET /api/v1/cards
// @Summary List cards with pagination
// @Description List all cards with pagination
// @Tags cards
// @Accept json
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(50)
// @Success 200 {object} dto.ListCardsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cards [get]
func (h *CardHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	log.Debug("📡 Getting cards with pagination")

	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 50

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else {
			log.Warn("Invalid offset parameter", zap.String("offset", offsetStr))
			h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 10000 {
			limit = parsedLimit
		} else {
			log.Warn("Invalid limit parameter", zap.String("limit", limitStr))
			h.WriteErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter (must be 1-10000)")
			return
		}
	}

	// List cards from repository
	allCards, err := h.cardRepo.GetAllCards(r.Context())
	if err != nil {
		log.Error("Failed to get cards", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get cards")
		return
	}

	// Manual pagination
	totalCount := len(allCards)
	start := offset
	end := offset + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}
	cards := allCards[start:end]

	// Convert to DTOs
	cardDtos := make([]dto.CardDto, len(cards))
	for i, card := range cards {
		cardDtos[i] = dto.ToCardDto(card)
	}

	// Create response
	response := dto.ListCardsResponse{
		Cards:      cardDtos,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}

	log.Debug("📡 Cards retrieved successfully",
		zap.Int("count", len(cardDtos)),
		zap.Int("total", totalCount),
		zap.Int("offset", offset),
		zap.Int("limit", limit))

	h.WriteJSONResponse(w, http.StatusOK, response)
}

// GetCorporations handles GET /api/v1/corporations
// @Summary List all corporations
// @Description List all available corporations with their starting bonuses
// @Tags cards
// @Accept json
// @Produce json
// @Success 200 {array} dto.CardDto
// @Failure 500 {object} dto.ErrorResponse
// @Router /corporations [get]
func (h *CardHandler) GetCorporations(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()
	log.Debug("📡 Getting all corporations")

	// Get corporations from service (they're just cards with type=corporation)
	corporations, err := h.cardRepo.GetCorporations(r.Context())
	if err != nil {
		log.Error("Failed to get corporations", zap.Error(err))
		h.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get corporations")
		return
	}

	// Convert to DTOs (corporations have StartingCredits/Resources/Production populated)
	corporationDtos := dto.ToCardDtoSlice(corporations)

	log.Debug("📡 Corporations retrieved successfully",
		zap.Int("count", len(corporationDtos)))

	h.WriteJSONResponse(w, http.StatusOK, corporationDtos)
}
