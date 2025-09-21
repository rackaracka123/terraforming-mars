package com.terraformingmars.services

import com.terraformingmars.dto.*
import com.terraformingmars.models.Card
import kotlinx.serialization.json.Json

object CardLoader {
    private val json = Json {
        ignoreUnknownKeys = true
        classDiscriminator = "type"
    }

    private var cachedCards: List<Card>? = null

    fun loadCards(resourcePath: String = "assets/terraforming_mars_cards.json"): List<Card> {
        if (cachedCards != null) {
            return cachedCards!!
        }

        val inputStream = this::class.java.classLoader.getResourceAsStream(resourcePath)
            ?: throw IllegalArgumentException("Cards resource not found at: $resourcePath")

        val jsonText = inputStream.bufferedReader().use { it.readText() }
        val cardDtos = json.decodeFromString<List<CardDto>>(jsonText)

        cachedCards = cardDtos.map { dto ->
            Card(
                id = dto.id,
                name = dto.name,
                type = dto.type,
                cost = dto.cost,
                description = dto.description,
                tags = dto.tags,
                requirements = dto.requirements,
                behaviors = dto.behaviors,
                resourceStorage = dto.resourceStorage,
                vpConditions = dto.vpConditions
            )
        }

        return cachedCards!!
    }

    fun getCardById(cardId: String): Card? {
        return loadCards().find { it.id == cardId }
    }

    fun getCardsByType(type: CardType): List<Card> {
        return loadCards().filter { it.type == type }
    }

    fun getCardsByTag(tag: CardTag): List<Card> {
        return loadCards().filter { it.tags?.contains(tag) == true }
    }

    fun clearCache() {
        cachedCards = null
    }
}