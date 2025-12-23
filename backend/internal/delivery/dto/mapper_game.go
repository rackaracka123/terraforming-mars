package dto

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
)

// ToGameDto converts migration Game to GameDto with personalized view
// The playerID parameter determines which player is "currentPlayer" vs "otherPlayers"
func ToGameDto(g *game.Game, cardRegistry cards.CardRegistry, playerID string) GameDto {
	players := g.GetAllPlayers()

	// Create personalized view: viewing player is currentPlayer, others are otherPlayers
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	var viewingPlayer *player.Player
	for _, p := range players {
		if p.ID() == playerID {
			viewingPlayer = p
			currentPlayer = ToPlayerDto(p, g, cardRegistry)
		} else {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(p, g, cardRegistry))
		}
	}

	// If viewing player not found, use first player as fallback
	if viewingPlayer == nil && len(players) > 0 {
		currentPlayer = ToPlayerDto(players[0], g, cardRegistry)
		for i := 1; i < len(players); i++ {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(players[i], g, cardRegistry))
		}
		playerID = players[0].ID()
	}

	// Get settings
	settings := g.Settings()
	settingsDto := GameSettingsDto{
		MaxPlayers:      settings.MaxPlayers,
		DevelopmentMode: settings.DevelopmentMode,
		DemoGame:        settings.DemoGame,
		CardPacks:       settings.CardPacks,
	}

	// Convert global parameters
	globalParams := g.GlobalParameters()
	globalParamsDto := GlobalParametersDto{
		Temperature: globalParams.Temperature(),
		Oxygen:      globalParams.Oxygen(),
		Oceans:      globalParams.Oceans(),
	}

	// Get board tiles
	board := g.Board()
	tiles := board.Tiles()
	tileDtos := make([]TileDto, len(tiles))
	for i, tile := range tiles {
		tileDtos[i] = TileDto{
			Coordinates: HexPositionDto{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Type:        string(tile.Type),
			OwnerID:     tile.OwnerID,
			Tags:        tile.Tags,
			Bonuses:     convertTileBonuses(tile.Bonuses),
			Location:    string(tile.Location),
			DisplayName: tile.DisplayName,
		}
		if tile.OccupiedBy != nil {
			occupant := &TileOccupantDto{
				Type: string(tile.OccupiedBy.Type),
				Tags: tile.OccupiedBy.Tags,
			}
			tileDtos[i].OccupiedBy = occupant
		}
	}

	// Payment constants (default values)
	paymentConstants := PaymentConstantsDto{
		SteelValue:    2, // Default steel value
		TitaniumValue: 3, // Default titanium value
	}

	// Convert final scores if game is completed
	var finalScoreDtos []FinalScoreDto
	if g.Status() == game.GameStatusCompleted {
		finalScores := g.GetFinalScores()
		if finalScores != nil {
			finalScoreDtos = make([]FinalScoreDto, len(finalScores))
			for i, fs := range finalScores {
				finalScoreDtos[i] = FinalScoreDto{
					PlayerID:    fs.PlayerID,
					PlayerName:  fs.PlayerName,
					VPBreakdown: ToVPBreakdownDto(fs.Breakdown),
					IsWinner:    fs.IsWinner,
					Placement:   fs.Placement,
				}
			}
		}
	}

	return GameDto{
		ID:               g.ID(),
		Status:           GameStatus(g.Status()),
		Settings:         settingsDto,
		HostPlayerID:     g.HostPlayerID(),
		CurrentPhase:     GamePhase(g.CurrentPhase()),
		GlobalParameters: globalParamsDto,
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  playerID, // The player viewing this game state
		CurrentTurn:      getCurrentTurnPlayerID(g),
		Generation:       g.Generation(),
		TurnOrder:        g.TurnOrder(),
		Board: BoardDto{
			Tiles: tileDtos,
		},
		PaymentConstants: paymentConstants,
		Milestones:       ToMilestonesDto(g.Milestones()),
		Awards:           ToAwardsDto(g.Awards()),
		FinalScores:      finalScoreDtos,
	}
}

// getCurrentTurnPlayerID extracts the player ID from the current turn
func getCurrentTurnPlayerID(g *game.Game) *string {
	turn := g.CurrentTurn()
	if turn == nil {
		return nil
	}
	playerID := turn.PlayerID()
	return &playerID
}

// convertTileBonuses converts migration TileBonus to DTO
func convertTileBonuses(bonuses []board.TileBonus) []TileBonusDto {
	dtos := make([]TileBonusDto, len(bonuses))
	for i, bonus := range bonuses {
		dtos[i] = TileBonusDto{
			Type:   string(bonus.Type),
			Amount: bonus.Amount,
		}
	}
	return dtos
}

// ToMilestonesDto converts all milestones to DTOs including claim status
func ToMilestonesDto(milestones *game.Milestones) []MilestoneDto {
	dtos := make([]MilestoneDto, len(game.AllMilestones))
	for i, info := range game.AllMilestones {
		var claimedBy *string
		isClaimed := milestones.IsClaimed(info.Type)
		if isClaimed {
			// Find who claimed it
			for _, claimed := range milestones.ClaimedMilestones() {
				if claimed.Type == info.Type {
					claimedBy = &claimed.PlayerID
					break
				}
			}
		}
		dtos[i] = MilestoneDto{
			Type:        string(info.Type),
			Name:        info.Name,
			Description: info.Description,
			IsClaimed:   isClaimed,
			ClaimedBy:   claimedBy,
			ClaimCost:   game.MilestoneClaimCost,
		}
	}
	return dtos
}

