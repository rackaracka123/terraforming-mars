package com.terraformingmars.services

import com.terraformingmars.dto.*
import com.terraformingmars.models.*
import com.terraformingmars.repositories.GameRepository
import com.terraformingmars.repositories.CardRepository
import java.util.UUID

class GameService(
    private val gameRepository: GameRepository,
    private val eventNotifier: GameEventNotifier,
    private val cardRepository: CardRepository
) {

    private val gameEngine = GameEngine(gameRepository, cardRepository, eventNotifier)
    private val standardProjects = StandardProjects(gameRepository, eventNotifier)

    fun createGame(maxPlayers: Int, hostPlayerName: String): Game {
        val gameId = UUID.randomUUID().toString()

        val hostPlayerId = UUID.randomUUID().toString()

        // Initialize the card deck for this game first
        cardRepository.initializeGameDeck(gameId)

        // Create the host player without cards (cards will be dealt when game starts)
        val hostPlayer = Player(
            id = hostPlayerId,
            name = hostPlayerName,
            corporation = null,
            cards = emptyList(), // No cards during lobby phase
            resources = Resources(0, 0, 0, 0, 0, 0),
            resourceProduction = Production(1, 0, 0, 0, 0, 0),
            terraformRating = 20,
            playedCards = emptyList(),
            passed = false,
            availableActions = 2,
            victoryPoints = 0,
            isConnected = true,
            productionSelection = null,
            startingSelection = null // No starting selection during lobby
        )

        val game = Game(
            id = gameId,
            status = GameStatus.LOBBY,
            settings = GameSettings(maxPlayers = maxPlayers),
            hostPlayerId = hostPlayerId,
            currentPhase = GamePhase.WAITING_FOR_GAME_START,
            globalParameters = GlobalParameters(
                temperature = -30,
                oxygen = 0,
                oceans = 0
            ),
            players = listOf(hostPlayer),
            currentTurn = null,
            generation = 1,
            remainingActions = 2,
            turnOrder = listOf(hostPlayerId)
        )

        return gameRepository.save(game)
    }

    suspend fun onPlayerConnect(gameId: String, playerId: String?, playerName: String): String? {
        println("DEBUG: GameService.onPlayerConnect - Looking for game with ID: $gameId")
        val game = gameRepository.findById(gameId)
        if (game == null) {
            println("DEBUG: Game not found in repository for ID: $gameId")
            println("DEBUG: Available games in repository:")
            gameRepository.findAll().forEach { g ->
                println("DEBUG: - Game ID: ${g.id}, Status: ${g.status}, Players: ${g.players.size}")
            }
            eventNotifier.notifyError(playerId, "Game not found")
            return null
        }
        println("DEBUG: Found game with ID: $gameId, Status: ${game.status}, Players: ${game.players.size}")

        if (playerId != null) {
            // Reconnection
            val existingPlayer = game.players.find { it.id == playerId }
            if (existingPlayer == null) {
                eventNotifier.notifyError(playerId, "Player not found in game")
                return null
            }
            // Note: Player.isConnected is not mutable, need to update via game repository
            gameRepository.update(game)
            eventNotifier.notifyPlayerReconnected(gameId, playerId, playerName, game)
            return playerId
        } else {
            // New player joining
            if (game.players.size >= game.settings.maxPlayers) {
                eventNotifier.notifyError(null, "Game is full")
                return null
            }

            if (game.status != GameStatus.LOBBY) {
                eventNotifier.notifyError(null, "Game has already started")
                return null
            }

            // Create new player
            val newPlayerId = UUID.randomUUID().toString()

            // Don't deal cards during lobby phase - cards will be dealt when game starts
            val player = Player(
                id = newPlayerId,
                name = playerName,
                corporation = null,
                cards = emptyList(), // No cards during lobby phase
                resources = Resources(0, 0, 0, 0, 0, 0),
                resourceProduction = Production(1, 0, 0, 0, 0, 0),
                terraformRating = 20,
                playedCards = emptyList(),
                passed = false,
                availableActions = 2,
                victoryPoints = 0,
                isConnected = true,
                productionSelection = null,
                startingSelection = null // No starting selection during lobby
            )

            // Add player to game (create new list since players is immutable)
            val updatedPlayers = game.players + player

            // Set as host if first player and add to turn order
            val updatedGame = if (game.hostPlayerId == null) {
                game.copy(hostPlayerId = newPlayerId, players = updatedPlayers, turnOrder = listOf(newPlayerId))
            } else {
                game.copy(players = updatedPlayers, turnOrder = game.turnOrder + newPlayerId)
            }

            gameRepository.update(updatedGame)
            eventNotifier.notifyPlayerConnected(gameId, newPlayerId, playerName, updatedGame)
            return newPlayerId
        }
    }

    suspend fun onPlayerDisconnected(gameId: String, playerId: String) {
        val game = gameRepository.findById(gameId)
        if (game == null) return

        val player = game.players.find { it.id == playerId }
        if (player == null) return

        // Note: Player.isConnected is not mutable, need to update via game copy
        gameRepository.update(game)

        eventNotifier.notifyPlayerDisconnected(gameId, playerId, player.name, game)
        eventNotifier.notifyGameUpdated(gameId, game)
    }

    data class JoinGameResult(val game: GameDto, val playerId: String)

    suspend fun onJoinGame(gameId: String, playerName: String): JoinGameResult? {
        val game = gameRepository.findById(gameId)
        if (game == null) {
            throw IllegalStateException("Game not found")
        }

        if (game.players.size >= game.settings.maxPlayers) {
            throw IllegalStateException("Game is full")
        }

        if (game.status != GameStatus.LOBBY) {
            throw IllegalStateException("Game has already started")
        }

        // Create new player
        val newPlayerId = UUID.randomUUID().toString()

        // Don't deal cards during lobby phase - cards will be dealt when game starts
        val player = Player(
            id = newPlayerId,
            name = playerName,
            corporation = null,
            cards = emptyList(), // No cards during lobby phase
            resources = Resources(0, 0, 0, 0, 0, 0),
            resourceProduction = Production(1, 0, 0, 0, 0, 0),
            terraformRating = 20,
            playedCards = emptyList(),
            passed = false,
            availableActions = 2,
            victoryPoints = 0,
            isConnected = true,
            productionSelection = null,
            startingSelection = null // No starting selection during lobby
        )

        // Add player to game (create new list since players is immutable)
        val updatedPlayers = game.players + player

        // Set as host if first player and add to turn order
        val updatedGame = if (game.hostPlayerId == null) {
            game.copy(hostPlayerId = newPlayerId, players = updatedPlayers, turnOrder = listOf(newPlayerId))
        } else {
            game.copy(players = updatedPlayers, turnOrder = game.turnOrder + newPlayerId)
        }

        gameRepository.update(updatedGame)
        eventNotifier.notifyPlayerConnected(gameId, newPlayerId, playerName, updatedGame)
        eventNotifier.notifyGameUpdated(gameId, updatedGame)

        return JoinGameResult(
            game = updatedGame.toDto(newPlayerId),
            playerId = newPlayerId
        )
    }

    // Game engine methods
    suspend fun startGame(gameId: String): Game? {
        return gameEngine.startGame(gameId)
    }

    suspend fun selectStartingCards(gameId: String, playerId: String, selectedCardIds: List<String>): Game? {
        return gameEngine.selectStartingCards(gameId, playerId, selectedCardIds)
    }

    suspend fun endTurn(gameId: String, playerId: String): Game? {
        return gameEngine.endTurn(gameId, playerId)
    }

    suspend fun processProductionPhase(gameId: String): Game? {
        return gameEngine.processProductionPhase(gameId)
    }

    // Standard project methods
    suspend fun buildPowerPlant(gameId: String, playerId: String): Game? {
        return standardProjects.buildPowerPlant(gameId, playerId)
    }

    suspend fun launchAsteroid(gameId: String, playerId: String): Game? {
        return standardProjects.launchAsteroid(gameId, playerId)
    }

    suspend fun buildAquifer(gameId: String, playerId: String): Game? {
        return standardProjects.buildAquifer(gameId, playerId)
    }

    suspend fun plantGreenery(gameId: String, playerId: String): Game? {
        return standardProjects.plantGreenery(gameId, playerId)
    }

    suspend fun buildCity(gameId: String, playerId: String): Game? {
        return standardProjects.buildCity(gameId, playerId)
    }

    suspend fun sellPatents(gameId: String, playerId: String, cardIds: List<String>): Game? {
        return standardProjects.sellPatents(gameId, playerId, cardIds)
    }

    suspend fun convertPlants(gameId: String, playerId: String): Game? {
        return standardProjects.convertPlants(gameId, playerId)
    }

    suspend fun convertHeat(gameId: String, playerId: String): Game? {
        return standardProjects.convertHeat(gameId, playerId)
    }

    // Handle WebSocket commands directly
    suspend fun handleWebSocketCommand(command: WebSocketCommand): Game? {
        val gameId = command.gameId ?: return null
        val playerId = command.playerId ?: return null

        return when (command) {
            is BuildPowerPlantCommand -> buildPowerPlant(gameId, playerId)
            is LaunchAsteroidCommand -> launchAsteroid(gameId, playerId)
            is BuildAquiferCommand -> buildAquifer(gameId, playerId)
            is PlantGreeneryCommand -> plantGreenery(gameId, playerId)
            is BuildCityCommand -> buildCity(gameId, playerId)
            is StartGameCommand -> startGame(gameId)
            is SelectStartingCardCommand -> selectStartingCards(gameId, playerId, command.cardIds)
            else -> null // Command not handled
        }
    }

    // Query methods for HTTP controllers
    fun getGame(gameId: String): Game? {
        return gameRepository.findById(gameId)
    }

    fun listGames(): List<Game> {
        return gameRepository.findAll()
    }

    // Cards service methods
    fun listCards(offset: Int = 0, limit: Int = 50): List<Card> {
        return cardRepository.getAllCards(offset, limit)
    }

    fun getTotalCardCount(): Int {
        return cardRepository.getTotalCardCount()
    }
}