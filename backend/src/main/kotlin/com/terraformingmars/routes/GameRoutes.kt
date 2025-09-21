package com.terraformingmars.routes

import com.terraformingmars.dto.*
import com.terraformingmars.repositories.GameRepository
import com.terraformingmars.repositories.CardRepository
import com.terraformingmars.services.GameService
import com.terraformingmars.websocket.WebSocketSessionManager
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

fun Route.gameRoutes(gameService: GameService) {

    // POST /api/v1/games - Create game
    post("/games") {
        val request = call.receive<CreateGameRequest>()
        val game = gameService.createGame(request.maxPlayers, request.playerName)

        call.respond(HttpStatusCode.Created, CreateGameResponse(game = game.toDto(game.hostPlayerId!!)))
    }

    // GET /api/v1/games/{id} - Get game by ID
    get("/games/{id}") {
        val gameId = call.parameters["id"]!!
        val game = gameService.getGame(gameId) ?: return@get call.respond(
            HttpStatusCode.NotFound,
            ErrorResponse(error = "Game not found")
        )

        call.respond(GetGameResponse(game = game.toDto(game.hostPlayerId!!)))
    }

    // GET /api/v1/games - List games
    get("/games") {
        val games = gameService.listGames()
        val gameDtos = games.map { game ->
            game.toDto(game.hostPlayerId!!)
        }

        call.respond(ListGamesResponse(games = gameDtos))
    }

    // POST /api/v1/games/{id}/join - Join game
    post("/games/{id}/join") {
        val gameId = call.parameters["id"]!!
        val request = call.receive<JoinGameRequest>()

        try {
            val result = gameService.onJoinGame(gameId, request.playerName)
            if (result != null) {
                call.respond(JoinGameResponse(
                    game = result.game,
                    playerId = result.playerId
                ))
            } else {
                call.respond(
                    HttpStatusCode.BadRequest,
                    ErrorResponse(error = "Unable to join game")
                )
            }
        } catch (e: IllegalStateException) {
            call.respond(
                HttpStatusCode.BadRequest,
                ErrorResponse(error = e.message ?: "Unable to join game")
            )
        } catch (e: Exception) {
            call.respond(
                HttpStatusCode.InternalServerError,
                ErrorResponse(error = "Server error")
            )
        }
    }

    // GET /api/v1/cards - List all cards with pagination
    get("/cards") {
        val offset = call.request.queryParameters["offset"]?.toIntOrNull() ?: 0
        val limit = call.request.queryParameters["limit"]?.toIntOrNull() ?: 50

        try {
            val cards = gameService.listCards(offset, limit)
            val totalCount = gameService.getTotalCardCount()

            val cardDtos = cards.map { it.toDto() }

            call.respond(ListCardsResponse(
                cards = cardDtos,
                totalCount = totalCount,
                offset = offset,
                limit = limit
            ))
        } catch (e: Exception) {
            call.respond(
                HttpStatusCode.InternalServerError,
                ErrorResponse(error = "Failed to list cards: ${e.message}")
            )
        }
    }
}