package com.terraformingmars.repositories

import com.terraformingmars.models.Game
import java.util.concurrent.ConcurrentHashMap

class GameRepository {
    private val games = ConcurrentHashMap<String, Game>()

    fun save(game: Game): Game {
        games[game.id] = game
        return game
    }

    fun findById(id: String): Game? {
        return games[id]
    }

    fun findAll(): List<Game> {
        return games.values.toList()
    }

    fun findByStatus(status: String): List<Game> {
        return games.values.filter { it.status.toString() == status }
    }

    fun delete(id: String) {
        games.remove(id)
    }

    fun update(game: Game): Game {
        games[game.id] = game
        return game
    }
}