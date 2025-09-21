package com.terraformingmars.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// Enums matching the Go DTOs with lowercase serialization
@Serializable
enum class GamePhase {
    @SerialName("waiting_for_game_start")
    WAITING_FOR_GAME_START,
    @SerialName("starting_card_selection")
    STARTING_CARD_SELECTION,
    @SerialName("start_game_selection")
    START_GAME_SELECTION,
    @SerialName("action")
    ACTION,
    @SerialName("production_and_card_draw")
    PRODUCTION_AND_CARD_DRAW,
    @SerialName("complete")
    COMPLETE
}

@Serializable
enum class GameStatus {
    @SerialName("lobby")
    LOBBY,
    @SerialName("active")
    ACTIVE,
    @SerialName("completed")
    COMPLETED
}

@Serializable
enum class CardType {
    @SerialName("automated")
    AUTOMATED,
    @SerialName("active")
    ACTIVE,
    @SerialName("event")
    EVENT,
    @SerialName("corporation")
    CORPORATION,
    @SerialName("prelude")
    PRELUDE
}

@Serializable
enum class CardTag {
    @SerialName("space")
    SPACE,
    @SerialName("earth")
    EARTH,
    @SerialName("science")
    SCIENCE,
    @SerialName("power")
    POWER,
    @SerialName("building")
    BUILDING,
    @SerialName("microbe")
    MICROBE,
    @SerialName("animal")
    ANIMAL,
    @SerialName("plant")
    PLANT,
    @SerialName("event")
    EVENT,
    @SerialName("city")
    CITY,
    @SerialName("venus")
    VENUS,
    @SerialName("jovian")
    JOVIAN,
    @SerialName("wildlife")
    WILDLIFE,
    @SerialName("wild")
    WILD
}

// Data classes
@Serializable
data class ResourceSet(
    val credits: Int = 0,
    val steel: Int = 0,
    val titanium: Int = 0,
    val plants: Int = 0,
    val energy: Int = 0,
    val heat: Int = 0
)

@Serializable
data class CardRequirement(
    val type: String,
    val min: Int? = null,
    val max: Int? = null,
    val resource: String? = null
)

@Serializable
data class CardTrigger(
    val type: String
)

@Serializable
data class CardOutput(
    val type: String,
    val amount: Int? = null,
    val player: String? = null
)

@Serializable
data class CardBehaviorDto(
    val triggers: List<CardTrigger>? = null,
    val inputs: List<CardOutput>? = null,
    val outputs: List<CardOutput>? = null,
    val choices: List<CardChoice>? = null
)

@Serializable
data class CardChoice(
    val inputs: List<CardOutput>? = null,
    val outputs: List<CardOutput>? = null
)

@Serializable
data class VictoryPointCondition(
    val amount: Int,
    val condition: String,
    val maxTrigger: Int? = null,
    val per: VictoryPointPer? = null
)

@Serializable
data class VictoryPointPer(
    val type: String,
    val amount: Int? = null,
    val location: String? = null
)

@Serializable
data class ResourceStorageDto(
    val type: String,
    val capacity: Int? = null,
    val starting: Int
)

@Serializable
data class CardDto(
    val id: String,
    val name: String,
    val type: CardType,
    val cost: Int,
    val description: String,
    val tags: List<CardTag>? = null,
    val requirements: List<CardRequirement>? = null,
    val behaviors: List<CardBehaviorDto>? = null,
    val resourceStorage: ResourceStorageDto? = null,
    val vpConditions: List<VictoryPointCondition>? = null
)

@Serializable
data class CorporationDto(
    val id: String,
    val name: String,
    val description: String,
    val startingCredits: Int,
    val startingResources: ResourceSet,
    val startingProduction: ResourceSet,
    val tags: List<CardTag>,
    val specialEffects: List<String>,
    val number: String
)

@Serializable
data class ProductionPhaseDto(
    val availableCards: List<CardDto>,
    val selectionComplete: Boolean
)

@Serializable
data class GameSettingsDto(
    val maxPlayers: Int
)

@Serializable
data class GlobalParametersDto(
    val temperature: Int, // Range: -30 to +8°C
    val oxygen: Int,      // Range: 0-14%
    val oceans: Int       // Range: 0-9
)

@Serializable
data class ResourcesDto(
    val credits: Int,
    val steel: Int,
    val titanium: Int,
    val plants: Int,
    val energy: Int,
    val heat: Int
)

@Serializable
data class ResourceDelta(
    val credits: Int,
    val steel: Int,
    val titanium: Int,
    val plants: Int,
    val energy: Int,
    val heat: Int
)

@Serializable
data class ProductionDto(
    val credits: Int,
    val steel: Int,
    val titanium: Int,
    val plants: Int,
    val energy: Int,
    val heat: Int
)

@Serializable
data class PlayerDto(
    val id: String,
    val name: String,
    val corporation: String? = null,
    val cards: List<CardDto>,
    val resources: ResourcesDto,
    val resourceProduction: ProductionDto,
    val terraformRating: Int,
    val playedCards: List<String>,
    val passed: Boolean,
    val availableActions: Int,
    val victoryPoints: Int,
    val isConnected: Boolean,
    // Card selection state - nullable, exists only during selection phase
    val productionSelection: ProductionPhaseDto? = null,
    // Starting card selection - available during starting_card_selection phase
    val startingSelection: List<CardDto>? = null
)

@Serializable
data class OtherPlayerDto(
    val id: String,
    val name: String,
    val corporation: String,
    val handCardCount: Int, // Number of cards in hand (private)
    val resources: ResourcesDto,
    val resourceProduction: ProductionDto,
    val terraformRating: Int,
    val playedCards: List<String>, // Played cards are public
    val passed: Boolean,
    val availableActions: Int,
    val victoryPoints: Int,
    val isConnected: Boolean,
    // Card selection state - limited visibility for other players
    val isSelectingCards: Boolean // Whether player is currently selecting cards
)

@Serializable
data class GameDto(
    val id: String,
    val status: GameStatus,
    val settings: GameSettingsDto,
    val hostPlayerId: String,
    val currentPhase: GamePhase,
    val globalParameters: GlobalParametersDto,
    val currentPlayer: PlayerDto,       // Viewing player's full data
    val otherPlayers: List<OtherPlayerDto>, // Other players' limited data
    val viewingPlayerId: String,        // The player viewing this game state
    val currentTurn: String? = null,    // Whose turn it is (nullable)
    val generation: Int,
    val remainingActions: Int, // Remaining actions in the current turn
    val turnOrder: List<String> // Turn order of all players in game
)