// ToAwardsDto converts all awards to DTOs including funding status
func ToAwardsDto(awards *game.Awards) []AwardDto {
	dtos := make([]AwardDto, len(game.AllAwards))
	fundedCount := awards.FundedCount()

	for i, info := range game.AllAwards {
		var fundedBy *string
		isFunded := awards.IsFunded(info.Type)
		fundingCost := game.AwardFundingCosts[0] // Default cost for first award

		if isFunded {
			// Find who funded it and what the cost was
			for _, funded := range awards.FundedAwards() {
				if funded.Type == info.Type {
					fundedBy = &funded.FundedByPlayer
					fundingCost = funded.FundingCost
					break
				}
			}
		} else {
			// Calculate cost for next award
			if fundedCount < game.MaxFundedAwards {
				fundingCost = game.AwardFundingCosts[fundedCount]
			}
		}

		dtos[i] = AwardDto{
			Type:        string(info.Type),
			Name:        info.Name,
			Description: info.Description,
			IsFunded:    isFunded,
			FundedBy:    fundedBy,
			FundingCost: fundingCost,
		}
	}
	return dtos
}

// ToCardVPConditionDetailDto converts a card VP condition detail to DTO
func ToCardVPConditionDetailDto(detail game.CardVPConditionDetail) CardVPConditionDetailDto {
	return CardVPConditionDetailDto{
		ConditionType:  detail.ConditionType,
		Amount:         detail.Amount,
		Count:          detail.Count,
		MaxTrigger:     detail.MaxTrigger,
		ActualTriggers: detail.ActualTriggers,
		TotalVP:        detail.TotalVP,
		Explanation:    detail.Explanation,
	}
}

// ToCardVPDetailDto converts a card VP detail to DTO
func ToCardVPDetailDto(detail game.CardVPDetail) CardVPDetailDto {
	conditions := make([]CardVPConditionDetailDto, len(detail.Conditions))
	for i, cond := range detail.Conditions {
		conditions[i] = ToCardVPConditionDetailDto(cond)
	}
	return CardVPDetailDto{
		CardID:     detail.CardID,
		CardName:   detail.CardName,
		Conditions: conditions,
		TotalVP:    detail.TotalVP,
	}
}

// ToGreeneryVPDetailDto converts a greenery VP detail to DTO
func ToGreeneryVPDetailDto(detail game.GreeneryVPDetail) GreeneryVPDetailDto {
	return GreeneryVPDetailDto{
		Coordinate: detail.Coordinate,
		VP:         detail.VP,
	}
}

// ToCityVPDetailDto converts a city VP detail to DTO
func ToCityVPDetailDto(detail game.CityVPDetail) CityVPDetailDto {
	return CityVPDetailDto{
		CityCoordinate:     detail.CityCoordinate,
		AdjacentGreeneries: detail.AdjacentGreeneries,
		VP:                 detail.VP,
	}
}

// ToVPBreakdownDto converts a VP breakdown to DTO
func ToVPBreakdownDto(breakdown game.VPBreakdown) VPBreakdownDto {
	cardVPDetails := make([]CardVPDetailDto, len(breakdown.CardVPDetails))
	for i, detail := range breakdown.CardVPDetails {
		cardVPDetails[i] = ToCardVPDetailDto(detail)
	}

	greeneryVPDetails := make([]GreeneryVPDetailDto, len(breakdown.GreeneryVPDetails))
	for i, detail := range breakdown.GreeneryVPDetails {
		greeneryVPDetails[i] = ToGreeneryVPDetailDto(detail)
	}

	cityVPDetails := make([]CityVPDetailDto, len(breakdown.CityVPDetails))
	for i, detail := range breakdown.CityVPDetails {
		cityVPDetails[i] = ToCityVPDetailDto(detail)
	}

	return VPBreakdownDto{
		TerraformRating:   breakdown.TerraformRating,
		CardVP:            breakdown.CardVP,
		CardVPDetails:     cardVPDetails,
		MilestoneVP:       breakdown.MilestoneVP,
		AwardVP:           breakdown.AwardVP,
		GreeneryVP:        breakdown.GreeneryVP,
		GreeneryVPDetails: greeneryVPDetails,
		CityVP:            breakdown.CityVP,
		CityVPDetails:     cityVPDetails,
		TotalVP:           breakdown.TotalVP,
	}
}

// ToFinalScoreDto creates a final score DTO for a player
func ToFinalScoreDto(playerID, playerName string, breakdown game.VPBreakdown, isWinner bool, placement int) FinalScoreDto {
	return FinalScoreDto{
		PlayerID:    playerID,
		PlayerName:  playerName,
		VPBreakdown: ToVPBreakdownDto(breakdown),
		IsWinner:    isWinner,
		Placement:   placement,
	}
}
