package com.terraformingmars.services

import com.terraformingmars.models.Game

interface GameEventNotifier {
    suspend fun notifyPlayerConnected(gameId: String, playerId: String, playerName: String, game: Game)
    suspend fun notifyPlayerReconnected(gameId: String, playerId: String, playerName: String, game: Game)
    suspend fun notifyPlayerDisconnected(gameId: String, playerId: String, playerName: String, game: Game)
    suspend fun notifyGameUpdated(gameId: String, game: Game)
    suspend fun notifyGameStarted(gameId: String, game: Game)
    suspend fun notifyError(playerId: String?, message: String)
}