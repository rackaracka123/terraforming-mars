package com.terraformingmars.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
sealed class WebSocketCommand {
    abstract val gameId: String?
    abstract val playerId: String?
}

// Client -> Server Commands

@Serializable
@SerialName("player-connect")
data class PlayerConnectCommand(
    val playerName: String,
    override val gameId: String,
    override val playerId: String? = null // For reconnection
) : WebSocketCommand()

@Serializable
@SerialName("action.game-management.start-game")
data class StartGameCommand(
    override val gameId: String,
    override val playerId: String
) : WebSocketCommand()

@Serializable
@SerialName("action.game-management.skip-action")
data class SkipActionCommand(
    override val gameId: String,
    override val playerId: String
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.sell-patents")
data class SellPatentsCommand(
    override val gameId: String,
    override val playerId: String,
    val cardCount: Int
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.build-power-plant")
data class BuildPowerPlantCommand(
    override val gameId: String,
    override val playerId: String
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.launch-asteroid")
data class LaunchAsteroidCommand(
    override val gameId: String,
    override val playerId: String
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.build-aquifer")
data class BuildAquiferCommand(
    override val gameId: String,
    override val playerId: String,
    val hexPosition: HexPositionDto
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.plant-greenery")
data class PlantGreeneryCommand(
    override val gameId: String,
    override val playerId: String,
    val hexPosition: HexPositionDto
) : WebSocketCommand()

@Serializable
@SerialName("action.standard-project.build-city")
data class BuildCityCommand(
    override val gameId: String,
    override val playerId: String,
    val hexPosition: HexPositionDto
) : WebSocketCommand()

@Serializable
@SerialName("action.card.play-card")
data class PlayCardCommand(
    override val gameId: String,
    override val playerId: String,
    val cardId: String
) : WebSocketCommand()

@Serializable
@SerialName("action.card.select-starting-card")
data class SelectStartingCardCommand(
    override val gameId: String,
    override val playerId: String,
    val cardIds: List<String>
) : WebSocketCommand()

@Serializable
@SerialName("action.card.select-cards")
data class SelectCardsCommand(
    override val gameId: String,
    override val playerId: String,
    val cardIds: List<String>
) : WebSocketCommand()

// Server -> Client Events

@Serializable
sealed class WebSocketEvent

@Serializable
@SerialName("player-connected")
data class PlayerConnectedEvent(
    val playerId: String,
    val playerName: String,
    val game: GameDto
) : WebSocketEvent()

@Serializable
@SerialName("player-reconnected")
data class PlayerReconnectedEvent(
    val playerId: String,
    val playerName: String,
    val game: GameDto
) : WebSocketEvent()

@Serializable
@SerialName("player-disconnected")
data class PlayerDisconnectedEvent(
    val playerId: String,
    val playerName: String,
    val game: GameDto
) : WebSocketEvent()

@Serializable
@SerialName("game-updated")
data class GameUpdatedEvent(
    val game: GameDto
) : WebSocketEvent()

@Serializable
@SerialName("full-state")
data class FullStateEvent(
    val game: GameDto,
    val playerId: String
) : WebSocketEvent()

@Serializable
@SerialName("production-phase-started")
data class ProductionPhaseStartedEvent(
    val generation: Int,
    val playersData: List<PlayerProductionData>,
    val game: GameDto
) : WebSocketEvent()

@Serializable
@SerialName("error")
data class ErrorEvent(
    val message: String,
    val code: String? = null
) : WebSocketEvent()