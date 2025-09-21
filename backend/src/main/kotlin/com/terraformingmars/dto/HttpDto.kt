package com.terraformingmars.dto

import kotlinx.serialization.Serializable

// HTTP Request/Response DTOs matching the Go DTOs

@Serializable
data class CreateGameRequest(
    val maxPlayers: Int,
    val playerName: String
)

@Serializable
data class CreateGameResponse(
    val game: GameDto
)

@Serializable
data class JoinGameRequest(
    val playerName: String
)

@Serializable
data class JoinGameResponse(
    val game: GameDto,
    val playerId: String
)

@Serializable
data class GetGameResponse(
    val game: GameDto
)

@Serializable
data class ListGamesResponse(
    val games: List<GameDto>
)

@Serializable
data class UpdateResourcesRequest(
    val resources: ResourcesDto
)

@Serializable
data class GetPlayerResponse(
    val player: PlayerDto
)

@Serializable
data class UpdatePlayerResourcesResponse(
    val player: PlayerDto
)

@Serializable
data class ListCardsResponse(
    val cards: List<CardDto>,
    val totalCount: Int,
    val offset: Int,
    val limit: Int
)

@Serializable
data class ErrorResponse(
    val error: String,
    val code: String? = null,
    val details: String? = null
)