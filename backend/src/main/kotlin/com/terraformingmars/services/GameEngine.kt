package com.terraformingmars.services

import com.terraformingmars.dto.*
import com.terraformingmars.models.*
import com.terraformingmars.repositories.CardRepository
import com.terraformingmars.repositories.GameRepository

class GameEngine(
    private val gameRepository: GameRepository,
    private val cardRepository: CardRepository,
    private val eventNotifier: GameEventNotifier
) {

    /**
     * Start the game - transition from lobby to active gameplay
     */
    suspend fun startGame(gameId: String): Game? {
        val game = gameRepository.findById(gameId) ?: return null

        if (game.status != GameStatus.LOBBY) {
            return null // Game already started or completed
        }

        if (game.players.isEmpty()) {
            return null // Need at least 1 player to start
        }

        // Deal starting corporations and cards to each player
        val gameWithCards = dealStartingCards(game)

        // Update game state
        val updatedGame = gameWithCards.copy(
            status = GameStatus.ACTIVE,
            currentPhase = GamePhase.STARTING_CARD_SELECTION,
            turnOrder = generateTurnOrder(gameWithCards.players)
        )

        val savedGame = gameRepository.update(updatedGame)
        eventNotifier.notifyGameStarted(gameId, savedGame)

        return savedGame
    }

    /**
     * Process starting card selection from a player
     */
    suspend fun selectStartingCards(gameId: String, playerId: String, selectedCardIds: List<String>): Game? {
        val game = gameRepository.findById(gameId) ?: return null
        val player = game.players.find { it.id == playerId } ?: return null

        if (game.currentPhase != GamePhase.STARTING_CARD_SELECTION) {
            return null // Not in starting card selection phase
        }

        // Validate selected cards are from player's starting selection
        val startingSelection = player.startingSelection ?: return null
        val selectedCards = selectedCardIds.mapNotNull { cardId ->
            startingSelection.find { it.id == cardId }
        }

        // Calculate cost (3 credits per card kept)
        val cardCost = selectedCards.size * 3
        if (player.resources.credits < cardCost) {
            return null // Not enough credits
        }

        // Update player with selected cards and pay cost
        val updatedPlayer = player.copy(
            cards = selectedCards.map { cardDto ->
                cardRepository.getCardById(cardDto.id)!!
            },
            resources = player.resources.copy(credits = player.resources.credits - cardCost),
            startingSelection = null // Clear starting selection
        )

        // Update the player in the game
        val updatedPlayers = game.players.map { if (it.id == playerId) updatedPlayer else it }
        val gameWithUpdatedPlayer = game.copy(players = updatedPlayers)

        // Discard unselected cards back to deck
        val unselectedCards = startingSelection.filter { !selectedCardIds.contains(it.id) }
        val discardCards = unselectedCards.mapNotNull { cardRepository.getCardById(it.id) }
        cardRepository.discardCards(gameId, discardCards)

        // Check if all players have completed selection
        val allPlayersSelected = gameWithUpdatedPlayer.players.all { it.startingSelection?.isEmpty() ?: true }

        val finalGame = if (allPlayersSelected) {
            // Move to action phase
            gameWithUpdatedPlayer.copy(
                currentPhase = GamePhase.ACTION,
                currentTurn = gameWithUpdatedPlayer.turnOrder.firstOrNull()
            )
        } else {
            gameWithUpdatedPlayer
        }

        val savedGame = gameRepository.update(finalGame)
        eventNotifier.notifyGameUpdated(gameId, savedGame)

        return savedGame
    }

    /**
     * Process the end of a player's turn
     */
    suspend fun endTurn(gameId: String, playerId: String): Game? {
        val game = gameRepository.findById(gameId) ?: return null

        if (game.currentTurn != playerId) {
            return null // Not this player's turn
        }

        if (game.currentPhase != GamePhase.ACTION) {
            return null // Not in action phase
        }

        // Move to next player in turn order
        val currentIndex = game.turnOrder.indexOf(playerId)
        val nextIndex = (currentIndex + 1) % game.turnOrder.size
        val nextPlayerId = game.turnOrder[nextIndex]

        // Check if we've completed a full round
        val isEndOfGeneration = nextIndex == 0

        val updatedGame = if (isEndOfGeneration) {
            // Move to production phase
            game.copy(
                currentPhase = GamePhase.PRODUCTION_AND_CARD_DRAW,
                currentTurn = null,
                generation = game.generation + 1
            )
        } else {
            // Next player's turn
            game.copy(currentTurn = nextPlayerId)
        }

        val savedGame = gameRepository.update(updatedGame)
        eventNotifier.notifyGameUpdated(gameId, savedGame)

        return savedGame
    }

    /**
     * Process production phase for all players
     */
    suspend fun processProductionPhase(gameId: String): Game? {
        val game = gameRepository.findById(gameId) ?: return null

        if (game.currentPhase != GamePhase.PRODUCTION_AND_CARD_DRAW) {
            return null
        }

        // Apply production to all players
        val updatedPlayers = game.players.map { player ->
            player.copy(
                resources = Resources(
                    credits = player.resources.credits + player.resourceProduction.credits,
                    steel = player.resources.steel + player.resourceProduction.steel,
                    titanium = player.resources.titanium + player.resourceProduction.titanium,
                    plants = player.resources.plants + player.resourceProduction.plants,
                    energy = player.resources.energy + player.resourceProduction.energy,
                    heat = player.resources.heat + player.resourceProduction.heat + player.resources.energy // Convert energy to heat
                )
            )
        }

        // Deal new cards to each player (4 cards, player chooses how many to keep)
        val playersWithCards = updatedPlayers.map { player ->
            val newCards = cardRepository.dealCards(gameId, 4)
            player.copy(
                resources = player.resources.copy(energy = 0), // Energy converted to heat
                productionSelection = ProductionPhase(
                    availableCards = newCards,
                    selectionComplete = false
                )
            )
        }

        val updatedGame = game.copy(
            players = playersWithCards
        )

        val savedGame = gameRepository.update(updatedGame)
        eventNotifier.notifyGameUpdated(gameId, savedGame)

        return savedGame
    }

    /**
     * Check if the game should end (all global parameters maxed)
     */
    fun checkGameEnd(game: Game): Boolean {
        return game.globalParameters.temperature >= 8 &&
               game.globalParameters.oxygen >= 14 &&
               game.globalParameters.oceans >= 9
    }

    private fun dealStartingCards(game: Game): Game {
        // Each player gets 10 cards to choose from and 2 corporation cards
        val updatedPlayers = game.players.map { player ->
            val corporations = cardRepository.getAllCorporations().shuffled().take(2)
            val startingCards = cardRepository.dealCards(game.id, 10)

            // Update player with starting selection
            player.copy(
                startingSelection = startingCards
                // TODO: Also provide corporation selection
            )
        }

        return game.copy(players = updatedPlayers)
    }

    private fun generateTurnOrder(players: List<Player>): List<String> {
        // For now, just shuffle the players
        // TODO: Implement proper turn order based on corporation selection
        return players.map { it.id }.shuffled()
    }
}