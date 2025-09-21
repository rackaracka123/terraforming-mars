package com.terraformingmars.websocket

import com.terraformingmars.dto.*
import com.terraformingmars.repositories.GameRepository
import com.terraformingmars.repositories.CardRepository
import com.terraformingmars.services.GameService
import io.ktor.serialization.kotlinx.*
import io.ktor.server.application.*
import io.ktor.server.routing.*
import io.ktor.server.websocket.*
import io.ktor.websocket.*
import kotlinx.coroutines.flow.consumeAsFlow
import kotlinx.coroutines.flow.collect
import kotlinx.serialization.json.Json
import kotlinx.serialization.decodeFromString
import kotlin.time.Duration.Companion.seconds

fun Application.configureWebSockets() {
    install(WebSockets) {
        pingPeriod = 15.seconds
        timeout = 15.seconds
        maxFrameSize = Long.MAX_VALUE
        masking = false
        contentConverter = KotlinxWebsocketSerializationConverter(Json {
            classDiscriminator = "type"
        })
    }
}

fun Route.webSocketRoutes(sessionManager: WebSocketSessionManager) {
    webSocket("/ws") {
        val json = Json {
            ignoreUnknownKeys = true
            classDiscriminator = "type"
        }

        sessionManager.addSession(this)
        println("WebSocket connection established")

        try {
            incoming.consumeAsFlow().collect { frame ->
                when (frame) {
                    is Frame.Text -> {
                        val text = frame.readText()
                        println("Received WebSocket message: $text")

                        try {
                            val command = json.decodeFromString<WebSocketCommand>(text)
                            println("Parsed command: ${command::class.simpleName}")
                            sessionManager.handleCommand(this@webSocket, command)
                        } catch (parseError: Exception) {
                            println("Failed to parse command: ${parseError.message}")
                            sessionManager.sendError(this@webSocket, "Invalid message format: ${parseError.message}")
                        }
                    }
                    is Frame.Close -> {
                        println("Received close frame")
                    }
                    else -> {
                        println("Received other frame type: ${frame::class.simpleName}")
                    }
                }
            }
        } catch (e: Exception) {
            println("WebSocket error: ${e.message}")
            e.printStackTrace()
            sessionManager.sendError(this, "Connection error: ${e.message}")
        } finally {
            println("WebSocket connection closing")
            sessionManager.removeSession(this)
        }
    }
}