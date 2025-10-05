import React, { useEffect, useState, useCallback, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { apiService } from "../../services/apiService";
import SimpleGameCard from "../ui/cards/SimpleGameCard";
import CopyLinkButton from "../ui/buttons/CopyLinkButton";
import { CardDto } from "@/types/generated/api-types";
import GameIcon from "../ui/display/GameIcon.tsx";

const CardsPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [cards, setCards] = useState<CardDto[]>([]);
  const [allCards, setAllCards] = useState<CardDto[]>([]);
  const [selectedCards, setSelectedCards] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isScrolled, setIsScrolled] = useState(false);
  const [visibleRange, setVisibleRange] = useState({ start: 0, end: 50 });
  const [sortBy, setSortBy] = useState<
    "unsorted" | "name-asc" | "name-desc" | "type-asc" | "type-desc"
  >("unsorted");
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());
  const [selectedTypes, setSelectedTypes] = useState<Set<string>>(new Set());
  const [showFilters, setShowFilters] = useState(false);
  const [showSortDropdown, setShowSortDropdown] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);

  // Get card type colors matching the CSS
  const getCardTypeColor = (type: string) => {
    const colorMap: { [key: string]: string } = {
      automated: "#4caf50",
      active: "#2196f3",
      event: "#f44336",
      corporation: "#ffc107",
      prelude: "#e91e63",
    };
    return colorMap[type.toLowerCase()] || "#4a90e2";
  };

  // Initialize state from URL parameters
  const initializeFromURL = useCallback(() => {
    const query = searchParams.get("q") || "";
    const sort = (searchParams.get("sort") as typeof sortBy) || "unsorted";
    const tags = searchParams.get("tags");
    const types = searchParams.get("types");

    setSearchQuery(query);
    setSortBy(sort);
    setSelectedTags(new Set(tags ? tags.split(",").filter(Boolean) : []));
    setSelectedTypes(new Set(types ? types.split(",").filter(Boolean) : []));

    // Don't select cards when viewing a permalink - just show them
    // Don't auto-close filters panel when filters are removed
  }, [searchParams]);

  // Update URL parameters when state changes
  const updateURL = useCallback(() => {
    const params = new URLSearchParams();

    // Preserve cId parameters if they exist
    const cardIds = searchParams.getAll("cId");
    cardIds.forEach((id) => params.append("cId", id));

    if (searchQuery.trim()) {
      params.set("q", searchQuery.trim());
    }

    if (sortBy !== "unsorted") {
      params.set("sort", sortBy);
    }

    if (selectedTags.size > 0) {
      params.set("tags", Array.from(selectedTags).sort().join(","));
    }

    if (selectedTypes.size > 0) {
      params.set("types", Array.from(selectedTypes).sort().join(","));
    }

    setSearchParams(params, { replace: true });
  }, [
    searchQuery,
    sortBy,
    selectedTags,
    selectedTypes,
    searchParams,
    setSearchParams,
  ]);

  const loadAllCards = useCallback(async () => {
    if (loading) return;

    setLoading(true);
    setError(null);

    try {
      // Load all cards by setting a very high limit
      const response = await apiService.listCards(0, 10000);
      setAllCards(response.cards);
      setCards(response.cards);
    } catch (err: any) {
      setError(err.message || "Failed to load cards");
    } finally {
      setLoading(false);
    }
  }, [loading]);

  // Get unique tags and types from all cards
  const availableTags = useMemo(() => {
    const tags = new Set<string>();
    allCards.forEach((card) => {
      card.tags.forEach((tag) => tags.add(tag));
    });
    return Array.from(tags).sort();
  }, [allCards]);

  const availableTypes = useMemo(() => {
    const types = new Set<string>();
    allCards.forEach((card) => {
      if (card.type) types.add(card.type);
    });
    return Array.from(types).sort();
  }, [allCards]);

  // Combined filter and sort functionality
  const applyFiltersAndSort = useCallback(() => {
    let filtered = [...allCards];

    // Check if we have card IDs in URL - if so, only show those cards
    const cardIds = searchParams.getAll("cId");
    if (cardIds.length > 0) {
      filtered = filtered.filter((card) => cardIds.includes(card.id));
    } else {
      // Apply search filter (including ID search)
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase();
        filtered = filtered.filter(
          (card) =>
            card.id.toLowerCase().includes(query) ||
            card.name.toLowerCase().includes(query) ||
            card.description.toLowerCase().includes(query) ||
            card.tags.some((tag) => tag.toLowerCase().includes(query)),
        );
      }

      // Apply tag filter (if any tags selected, show cards with ANY of the selected tags)
      if (selectedTags.size > 0) {
        filtered = filtered.filter((card) =>
          card.tags.some((tag) => selectedTags.has(tag)),
        );
      }

      // Apply type filter (if any types selected, show cards with ANY of the selected types)
      if (selectedTypes.size > 0) {
        filtered = filtered.filter((card) => selectedTypes.has(card.type));
      }
    }

    // Apply sorting (only if not unsorted)
    if (sortBy !== "unsorted") {
      filtered.sort((a, b) => {
        switch (sortBy) {
          case "name-asc":
            return a.name.localeCompare(b.name);
          case "name-desc":
            return b.name.localeCompare(a.name);
          case "type-asc":
            return a.type.localeCompare(b.type);
          case "type-desc":
            return b.type.localeCompare(a.type);
          default:
            return 0;
        }
      });
    }

    setCards(filtered);
  }, [
    allCards,
    searchQuery,
    selectedTags,
    selectedTypes,
    sortBy,
    searchParams,
  ]);

  // Search functionality
  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
  }, []);

  // Tag toggle functionality
  const toggleTag = useCallback((tag: string) => {
    setSelectedTags((prev) => {
      const newTags = new Set(prev);
      if (newTags.has(tag)) {
        newTags.delete(tag);
      } else {
        newTags.add(tag);
      }
      return newTags;
    });
  }, []);

  // Type toggle functionality
  const toggleType = useCallback((type: string) => {
    setSelectedTypes((prev) => {
      const newTypes = new Set(prev);
      if (newTypes.has(type)) {
        newTypes.delete(type);
      } else {
        newTypes.add(type);
      }
      return newTypes;
    });
  }, []);

  // Sort option selection with toggle functionality
  const handleSortSelection = useCallback(
    (selectedSort: typeof sortBy) => {
      if (selectedSort === sortBy) {
        // Toggle off - return to unsorted
        setSortBy("unsorted");
      } else {
        setSortBy(selectedSort);
      }
      setShowSortDropdown(false);
    },
    [sortBy],
  );

  // Get display text for current sort
  const getSortDisplayText = useCallback(() => {
    switch (sortBy) {
      case "unsorted":
        return "No Sorting";
      case "name-asc":
        return "Name (A-Z)";
      case "name-desc":
        return "Name (Z-A)";
      case "type-asc":
        return "Type (A-Z)";
      case "type-desc":
        return "Type (Z-A)";
      default:
        return "No Sorting";
    }
  }, [sortBy]);

  // Constants for virtual scrolling
  const CARD_HEIGHT = 420; // Approximate height of a card including gap
  const ROW_BUFFER = 10; // Rows to render above and below viewport

  // Calculate grid dimensions based on screen size
  const getCardsPerRow = useCallback(() => {
    const width = window.innerWidth;
    if (width <= 480) return 1;
    if (width <= 768) return 2;
    if (width <= 1100) return 3;
    if (width <= 1400) return 4;
    return 5;
  }, []);

  // Calculate visible range of cards
  const calculateVisibleRange = useCallback(() => {
    const scrollTop = window.scrollY;
    const viewportHeight = window.innerHeight;
    const headerOffset = showFilters ? 310 : 190; // Account for sticky header and spacer (expanded when filters visible)

    const cardsPerRow = getCardsPerRow();
    const adjustedScrollTop = Math.max(0, scrollTop - headerOffset);

    // Calculate which rows are visible
    const startRow = Math.floor(adjustedScrollTop / CARD_HEIGHT);
    const endRow = Math.ceil(
      (adjustedScrollTop + viewportHeight) / CARD_HEIGHT,
    );

    // Add buffer rows
    const bufferedStartRow = Math.max(0, startRow - ROW_BUFFER);
    const bufferedEndRow = endRow + ROW_BUFFER;

    // Convert to card indices
    const startIndex = bufferedStartRow * cardsPerRow;
    const endIndex = bufferedEndRow * cardsPerRow;

    return { start: startIndex, end: endIndex };
  }, [getCardsPerRow, showFilters]);

  // Scroll handler for sticky header and virtual scrolling
  const handleScroll = useCallback(() => {
    const scrollTop = window.scrollY;
    setIsScrolled(scrollTop > 50);

    const newRange = calculateVisibleRange();
    setVisibleRange(newRange);
  }, [calculateVisibleRange]);

  useEffect(() => {
    void loadAllCards();
    // Trigger fade-in animation
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  // Initialize state from URL on mount
  useEffect(() => {
    initializeFromURL();
  }, [initializeFromURL]);

  // Update URL when state changes
  useEffect(() => {
    updateURL();
  }, [updateURL]);

  // Apply filters when dependencies change
  useEffect(() => {
    if (allCards.length > 0) {
      applyFiltersAndSort();
    }
  }, [selectedTags, selectedTypes, sortBy, searchQuery, applyFiltersAndSort]);

  useEffect(() => {
    window.addEventListener("scroll", handleScroll);
    window.addEventListener("resize", handleScroll);
    // Initialize visible range
    handleScroll();
    return () => {
      window.removeEventListener("scroll", handleScroll);
      window.removeEventListener("resize", handleScroll);
    };
  }, [handleScroll]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest(".sort-dropdown-container")) {
        setShowSortDropdown(false);
      }
    };

    if (showSortDropdown) {
      document.addEventListener("click", handleClickOutside);
      return () => document.removeEventListener("click", handleClickOutside);
    }
    return undefined;
  }, [showSortDropdown]);

  const handleCardSelect = (cardId: string) => {
    setSelectedCards((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(cardId)) {
        newSet.delete(cardId);
      } else {
        newSet.add(cardId);
      }
      return newSet;
    });
  };

  const handleBackToHome = () => {
    navigate("/");
  };

  // Check if we're in permalink view mode
  const isPermalinkView = searchParams.getAll("cId").length > 0;

  const handleClearView = () => {
    // Clear all query params to return to normal view
    setSearchParams(new URLSearchParams(), { replace: true });
    setSelectedCards(new Set());
  };

  const generatePermalinkUrl = useCallback(() => {
    const params = new URLSearchParams();

    // Add selected card IDs as multiple cId parameters
    Array.from(selectedCards).forEach((cardId) => {
      params.append("cId", cardId);
    });

    return `${window.location.origin}${window.location.pathname}?${params.toString()}`;
  }, [selectedCards]);

  // Calculate total height for the container
  const totalHeight = useMemo(() => {
    const cardsPerRow = getCardsPerRow();
    const totalRows = Math.ceil(cards.length / cardsPerRow);
    return totalRows * CARD_HEIGHT;
  }, [cards.length, getCardsPerRow]);

  // Get only visible cards
  const visibleCards = useMemo(() => {
    return cards
      .slice(visibleRange.start, visibleRange.end)
      .map((card, index) => {
        const actualIndex = visibleRange.start + index;
        const cardsPerRow = getCardsPerRow();
        const row = Math.floor(actualIndex / cardsPerRow);
        const col = actualIndex % cardsPerRow;

        return {
          card,
          position: {
            row,
            col,
            top: row * CARD_HEIGHT,
            left: `${(col * 100) / cardsPerRow}%`,
            width: `${100 / cardsPerRow}%`,
          },
        };
      });
  }, [cards, visibleRange, getCardsPerRow]);

  return (
    <div
      className="cards-page"
      style={{
        opacity: isFadedIn ? 1 : 0,
        transition: "opacity 0.3s ease-in",
      }}
    >
      <div className={`sticky-header ${isScrolled ? "scrolled" : ""}`}>
        <div className="sticky-content">
          <button onClick={handleBackToHome} className="back-button">
            ← Back to Home
          </button>
          <h1>Terraforming Mars Cards</h1>
          <div className="right-section">
            <div className="cards-info-header">
              {cards.length > 0 && (
                <span>
                  {searchQuery ||
                  selectedTags.size > 0 ||
                  selectedTypes.size > 0
                    ? `${cards.length} of ${allCards.length}`
                    : `${cards.length}`}{" "}
                  cards
                </span>
              )}
            </div>
            {isPermalinkView && (
              <button
                className="clear-view-button"
                onClick={handleClearView}
                title="Clear permalink view and return to all cards"
              >
                Clear View
              </button>
            )}
            {selectedCards.size > 0 && !isPermalinkView && (
              <CopyLinkButton
                textToCopy={generatePermalinkUrl()}
                defaultText={`Link (${selectedCards.size})`}
                className="link-button"
              />
            )}
            <button
              className="filter-toggle-button"
              onClick={() => setShowFilters(!showFilters)}
            >
              <span className={`filter-arrow ${showFilters ? "open" : ""}`}>
                ▶
              </span>
              <span>Filters</span>
            </button>
            <div className="sort-dropdown-container">
              <button
                className="sort-dropdown-button"
                onClick={() => setShowSortDropdown(!showSortDropdown)}
              >
                <span className="sort-display-text">
                  {getSortDisplayText()}
                </span>
                <span
                  className={`dropdown-arrow ${showSortDropdown ? "open" : ""}`}
                >
                  ▼
                </span>
              </button>
              {showSortDropdown && (
                <div className="sort-dropdown-menu">
                  {[
                    { key: "name-asc", label: "Name (A-Z)" },
                    { key: "name-desc", label: "Name (Z-A)" },
                    { key: "type-asc", label: "Type (A-Z)" },
                    { key: "type-desc", label: "Type (Z-A)" },
                  ].map((option) => (
                    <button
                      key={option.key}
                      className={`sort-option ${sortBy === option.key ? "active" : ""}`}
                      onClick={() =>
                        handleSortSelection(option.key as typeof sortBy)
                      }
                    >
                      <span>{option.label}</span>
                      {sortBy === option.key && (
                        <span className="check-icon">✓</span>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="search-container">
              <input
                type="text"
                placeholder="Search cards (ID, name, tags)..."
                value={searchQuery}
                onChange={(e) => handleSearch(e.target.value)}
                className="search-input"
              />
            </div>
          </div>
        </div>
        {showFilters && (
          <div className="filters-panel">
            <div className="filter-section">
              <h3>Tags</h3>
              <div className="filter-chips">
                {availableTags.map((tag) => (
                  <button
                    key={tag}
                    className={`filter-chip tag-chip ${selectedTags.has(tag) ? "active" : ""}`}
                    onClick={() => toggleTag(tag)}
                    title={tag}
                  >
                    <GameIcon iconType={tag} size="small" />
                  </button>
                ))}
              </div>
            </div>
            <div className="filter-section">
              <h3>Card Types</h3>
              <div className="filter-chips">
                {availableTypes.map((type) => (
                  <button
                    key={type}
                    className={`filter-chip type-chip ${selectedTypes.has(type) ? "active" : ""}`}
                    onClick={() => toggleType(type)}
                    style={{
                      borderColor: getCardTypeColor(type),
                      backgroundColor: selectedTypes.has(type)
                        ? getCardTypeColor(type)
                        : "rgba(255, 255, 255, 0.1)",
                      color: selectedTypes.has(type)
                        ? "#ffffff"
                        : getCardTypeColor(type),
                    }}
                  >
                    {type}
                  </button>
                ))}
              </div>
            </div>
            {(selectedTags.size > 0 || selectedTypes.size > 0) && (
              <button
                className="clear-filters-button"
                onClick={() => {
                  setSelectedTags(new Set());
                  setSelectedTypes(new Set());
                }}
              >
                Clear All Filters
              </button>
            )}
          </div>
        )}
      </div>

      <div className="container">
        <div className={`header-spacer ${showFilters ? "expanded" : ""}`}></div>

        <div className="content-header">
          {error && <div className="error-message">{error}</div>}
          <div className="divider"></div>
        </div>

        <div
          className="cards-virtual-container"
          style={{ height: `${totalHeight}px`, position: "relative" }}
        >
          {visibleCards.map(({ card, position }) => (
            <div
              key={card.id}
              className="card-wrapper"
              style={{
                position: "absolute",
                top: `${position.top}px`,
                left: position.left,
                width: position.width,
                padding: "0 10px",
                height: `${CARD_HEIGHT}px`,
              }}
            >
              <SimpleGameCard
                card={card}
                isSelected={selectedCards.has(card.id)}
                onSelect={handleCardSelect}
                animationDelay={0}
                showCheckbox={!isPermalinkView}
              />
            </div>
          ))}
        </div>

        {loading && <div className="loading-message">Loading all cards...</div>}
      </div>

      <style>{`
        .cards-page {
          background: #000000;
          color: white;
          min-height: 100vh;
          font-family:
            -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
            "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
            sans-serif;
          position: relative;
          z-index: 10;
        }

        .sticky-header {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          z-index: 1000;
          transition: all 0.3s ease;
          padding: 0;
        }

        .sticky-header.scrolled {
          background: rgba(5, 5, 10, 0.98);
          backdrop-filter: blur(10px);
          border-bottom: 1px solid rgba(30, 60, 150, 0.2);
          box-shadow: 0 2px 20px rgba(0, 0, 0, 0.5);
        }

        .sticky-content {
          max-width: 1400px;
          margin: 0 auto;
          padding: 20px;
          display: flex;
          align-items: center;
          justify-content: space-between;
          flex-wrap: wrap;
          gap: 20px;
        }

        .container {
          max-width: 1400px;
          margin: 0 auto;
          padding: 40px 20px;
        }

        .header-spacer {
          height: 100px;
          transition: height 0.3s ease;
        }

        .header-spacer.expanded {
          height: 220px;
        }

        .content-header {
          text-align: center;
          margin-bottom: 20px;
        }

        .back-button {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 12px 20px;
          color: white;
          cursor: pointer;
          transition: all 0.3s ease;
          font-size: 14px;
          backdrop-filter: blur(10px);
        }

        .back-button:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.6);
          transform: translateY(-2px);
        }

        .link-button {
          background: linear-gradient(
            135deg,
            rgba(76, 175, 80, 0.2) 0%,
            rgba(76, 175, 80, 0.3) 50%,
            rgba(76, 175, 80, 0.2) 100%
          );
          border: 2px solid rgba(76, 175, 80, 0.5);
          border-radius: 8px;
          padding: 12px 20px;
          color: white;
          cursor: pointer;
          transition: all 0.3s ease;
          font-size: 14px;
          backdrop-filter: blur(10px);
          font-weight: 600;
        }

        .link-button:hover {
          border-color: rgba(76, 175, 80, 0.8);
          box-shadow: 0 4px 15px rgba(76, 175, 80, 0.3);
          transform: translateY(-2px);
          background: linear-gradient(
            135deg,
            rgba(76, 175, 80, 0.3) 0%,
            rgba(76, 175, 80, 0.4) 50%,
            rgba(76, 175, 80, 0.3) 100%
          );
        }

        .clear-view-button {
          background: linear-gradient(
            135deg,
            rgba(255, 107, 107, 0.2) 0%,
            rgba(255, 107, 107, 0.3) 50%,
            rgba(255, 107, 107, 0.2) 100%
          );
          border: 2px solid rgba(255, 107, 107, 0.5);
          border-radius: 8px;
          padding: 12px 20px;
          color: white;
          cursor: pointer;
          transition: all 0.3s ease;
          font-size: 14px;
          backdrop-filter: blur(10px);
          font-weight: 600;
        }

        .clear-view-button:hover {
          border-color: rgba(255, 107, 107, 0.8);
          box-shadow: 0 4px 15px rgba(255, 107, 107, 0.3);
          transform: translateY(-2px);
          background: linear-gradient(
            135deg,
            rgba(255, 107, 107, 0.3) 0%,
            rgba(255, 107, 107, 0.4) 50%,
            rgba(255, 107, 107, 0.3) 100%
          );
        }

        h1 {
          font-family: 'Orbitron', sans-serif;
          font-size: 20px;
          color: #ffffff;
          margin: 0;
          text-shadow: 0 0 30px rgba(30, 60, 150, 0.6);
          font-weight: bold;
          flex: 1;
          text-align: center;
          letter-spacing: 2px;
        }

        .right-section {
          display: flex;
          align-items: center;
          gap: 20px;
        }

        .cards-info-header {
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
          white-space: nowrap;
        }

        .search-container {
          display: flex;
          align-items: center;
        }

        .search-input {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 12px 16px;
          color: white;
          font-size: 14px;
          width: 300px;
          backdrop-filter: blur(10px);
          transition: all 0.3s ease;
        }

        .search-input::placeholder {
          color: rgba(255, 255, 255, 0.5);
        }

        .search-input:focus {
          outline: none;
          border-color: rgba(30, 60, 150, 0.8);
          background: rgba(10, 10, 15, 0.9);
          box-shadow: 0 0 25px rgba(30, 60, 150, 0.5);
        }

        .filter-toggle-button {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 12px 16px;
          color: white;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .filter-toggle-button:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.6);
        }

        .filter-arrow {
          transition: transform 0.3s ease;
          font-size: 12px;
        }

        .filter-arrow.open {
          transform: rotate(90deg);
        }

        .sort-dropdown-container {
          position: relative;
        }

        .sort-dropdown-button {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          padding: 12px 16px;
          color: white;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          display: flex;
          align-items: center;
          gap: 8px;
          min-width: 140px;
          justify-content: space-between;
        }

        .sort-dropdown-button:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.6);
        }

        .sort-display-text {
          font-weight: 500;
        }

        .dropdown-arrow {
          transition: transform 0.2s ease;
          font-size: 12px;
        }

        .dropdown-arrow.open {
          transform: rotate(180deg);
        }

        .sort-dropdown-menu {
          position: absolute;
          top: 100%;
          left: 0;
          right: 0;
          background: rgba(5, 5, 10, 0.98);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          backdrop-filter: blur(10px);
          z-index: 1000;
          margin-top: 4px;
          box-shadow: 0 4px 20px rgba(30, 60, 150, 0.3);
        }

        .sort-option {
          width: 100%;
          background: none;
          border: none;
          padding: 12px 16px;
          color: white;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          align-items: center;
          justify-content: space-between;
          text-align: left;
        }

        .sort-option:hover {
          background: rgba(255, 255, 255, 0.1);
        }

        .sort-option.active {
          background: rgba(30, 60, 150, 0.2);
          color: rgba(30, 60, 150, 1);
        }

        .check-icon {
          color: rgba(30, 60, 150, 1);
          font-weight: bold;
        }

        .filters-panel {
          background: rgba(5, 5, 10, 0.98);
          backdrop-filter: blur(10px);
          border-bottom: 1px solid rgba(30, 60, 150, 0.2);
          padding: 20px;
          max-width: 1400px;
          margin: 0 auto;
          animation: slideDown 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards;
          transform-origin: top;
        }

        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-20px) scaleY(0.8);
          }
          to {
            opacity: 1;
            transform: translateY(0) scaleY(1);
          }
        }

        .filter-section {
          margin-bottom: 20px;
        }

        .filter-section:last-child {
          margin-bottom: 0;
        }

        .filter-section h3 {
          color: white;
          font-size: 16px;
          margin: 0 0 12px 0;
          font-weight: 600;
        }

        .filter-chips {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
          justify-content: center;
        }

        .filter-chip {
          background: rgba(10, 10, 15, 0.8);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 20px;
          padding: 6px 12px;
          color: white;
          font-size: 12px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(5px);
        }

        .filter-chip:hover {
          background: rgba(10, 10, 15, 0.9);
          border-color: rgba(30, 60, 150, 0.8);
          box-shadow: 0 0 15px rgba(30, 60, 150, 0.4);
        }

        .filter-chip.active {
          background: rgba(30, 60, 150, 0.3);
          border-color: rgba(30, 60, 150, 0.8);
          color: white;
          box-shadow: 0 0 20px rgba(30, 60, 150, 0.5);
        }

        .clear-filters-button {
          background: rgba(255, 107, 107, 0.2);
          border: 1px solid rgba(255, 107, 107, 0.5);
          border-radius: 8px;
          padding: 8px 16px;
          color: #ff6b6b;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.3s ease;
          margin-top: 16px;
        }

        .clear-filters-button:hover {
          background: rgba(255, 107, 107, 0.3);
          border-color: rgba(255, 107, 107, 0.8);
        }

        .tag-chip {
          display: flex;
          align-items: center;
          justify-content: center;
          min-width: 40px;
          min-height: 32px;
          padding: 6px 8px;
        }

        .type-chip {
          font-weight: 600;
          text-transform: capitalize;
          transition: all 0.3s ease;
        }

        .type-chip:hover {
          transform: scale(1.05);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        }


        .divider {
          width: 100%;
          height: 1px;
          background: linear-gradient(
            90deg,
            transparent 0%,
            rgba(255, 255, 255, 0.2) 20%,
            rgba(255, 255, 255, 0.4) 50%,
            rgba(255, 255, 255, 0.2) 80%,
            transparent 100%
          );
          margin: 10px 0 40px 0;
        }

        .error-message {
          color: #ff6b6b;
          margin-bottom: 40px;
          padding: 16px;
          background: rgba(255, 107, 107, 0.1);
          border: 1px solid rgba(255, 107, 107, 0.3);
          border-radius: 8px;
          font-size: 14px;
          text-align: center;
        }

        .cards-virtual-container {
          max-width: 1300px;
          margin: 0 auto 40px auto;
          position: relative;
        }

        .card-wrapper {
          box-sizing: border-box;
        }

        .loading-message {
          text-align: center;
          padding: 20px;
          color: rgba(30, 60, 150, 1);
          font-size: 16px;
          background: rgba(30, 60, 150, 0.15);
          border: 2px solid rgba(30, 60, 150, 0.4);
          border-radius: 8px;
          animation: pulse 2s infinite;
          box-shadow: 0 0 15px rgba(30, 60, 150, 0.3);
        }


        @keyframes pulse {
          0%, 100% {
            opacity: 1;
          }
          50% {
            opacity: 0.7;
          }
        }


        @media (max-width: 768px) {
          .sticky-content {
            flex-direction: column;
            text-align: center;
            gap: 15px;
          }

          .right-section {
            flex-direction: column;
            width: 100%;
            gap: 15px;
          }

          h1 {
            font-size: 28px;
          }

          .search-input {
            width: 100%;
            max-width: 300px;
          }

          .container {
            padding: 20px 15px;
          }

          .header-spacer.expanded {
            height: 360px;
          }

          .filters-panel {
            padding: 15px;
          }

          .filter-chips {
            gap: 6px;
          }

          .filter-chip {
            font-size: 11px;
            padding: 5px 10px;
          }
        }

        @media (max-width: 480px) {
          h1 {
            font-size: 24px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardsPage;
