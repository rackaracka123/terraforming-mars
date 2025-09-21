package com.terraformingmars.websocket

import com.terraformingmars.dto.*
import com.terraformingmars.models.Game
import com.terraformingmars.repositories.GameRepository
import com.terraformingmars.services.GameEventNotifier
import com.terraformingmars.services.GameService
import io.ktor.server.websocket.*
import io.ktor.server.websocket.sendSerialized
import io.ktor.websocket.*
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.time.LocalDateTime
import java.util.*
import java.util.concurrent.ConcurrentHashMap

enum class SessionState {
    PENDING,
    CONNECTING,
    CONNECTED,
    DISCONNECTED
}

data class SessionData(
    val sessionId: String,
    val session: WebSocketServerSession,
    var state: SessionState,
    var playerId: String? = null,
    var gameId: String? = null,
    var playerName: String? = null,
    val connectedAt: LocalDateTime = LocalDateTime.now(),
    var lastActivity: LocalDateTime = LocalDateTime.now()
)

class WebSocketSessionManager(
    private val gameRepository: GameRepository
) : GameEventNotifier {
    private lateinit var gameService: GameService

    fun setGameService(gameService: GameService) {
        this.gameService = gameService
    }
    private val sessions = ConcurrentHashMap<String, SessionData>()
    private val playerToSession = ConcurrentHashMap<String, String>()
    private val gameToSessions = ConcurrentHashMap<String, MutableSet<String>>()
    private val mutex = Mutex()
    private val json = Json {
        ignoreUnknownKeys = true
        classDiscriminator = "type"
    }

    suspend fun addSession(session: WebSocketServerSession): String {
        val sessionId = UUID.randomUUID().toString()
        val sessionData = SessionData(
            sessionId = sessionId,
            session = session,
            state = SessionState.PENDING
        )

        sessions[sessionId] = sessionData
        println("WebSocket session added: $sessionId")
        return sessionId
    }

    suspend fun removeSession(session: WebSocketServerSession) = mutex.withLock {
        val sessionData = sessions.values.find { it.session == session }
        if (sessionData != null) {
            sessionData.state = SessionState.DISCONNECTED

            sessionData.playerId?.let { playerId ->
                playerToSession.remove(playerId)

                sessionData.gameId?.let { gameId ->
                    gameToSessions[gameId]?.remove(sessionData.sessionId)
                    if (gameToSessions[gameId]?.isEmpty() == true) {
                        gameToSessions.remove(gameId)
                    }

                    gameService.onPlayerDisconnected(gameId, playerId)
                }
            }

            sessions.remove(sessionData.sessionId)
            println("WebSocket session removed: ${sessionData.sessionId}")
        }
    }

    suspend fun handleCommand(session: WebSocketServerSession, command: WebSocketCommand) = mutex.withLock {
        val sessionData = sessions.values.find { it.session == session }
        if (sessionData == null) {
            sendError(session, "Session not found")
            return@withLock
        }

        sessionData.lastActivity = LocalDateTime.now()

        try {
            when (command) {
                is PlayerConnectCommand -> handlePlayerConnect(sessionData, command)
                is StartGameCommand -> handleGameCommand(sessionData, command)
                is SkipActionCommand -> handleGameCommand(sessionData, command)
                is SellPatentsCommand -> handleGameCommand(sessionData, command)
                is BuildPowerPlantCommand -> handleGameCommand(sessionData, command)
                is LaunchAsteroidCommand -> handleGameCommand(sessionData, command)
                is BuildAquiferCommand -> handleGameCommand(sessionData, command)
                is PlantGreeneryCommand -> handleGameCommand(sessionData, command)
                is BuildCityCommand -> handleGameCommand(sessionData, command)
                is PlayCardCommand -> handleGameCommand(sessionData, command)
                is SelectStartingCardCommand -> handleGameCommand(sessionData, command)
                is SelectCardsCommand -> handleGameCommand(sessionData, command)
            }
        } catch (e: Exception) {
            println("Error handling command: ${e.message}")
            sendError(session, "Failed to process command: ${e.message}")
        }
    }

    private suspend fun handlePlayerConnect(sessionData: SessionData, command: PlayerConnectCommand) {
        if (sessionData.state != SessionState.PENDING) {
            sendError(sessionData.session, "Session already connected")
            return
        }

        // Let GameService handle the domain logic and notifications
        println("Calling gameService.onPlayerConnect with gameId=${command.gameId}, playerId=${command.playerId}, playerName=${command.playerName}")
        val actualPlayerId = gameService.onPlayerConnect(command.gameId, command.playerId, command.playerName)
        println("GameService returned actualPlayerId: $actualPlayerId")

        if (actualPlayerId == null) {
            println("Failed to get playerId from GameService, sending error")
            sendError(sessionData.session, "Failed to connect player to game")
            return
        }

        // Update session state and mapping
        println("Updating session state to CONNECTED")
        sessionData.state = SessionState.CONNECTED
        sessionData.playerId = actualPlayerId
        sessionData.gameId = command.gameId
        sessionData.playerName = command.playerName

        playerToSession[actualPlayerId] = sessionData.sessionId
        gameToSessions.getOrPut(command.gameId) { mutableSetOf() }.add(sessionData.sessionId)

        // Send full state to the connecting player
        println("Finding game by ID: ${command.gameId}")
        val game = gameRepository.findById(command.gameId)
        if (game != null) {
            println("Found game, sending full state to player: $actualPlayerId")
            sendFullState(sessionData.session, game.toDto(actualPlayerId), actualPlayerId)
            println("Full state sent successfully")
        } else {
            println("ERROR: Game not found with ID: ${command.gameId}")
        }
    }

    private suspend fun handleGameCommand(sessionData: SessionData, command: WebSocketCommand) {
        if (sessionData.state != SessionState.CONNECTED) {
            sendError(sessionData.session, "Player not connected")
            return
        }

        val playerId = sessionData.playerId ?: run {
            sendError(sessionData.session, "Player ID not found")
            return
        }

        val gameId = command.gameId ?: run {
            sendError(sessionData.session, "Game ID not found")
            return
        }

        // Handle the command through GameService
        val result = gameService.handleWebSocketCommand(command)
        if (result == null) {
            sendError(sessionData.session, "Command failed or invalid")
        }
    }

    private suspend fun broadcast(gameId: String, event: WebSocketEvent, excludePlayerId: String? = null) {
        gameToSessions[gameId]?.forEach { sessionId ->
            sessions[sessionId]?.let { sessionData ->
                if (sessionData.playerId != excludePlayerId) {
                    sessionData.session.sendSerialized(event)
                }
            }
        }
    }

    suspend fun sendError(session: WebSocketServerSession, message: String) {
        session.sendSerialized(ErrorEvent(message))
    }

    private suspend fun sendFullState(session: WebSocketServerSession, game: GameDto, playerId: String) {
        session.sendSerialized(FullStateEvent(game, playerId))
    }

    private suspend fun broadcastGameUpdate(gameId: String, game: Game) {
        gameToSessions[gameId]?.forEach { sessionId ->
            sessions[sessionId]?.let { sessionData ->
                sessionData.playerId?.let { playerId ->
                    sessionData.session.sendSerialized(GameUpdatedEvent(game.toDto(playerId)))
                }
            }
        }
    }

    private suspend fun broadcastToGame(
        gameId: String,
        game: Game,
        excludePlayerId: String? = null,
        eventFactory: (viewingPlayerId: String) -> WebSocketEvent
    ) {
        gameToSessions[gameId]?.forEach { sessionId ->
            sessions[sessionId]?.let { sessionData ->
                if (sessionData.playerId != excludePlayerId) {
                    sessionData.playerId?.let { viewingPlayerId ->
                        sessionData.session.sendSerialized(eventFactory(viewingPlayerId))
                    }
                }
            }
        }
    }

    // GameEventNotifier implementation
    override suspend fun notifyPlayerConnected(gameId: String, playerId: String, playerName: String, game: Game) {
        broadcastToGame(gameId, game, playerId) { viewingPlayerId ->
            PlayerConnectedEvent(playerId, playerName, game.toDto(viewingPlayerId))
        }
    }

    override suspend fun notifyPlayerReconnected(gameId: String, playerId: String, playerName: String, game: Game) {
        broadcastToGame(gameId, game, playerId) { viewingPlayerId ->
            PlayerReconnectedEvent(playerId, playerName, game.toDto(viewingPlayerId))
        }
    }

    override suspend fun notifyPlayerDisconnected(gameId: String, playerId: String, playerName: String, game: Game) {
        broadcastToGame(gameId, game, playerId) { viewingPlayerId ->
            PlayerDisconnectedEvent(playerId, playerName, game.toDto(viewingPlayerId))
        }
    }

    override suspend fun notifyGameUpdated(gameId: String, game: Game) {
        broadcastGameUpdate(gameId, game)
    }

    override suspend fun notifyGameStarted(gameId: String, game: Game) {
        broadcastGameUpdate(gameId, game)
    }

    override suspend fun notifyError(playerId: String?, message: String) {
        if (playerId != null) {
            playerToSession[playerId]?.let { sessionId ->
                sessions[sessionId]?.let { sessionData ->
                    sessionData.session.sendSerialized(ErrorEvent(message))
                }
            }
        }
    }
}
