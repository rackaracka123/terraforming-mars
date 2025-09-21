package com.terraformingmars.repositories

import com.terraformingmars.dto.CardType
import com.terraformingmars.models.Card
import com.terraformingmars.services.CardLoader
import java.util.concurrent.ConcurrentHashMap

data class GameDeck(
    val availableDeck: MutableList<Card> = mutableListOf(),
    val discardPile: MutableList<Card> = mutableListOf(),
    val removedCards: MutableList<Card> = mutableListOf()
)

class CardRepository {
    private val gameDecks = ConcurrentHashMap<String, GameDeck>()

    /**
     * Initialize a new deck for a game
     */
    fun initializeGameDeck(gameId: String) {
        // Load all project cards (excluding corporations)
        val allCards = CardLoader.loadCards()
        val projectCards = allCards.filter { it.type != CardType.CORPORATION }

        // Shuffle the deck for this game
        val shuffledDeck = projectCards.shuffled().toMutableList()

        gameDecks[gameId] = GameDeck(availableDeck = shuffledDeck)
    }

    /**
     * Remove a game's deck when game is completed or deleted
     */
    fun removeGameDeck(gameId: String) {
        gameDecks.remove(gameId)
    }

    private fun getGameDeck(gameId: String): GameDeck {
        return gameDecks[gameId] ?: throw IllegalArgumentException("Game deck not found for game: $gameId")
    }

    /**
     * Deal a specified number of cards to a player
     */
    fun dealCards(gameId: String, count: Int): List<Card> {
        val gameDeck = getGameDeck(gameId)

        if (gameDeck.availableDeck.size < count) {
            reshuffleDiscardPile(gameDeck)
        }

        if (gameDeck.availableDeck.size < count) {
            // Not enough cards even after reshuffling - deal what we can
            val cardsToReturn = gameDeck.availableDeck.take(gameDeck.availableDeck.size)
            gameDeck.availableDeck.clear()
            return cardsToReturn
        }

        val dealtCards = gameDeck.availableDeck.take(count)
        repeat(count) { gameDeck.availableDeck.removeAt(0) }
        return dealtCards
    }

    /**
     * Add cards to the discard pile
     */
    fun discardCards(gameId: String, cards: List<Card>) {
        val gameDeck = getGameDeck(gameId)
        gameDeck.discardPile.addAll(cards)
    }

    /**
     * Remove cards from the game entirely (for cards that are removed from play)
     */
    fun removeCards(gameId: String, cards: List<Card>) {
        val gameDeck = getGameDeck(gameId)
        gameDeck.removedCards.addAll(cards)
    }

    /**
     * Reshuffle discard pile back into available deck
     */
    private fun reshuffleDiscardPile(gameDeck: GameDeck) {
        if (gameDeck.discardPile.isNotEmpty()) {
            gameDeck.availableDeck.addAll(gameDeck.discardPile.shuffled())
            gameDeck.discardPile.clear()
        }
    }

    /**
     * Get the number of cards remaining in deck
     */
    fun getRemainingCardCount(gameId: String): Int {
        val gameDeck = getGameDeck(gameId)
        return gameDeck.availableDeck.size
    }

    /**
     * Get the number of cards in discard pile
     */
    fun getDiscardPileCount(gameId: String): Int {
        val gameDeck = getGameDeck(gameId)
        return gameDeck.discardPile.size
    }

    /**
     * Get all corporation cards for selection
     */
    fun getAllCorporations(): List<Card> {
        val allCards = CardLoader.loadCards()
        return allCards.filter { it.type == CardType.CORPORATION }
    }

    /**
     * Get a specific card by ID (useful for card effects that search for specific cards)
     */
    fun getCardById(cardId: String): Card? {
        return CardLoader.getCardById(cardId)
    }

    /**
     * Check if a specific card is still available in the deck
     */
    fun isCardAvailable(gameId: String, cardId: String): Boolean {
        val gameDeck = getGameDeck(gameId)
        return gameDeck.availableDeck.any { it.id == cardId }
    }

    /**
     * Search for cards meeting specific criteria in the available deck
     */
    fun searchAvailableCards(gameId: String, predicate: (Card) -> Boolean): List<Card> {
        val gameDeck = getGameDeck(gameId)
        return gameDeck.availableDeck.filter(predicate)
    }

    /**
     * Get all cards with pagination support (for API endpoint)
     */
    fun getAllCards(offset: Int = 0, limit: Int = 50): List<Card> {
        val allCards = CardLoader.loadCards()
        return if (offset >= allCards.size) {
            emptyList()
        } else {
            allCards.drop(offset).take(limit)
        }
    }

    /**
     * Get total count of all cards (for API endpoint)
     */
    fun getTotalCardCount(): Int {
        return CardLoader.loadCards().size
    }
}