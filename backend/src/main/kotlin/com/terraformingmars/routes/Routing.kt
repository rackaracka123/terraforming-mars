package com.terraformingmars.routes

import com.terraformingmars.repositories.GameRepository
import com.terraformingmars.repositories.CardRepository
import com.terraformingmars.services.GameService
import com.terraformingmars.websocket.WebSocketSessionManager
import com.terraformingmars.websocket.webSocketRoutes
import io.ktor.server.application.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

fun Application.configureRouting() {
    // Create shared service instance (which creates its own repositories internally)
    val gameRepository = GameRepository()
    val cardRepository = CardRepository()
    val sessionManager = WebSocketSessionManager(gameRepository)
    val gameService = GameService(gameRepository, sessionManager, cardRepository)
    sessionManager.setGameService(gameService)

    routing {
        get("/") {
            call.respondText("Terraforming Mars Kotlin/Ktor Backend")
        }

        // Health check endpoint
        get("/health") {
            call.respondText("OK")
        }

        // WebSocket routes with shared sessionManager
        webSocketRoutes(sessionManager)

        // API routes with shared service
        route("/api/v1") {
            gameRoutes(gameService)
        }
    }
}