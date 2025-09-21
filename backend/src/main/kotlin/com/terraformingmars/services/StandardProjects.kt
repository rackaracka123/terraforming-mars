package com.terraformingmars.services

import com.terraformingmars.dto.*
import com.terraformingmars.models.Game
import com.terraformingmars.models.Player
import com.terraformingmars.repositories.GameRepository

data class ResourceCost(
    val credits: Int = 0,
    val plants: Int = 0,
    val heat: Int = 0
)

data class GlobalParameterChange(
    val temperature: Int = 0,
    val oxygen: Int = 0,
    val oceans: Int = 0
)

data class ProductionChange(
    val credits: Int = 0,
    val energy: Int = 0
)

class StandardProjects(
    private val gameRepository: GameRepository,
    private val eventNotifier: GameEventNotifier
) {

    suspend fun buildPowerPlant(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(credits = 11),
            productionChange = ProductionChange(energy = 1)
        )

    suspend fun launchAsteroid(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(credits = 14),
            globalChange = GlobalParameterChange(temperature = 2),
            terraformRatingGain = 1
        )

    suspend fun buildAquifer(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(credits = 18),
            globalChange = GlobalParameterChange(oceans = 1),
            terraformRatingGain = 1
        )

    suspend fun plantGreenery(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(credits = 23),
            globalChange = GlobalParameterChange(oxygen = 1),
            terraformRatingGain = 1
        )

    suspend fun buildCity(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(credits = 25),
            productionChange = ProductionChange(credits = 1, energy = -1)
        )

    suspend fun convertPlants(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(plants = 8),
            globalChange = GlobalParameterChange(oxygen = 1),
            terraformRatingGain = 1
        )

    suspend fun convertHeat(gameId: String, playerId: String): Game? =
        executeStandardProject(
            gameId, playerId,
            cost = ResourceCost(heat = 8),
            globalChange = GlobalParameterChange(temperature = 2),
            terraformRatingGain = 1
        )

    suspend fun sellPatents(gameId: String, playerId: String, cardIds: List<String>): Game? {
        val game = gameRepository.findById(gameId) ?: return null
        val player = game.players.find { it.id == playerId } ?: return null

        if (cardIds.isEmpty()) return null

        val cardsToSell = cardIds.mapNotNull { cardId ->
            player.cards.find { it.id == cardId }
        }

        if (cardsToSell.size != cardIds.size) return null

        val creditsEarned = cardsToSell.size * 10
        val remainingCards = player.cards.filter { !cardIds.contains(it.id) }

        val updatedPlayer = player.copy(
            cards = remainingCards,
            resources = player.resources.copy(credits = player.resources.credits + creditsEarned)
        )

        return updateGameWithPlayer(game, updatedPlayer, gameId)
    }

    private suspend fun executeStandardProject(
        gameId: String,
        playerId: String,
        cost: ResourceCost = ResourceCost(),
        globalChange: GlobalParameterChange = GlobalParameterChange(),
        productionChange: ProductionChange = ProductionChange(),
        terraformRatingGain: Int = 0
    ): Game? {
        val game = gameRepository.findById(gameId) ?: return null
        val player = game.players.find { it.id == playerId } ?: return null

        // Validate costs and constraints
        if (!canAffordCost(player, cost)) return null
        if (!canApplyGlobalChange(game, globalChange)) return null
        if (!canApplyProductionChange(player, productionChange)) return null

        // Apply changes
        val updatedPlayer = applyChangesToPlayer(player, cost, productionChange, terraformRatingGain)
        val updatedGame = applyGlobalChanges(game, globalChange)

        return updateGameWithPlayer(updatedGame, updatedPlayer, gameId)
    }

    private fun canAffordCost(player: Player, cost: ResourceCost): Boolean =
        player.resources.credits >= cost.credits &&
        player.resources.plants >= cost.plants &&
        player.resources.heat >= cost.heat

    private fun canApplyGlobalChange(game: Game, change: GlobalParameterChange): Boolean {
        val params = game.globalParameters
        return (change.temperature == 0 || params.temperature < 8) &&
               (change.oxygen == 0 || params.oxygen < 14) &&
               (change.oceans == 0 || params.oceans < 9)
    }

    private fun canApplyProductionChange(player: Player, change: ProductionChange): Boolean =
        player.resourceProduction.energy + change.energy >= 0

    private fun applyChangesToPlayer(
        player: Player,
        cost: ResourceCost,
        productionChange: ProductionChange,
        terraformRatingGain: Int
    ): Player {
        return player.copy(
            resources = player.resources.copy(
                credits = player.resources.credits - cost.credits,
                plants = player.resources.plants - cost.plants,
                heat = player.resources.heat - cost.heat
            ),
            resourceProduction = player.resourceProduction.copy(
                credits = player.resourceProduction.credits + productionChange.credits,
                energy = player.resourceProduction.energy + productionChange.energy
            ),
            terraformRating = player.terraformRating + terraformRatingGain
        )
    }

    private fun applyGlobalChanges(game: Game, change: GlobalParameterChange): Game {
        return game.copy(
            globalParameters = game.globalParameters.copy(
                temperature = game.globalParameters.temperature + change.temperature,
                oxygen = game.globalParameters.oxygen + change.oxygen,
                oceans = game.globalParameters.oceans + change.oceans
            )
        )
    }

    private suspend fun updateGameWithPlayer(game: Game, updatedPlayer: Player, gameId: String): Game {
        val updatedPlayers = game.players.map { if (it.id == updatedPlayer.id) updatedPlayer else it }
        val updatedGame = game.copy(players = updatedPlayers)

        val savedGame = gameRepository.update(updatedGame)
        eventNotifier.notifyGameUpdated(gameId, savedGame)
        return savedGame
    }
}