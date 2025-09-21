package com.terraformingmars.models

import com.terraformingmars.dto.*

// Domain models for internal use
data class GameSettings(
    val maxPlayers: Int
)

data class GlobalParameters(
    val temperature: Int, // Range: -30 to +8°C
    val oxygen: Int,      // Range: 0-14%
    val oceans: Int       // Range: 0-9
)

data class Resources(
    val credits: Int,
    val steel: Int,
    val titanium: Int,
    val plants: Int,
    val energy: Int,
    val heat: Int
)

data class Production(
    val credits: Int,
    val steel: Int,
    val titanium: Int,
    val plants: Int,
    val energy: Int,
    val heat: Int
)

data class ProductionPhase(
    val availableCards: List<Card>,
    val selectionComplete: Boolean
)

data class Game(
    val id: String,
    val status: GameStatus,
    val settings: GameSettings,
    val hostPlayerId: String?,
    val currentPhase: GamePhase,
    val globalParameters: GlobalParameters,
    val players: List<Player>,
    val currentTurn: String?,
    val generation: Int,
    val remainingActions: Int,
    val turnOrder: List<String>
) {
    fun toDto(viewingPlayerId: String): GameDto {
        val viewingPlayer = players.find { it.id == viewingPlayerId }
            ?: throw IllegalArgumentException("Viewing player not found in game")

        val otherPlayers = players.filter { it.id != viewingPlayerId }
            .map { it.toOtherPlayerDto() }

        return GameDto(
            id = id,
            status = status,
            settings = GameSettingsDto(maxPlayers = settings.maxPlayers),
            hostPlayerId = hostPlayerId ?: "",
            currentPhase = currentPhase,
            globalParameters = GlobalParametersDto(
                temperature = globalParameters.temperature,
                oxygen = globalParameters.oxygen,
                oceans = globalParameters.oceans
            ),
            currentPlayer = viewingPlayer.toDto(),
            otherPlayers = otherPlayers,
            viewingPlayerId = viewingPlayerId,
            currentTurn = currentTurn,
            generation = generation,
            remainingActions = remainingActions,
            turnOrder = turnOrder
        )
    }
}

data class Player(
    val id: String,
    val name: String,
    val corporation: String?,
    val cards: List<Card>,
    val resources: Resources,
    val resourceProduction: Production,
    val terraformRating: Int,
    val playedCards: List<String>,
    val passed: Boolean,
    val availableActions: Int,
    val victoryPoints: Int,
    var isConnected: Boolean,
    val productionSelection: ProductionPhase?,
    val startingSelection: List<Card>?
) {
    fun toDto(): PlayerDto = PlayerDto(
        id = id,
        name = name,
        corporation = corporation,
        cards = cards.map { it.toDto() },
        resources = ResourcesDto(
            credits = resources.credits,
            steel = resources.steel,
            titanium = resources.titanium,
            plants = resources.plants,
            energy = resources.energy,
            heat = resources.heat
        ),
        resourceProduction = ProductionDto(
            credits = resourceProduction.credits,
            steel = resourceProduction.steel,
            titanium = resourceProduction.titanium,
            plants = resourceProduction.plants,
            energy = resourceProduction.energy,
            heat = resourceProduction.heat
        ),
        terraformRating = terraformRating,
        playedCards = playedCards,
        passed = passed,
        availableActions = availableActions,
        victoryPoints = victoryPoints,
        isConnected = isConnected,
        productionSelection = productionSelection?.let { phase ->
            ProductionPhaseDto(
                availableCards = phase.availableCards.map { it.toDto() },
                selectionComplete = phase.selectionComplete
            )
        },
        startingSelection = startingSelection?.map { it.toDto() }
    )

    fun toOtherPlayerDto(): OtherPlayerDto = OtherPlayerDto(
        id = id,
        name = name,
        corporation = corporation ?: "",
        handCardCount = cards.size,
        resources = ResourcesDto(
            credits = resources.credits,
            steel = resources.steel,
            titanium = resources.titanium,
            plants = resources.plants,
            energy = resources.energy,
            heat = resources.heat
        ),
        resourceProduction = ProductionDto(
            credits = resourceProduction.credits,
            steel = resourceProduction.steel,
            titanium = resourceProduction.titanium,
            plants = resourceProduction.plants,
            energy = resourceProduction.energy,
            heat = resourceProduction.heat
        ),
        terraformRating = terraformRating,
        playedCards = playedCards,
        passed = passed,
        availableActions = availableActions,
        victoryPoints = victoryPoints,
        isConnected = isConnected,
        isSelectingCards = productionSelection != null
    )
}

data class Card(
    val id: String,
    val name: String,
    val type: CardType,
    val cost: Int,
    val description: String,
    val tags: List<CardTag>?,
    val requirements: List<CardRequirement>?,
    val behaviors: List<CardBehaviorDto>?,
    val resourceStorage: ResourceStorageDto?,
    val vpConditions: List<VictoryPointCondition>?
) {
    fun toDto(): CardDto = CardDto(
        id = id,
        name = name,
        type = type,
        cost = cost,
        description = description,
        tags = tags,
        requirements = requirements,
        behaviors = behaviors,
        resourceStorage = resourceStorage,
        vpConditions = vpConditions
    )
}