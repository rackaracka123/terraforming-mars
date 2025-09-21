package com.terraformingmars.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// WebSocket Message Types matching the Go DTOs
@Serializable
enum class MessageType {
    // Existing Client -> Server messages
    @SerialName("player-connect")
    PLAYER_CONNECT,

    // Existing Server -> Client messages
    @SerialName("game-updated")
    GAME_UPDATED,
    @SerialName("player-connected")
    PLAYER_CONNECTED,
    @SerialName("player-reconnected")
    PLAYER_RECONNECTED,
    @SerialName("player-disconnected")
    PLAYER_DISCONNECTED,
    @SerialName("error")
    ERROR,
    @SerialName("full-state")
    FULL_STATE,
    @SerialName("production-phase-started")
    PRODUCTION_PHASE_STARTED,

    // Standard project message types
    @SerialName("action.standard-project.sell-patents")
    ACTION_SELL_PATENTS,
    @SerialName("action.standard-project.launch-asteroid")
    ACTION_LAUNCH_ASTEROID,
    @SerialName("action.standard-project.build-power-plant")
    ACTION_BUILD_POWER_PLANT,
    @SerialName("action.standard-project.build-aquifer")
    ACTION_BUILD_AQUIFER,
    @SerialName("action.standard-project.plant-greenery")
    ACTION_PLANT_GREENERY,
    @SerialName("action.standard-project.build-city")
    ACTION_BUILD_CITY,

    // Game management message types
    @SerialName("action.game-management.start-game")
    ACTION_START_GAME,
    @SerialName("action.game-management.skip-action")
    ACTION_SKIP_ACTION,

    // Card message types
    @SerialName("action.card.play-card")
    ACTION_PLAY_CARD,
    @SerialName("action.card.select-starting-card")
    ACTION_SELECT_STARTING_CARD,
    @SerialName("action.card.select-cards")
    ACTION_SELECT_CARDS
}

@Serializable
enum class ActionType {
    @SerialName("select-starting-card")
    SELECT_STARTING_CARD,
    @SerialName("select-cards")
    SELECT_CARDS,
    @SerialName("start-game")
    START_GAME,
    @SerialName("skip-action")
    SKIP_ACTION,
    @SerialName("play-card")
    PLAY_CARD,
    // Standard Projects
    @SerialName("sell-patents")
    SELL_PATENTS,
    @SerialName("build-power-plant")
    BUILD_POWER_PLANT,
    @SerialName("launch-asteroid")
    LAUNCH_ASTEROID,
    @SerialName("build-aquifer")
    BUILD_AQUIFER,
    @SerialName("plant-greenery")
    PLANT_GREENERY,
    @SerialName("build-city")
    BUILD_CITY
}

// Base WebSocket Message
@Serializable
data class WebSocketMessage(
    val type: MessageType,
    val payload: String, // JSON string for flexible payload handling
    val gameId: String? = null
)

// Hex position for tile placement
@Serializable
data class HexPositionDto(
    val q: Int,
    val r: Int,
    val s: Int
)

// Payloads for different message types

@Serializable
data class PlayerConnectPayload(
    val playerName: String,
    val gameId: String,
    val playerId: String? = null // Optional: used for reconnection
)

@Serializable
data class GameUpdatedPayload(
    val game: GameDto
)

@Serializable
data class PlayerConnectedPayload(
    val playerId: String,
    val playerName: String,
    val game: GameDto
)

@Serializable
data class ErrorPayload(
    val message: String,
    val code: String? = null
)

@Serializable
data class FullStatePayload(
    val game: GameDto,
    val playerId: String
)

@Serializable
data class PlayerReconnectedPayload(
    val playerId: String,
    val playerName: String,
    val game: GameDto
)

@Serializable
data class PlayerDisconnectedPayload(
    val playerId: String,
    val playerName: String,
    val game: GameDto
)

@Serializable
data class PlayerProductionData(
    val playerId: String,
    val playerName: String,
    val beforeResources: ResourcesDto,
    val afterResources: ResourcesDto,
    val resourceDelta: ResourceDelta,
    val production: ProductionDto,
    val terraformRating: Int,
    val energyConverted: Int,
    val creditsIncome: Int
)

@Serializable
data class ProductionPhaseStartedPayload(
    val generation: Int,
    val playersData: List<PlayerProductionData>,
    val game: GameDto
)

// Action payloads

@Serializable
data class SelectStartingCardAction(
    val type: ActionType,
    val cardIds: List<String>
)

@Serializable
data class StartGameAction(
    val type: ActionType
)

@Serializable
data class SkipAction(
    val type: ActionType
)

@Serializable
data class PlayCardAction(
    val cardId: String
)

@Serializable
data class SellPatentsAction(
    val type: ActionType,
    val cardCount: Int
)

@Serializable
data class BuildPowerPlantAction(
    val type: ActionType
)

@Serializable
data class LaunchAsteroidAction(
    val type: ActionType
)

@Serializable
data class BuildAquiferAction(
    val type: ActionType,
    val hexPosition: HexPositionDto
)

@Serializable
data class PlantGreeneryAction(
    val type: ActionType,
    val hexPosition: HexPositionDto
)

@Serializable
data class BuildCityAction(
    val type: ActionType,
    val hexPosition: HexPositionDto
)