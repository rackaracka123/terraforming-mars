import React, { useState, useEffect, useRef, useCallback } from "react";
// Z-index imports removed - using DOM order and isolation for layering

interface HearthstoneCard {
  id: string;
  name: string;
  cost: number;
  type: "spell" | "minion" | "weapon";
  attack?: number;
  health?: number;
  description: string;
  rarity: "common" | "rare" | "epic" | "legendary";
  playable: boolean;
}

interface CardsHandOverlayProps {
  hideWhenModalOpen?: boolean;
}

const CardsHandOverlay: React.FC<CardsHandOverlayProps> = ({
  hideWhenModalOpen = false,
}) => {
  const [highlightedCard, setHighlightedCard] = useState<string | null>(null);
  const [draggedCard, setDraggedCard] = useState<string | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 }); // Offset from initial click to card position
  const [dragPosition, setDragPosition] = useState({ x: 0, y: 0 }); // Current mouse position
  const [isDragging, setIsDragging] = useState(false);
  const [justDragged, setJustDragged] = useState(false); // Track if we just finished dragging
  const [hoveredCard, setHoveredCard] = useState<string | null>(null); // Track which card is hovered for neighbor spreading
  const [cardScales, setCardScales] = useState<Record<string, number>>({}); // Track each card's scale
  const [cardRotations, setCardRotations] = useState<Record<string, number>>(
    {},
  ); // Track each card's rotation
  const [cardsExpanded, setCardsExpanded] = useState(false); // Track if cards are expanded/popped up
  const handRef = useRef<HTMLDivElement>(null);

  // Mock Hearthstone cards data
  const hearthstoneCards: HearthstoneCard[] = [
    {
      id: "1",
      name: "Fireball",
      cost: 4,
      type: "spell",
      description: "Deal 6 damage.",
      rarity: "common",
      playable: true,
    },
    {
      id: "2",
      name: "Chillwind Yeti",
      cost: 4,
      type: "minion",
      attack: 4,
      health: 5,
      description: "A solid minion.",
      rarity: "common",
      playable: true,
    },
    {
      id: "3",
      name: "Archmage Antonidas",
      cost: 7,
      type: "minion",
      attack: 5,
      health: 7,
      description: "Whenever you cast a spell, add a Fireball to your hand.",
      rarity: "legendary",
      playable: false,
    },
    {
      id: "4",
      name: "Lightning Bolt",
      cost: 1,
      type: "spell",
      description: "Deal 3 damage. Overload: (1)",
      rarity: "common",
      playable: true,
    },
    {
      id: "5",
      name: "Boulderfist Ogre",
      cost: 6,
      type: "minion",
      attack: 6,
      health: 7,
      description: "A big, dumb creature.",
      rarity: "common",
      playable: true,
    },
    {
      id: "6",
      name: "Polymorph",
      cost: 4,
      type: "spell",
      description: "Transform a minion into a 1/1 Sheep.",
      rarity: "common",
      playable: true,
    },
    {
      id: "7",
      name: "Ysera",
      cost: 9,
      type: "minion",
      attack: 4,
      health: 12,
      description: "At the end of your turn, add a Dream Card to your hand.",
      rarity: "legendary",
      playable: false,
    },
  ];

  // Calculate card positions with neighbor spreading for hovered or highlighted cards
  const calculateCardPosition = (
    index: number,
    totalCards: number,
    spreadIndex: number | null = null,
    expanded: boolean = false,
  ) => {
    const handWidth = expanded ? 800 : 400; // Much wider spread when expanded
    const handCurve = expanded ? 0.15 : 0.3; // Even less curve when expanded for better visibility
    const cardWidth = 120;
    const baseY = -20; // Base Y position

    // Base spacing - much wider when expanded
    const baseSpacing = expanded ? cardWidth * 1.2 : cardWidth * 0.4;
    const spacing = Math.min(
      baseSpacing,
      handWidth / Math.max(totalCards - 1, 1),
    );

    // Calculate neighbor spreading offset
    let spreadOffset = 0;
    if (spreadIndex !== null) {
      const distanceFromSpread = Math.abs(index - spreadIndex);
      if (distanceFromSpread === 1) {
        // Adjacent cards spread out more
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 40; // Much more spread for better visibility
      } else if (distanceFromSpread === 2) {
        // Second neighbors spread more
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 20; // More spread for second neighbors
      } else if (distanceFromSpread === 3) {
        // Third neighbors spread slightly
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 10; // Slight spread for third neighbors
      }
    }

    // Center the hand
    const totalWidth = spacing * (totalCards - 1);
    const startX = -totalWidth / 2;
    const x = startX + index * spacing + spreadOffset;

    // Create arc curve
    const normalizedX = x / (handWidth / 2); // -1 to 1
    const curveY = Math.pow(Math.abs(normalizedX), 2) * handCurve * 60;
    const y = baseY + curveY;

    // Compact rotation
    const rotation = normalizedX * 8; // Max 8 degrees

    return { x, y, rotation };
  };

  const getCardTypeColor = (type: string) => {
    switch (type) {
      case "spell":
        return "#8B5BFF";
      case "minion":
        return "#FFB800";
      case "weapon":
        return "#FF6B6B";
      default:
        return "#95a5a6";
    }
  };

  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case "common":
        return "#ffffff";
      case "rare":
        return "#0099cc";
      case "epic":
        return "#cc00ff";
      case "legendary":
        return "#ff8000";
      default:
        return "#ffffff";
    }
  };

  // Helper functions to manage card scales
  const getCardScale = (cardId: string) => {
    return cardScales[cardId] || 1; // Default scale is 1
  };

  const setCardScale = (cardId: string, scale: number) => {
    setCardScales((prev) => ({ ...prev, [cardId]: scale }));
  };

  const resetCardScale = useCallback((cardId: string) => {
    setCardScales((prev) => {
      const newScales = { ...prev };
      delete newScales[cardId];
      return newScales;
    });
  }, []);

  // Helper functions to manage card rotations
  const getCardRotation = (cardId: string, defaultRotation: number) => {
    return cardRotations[cardId] !== undefined
      ? cardRotations[cardId]
      : defaultRotation;
  };

  const setCardRotation = (cardId: string, rotation: number) => {
    setCardRotations((prev) => ({ ...prev, [cardId]: rotation }));
  };

  const resetCardRotation = useCallback((cardId: string) => {
    setCardRotations((prev) => {
      const newRotations = { ...prev };
      delete newRotations[cardId];
      return newRotations;
    });
  }, []);

  // Handle card hover scale and rotation changes
  const handleCardHover = (cardId: string) => {
    setHoveredCard(cardId);
    if (cardsExpanded && (!highlightedCard || highlightedCard !== cardId)) {
      setCardScale(cardId, 1.4); // Hover scale only when expanded
      setCardRotation(cardId, 0); // Straighten card on hover
    }
  };

  const handleCardLeave = (cardId: string) => {
    setHoveredCard(null);
    if (!highlightedCard || highlightedCard !== cardId) {
      resetCardScale(cardId); // Reset to default scale
      resetCardRotation(cardId); // Reset to default rotation
    }
  };

  // Handle card click (highlight and expand cards)
  const handleCardClick = (cardId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    if (isDragging || justDragged) return; // Ignore clicks right after dragging

    // Don't allow interaction with unplayable cards
    const card = hearthstoneCards.find((c) => c.id === cardId);
    if (card && !card.playable) return;

    // If cards are collapsed, just expand them without selecting a specific card
    if (!cardsExpanded) {
      setCardsExpanded(true);
      return; // Don't proceed to card selection when expanding
    }

    // Handle card selection when expanded
    if (highlightedCard === cardId) {
      setHighlightedCard(null);
      resetCardScale(cardId); // Reset scale when deselecting
      resetCardRotation(cardId); // Reset rotation when deselecting
    } else {
      setHighlightedCard(cardId);
      setCardScale(cardId, 1.6); // Selected scale
      setCardRotation(cardId, 0); // Straighten card when selected
    }
  };

  // Handle drag start
  const handleDragStart = (cardId: string, event: React.MouseEvent) => {
    event.preventDefault(); // Prevent browser default drag behavior

    // Expand cards when drag starts if not already expanded
    if (!cardsExpanded) {
      setCardsExpanded(true);
    }

    // Calculate initial offset from mouse to card position
    const cardIndex = hearthstoneCards.findIndex((c) => c.id === cardId);
    const spreadCard = highlightedCard || hoveredCard;
    const spreadIndex = spreadCard
      ? hearthstoneCards.findIndex((c) => c.id === spreadCard)
      : null;
    const cardPosition = calculateCardPosition(
      cardIndex,
      hearthstoneCards.length,
      spreadIndex,
      cardsExpanded,
    );
    const containerRect = handRef.current?.getBoundingClientRect();

    if (containerRect) {
      // Adjust for collapsed state - if cards were collapsed, we need to account for the offset
      let adjustedY = cardPosition.y;
      if (!cardsExpanded) {
        adjustedY += 90; // Account for collapsed offset
      }

      // Current card screen position
      const cardScreenX =
        containerRect.left + containerRect.width / 2 + cardPosition.x;
      const cardScreenY = containerRect.bottom + adjustedY;

      // Store offset from mouse to card position
      setDragOffset({
        x: cardScreenX - event.clientX,
        y: cardScreenY - event.clientY,
      });
    }

    setDraggedCard(cardId);
    setIsDragging(true);
    setDragPosition({ x: event.clientX, y: event.clientY });
    setHighlightedCard(null); // Clear highlight when dragging
  };

  // Add document event listeners for drag and click outside
  useEffect(() => {
    // Handle drag end - moved inside useEffect to avoid dependency warning
    const handleDragEnd = () => {
      const draggedCardId = draggedCard;
      setDraggedCard(null);
      setIsDragging(false);
      setDragPosition({ x: 0, y: 0 });
      setDragOffset({ x: 0, y: 0 }); // Reset offset
      setHighlightedCard(null); // Deselect card when released

      // Reset the dragged card's scale and rotation
      if (draggedCardId) {
        resetCardScale(draggedCardId);
        resetCardRotation(draggedCardId);
      }

      setJustDragged(true);

      // Clear the justDragged flag after a short delay to allow clicks again
      setTimeout(() => {
        setJustDragged(false);
      }, 100);
    };

    const handleDocumentClick = (event: MouseEvent) => {
      if (handRef.current && !handRef.current.contains(event.target as Node)) {
        setHighlightedCard(null);
        setCardsExpanded(false);
      }
    };

    const handleDocumentMouseMove = (event: MouseEvent) => {
      if (isDragging && draggedCard) {
        setDragPosition({ x: event.clientX, y: event.clientY });
      }
    };

    const handleDocumentMouseUp = () => {
      if (isDragging && draggedCard) {
        handleDragEnd();
      }
    };

    document.addEventListener("click", handleDocumentClick);
    document.addEventListener("mousemove", handleDocumentMouseMove);
    document.addEventListener("mouseup", handleDocumentMouseUp);

    return () => {
      document.removeEventListener("click", handleDocumentClick);
      document.removeEventListener("mousemove", handleDocumentMouseMove);
      document.removeEventListener("mouseup", handleDocumentMouseUp);
    };
  }, [isDragging, draggedCard, resetCardScale, resetCardRotation]);

  // Hide the overlay when modals are open
  if (hideWhenModalOpen) {
    return null;
  }

  return (
    <div className="hearthstone-hand-overlay">
      <div
        className={`card-hand-container ${cardsExpanded ? "expanded" : ""}`}
        ref={handRef}
      >
        {hearthstoneCards.map((card, index) => {
          // Determine which card should cause spreading (highlighted takes priority over hovered)
          const spreadCard = highlightedCard || hoveredCard;
          const spreadIndex = spreadCard
            ? hearthstoneCards.findIndex((c) => c.id === spreadCard)
            : null;
          const position = calculateCardPosition(
            index,
            hearthstoneCards.length,
            spreadIndex,
            cardsExpanded,
          );
          const isHighlighted = highlightedCard === card.id;
          const isDraggedCard = draggedCard === card.id;
          const isHovered = hoveredCard === card.id;

          // Calculate final position and transform
          let finalX = position.x;
          let finalY = position.y;
          const finalRotation = getCardRotation(card.id, position.rotation); // Get the card's tracked rotation
          let scale = getCardScale(card.id); // Get the card's tracked scale
          // Card layering handled by DOM order and CSS isolation instead of z-index

          // Apply expanded state offset and scaling
          if (!cardsExpanded) {
            finalY += 130; // Push cards down to show only top portion (halfway off-screen)
            scale = 1.0; // Normal scale when collapsed
          } else {
            // When expanded, make cards larger and spread more
            scale = Math.max(scale, 1.5); // Minimum 1.5x scale when expanded
          }

          if (isHovered && !isDragging && !isHighlighted && cardsExpanded) {
            finalY -= 80; // Move up even more to show more of the card
            // Scale is managed by hover handlers
            // Card elevation handled by isolation and transform
          }

          if (isHighlighted && !isDragging && cardsExpanded) {
            finalY -= 100; // Pop out significantly from the fan
            // Scale is managed by click handlers
            // Card prominence handled by CSS isolation
          }

          if (isDraggedCard) {
            // Apply mouse position with offset to maintain relative position
            const containerRect = handRef.current?.getBoundingClientRect();
            if (containerRect) {
              // Apply mouse position + stored offset, then convert to container coordinates
              const targetScreenX = dragPosition.x + dragOffset.x;
              const targetScreenY = dragPosition.y + dragOffset.y;

              finalX =
                targetScreenX - (containerRect.left + containerRect.width / 2);
              finalY = targetScreenY - containerRect.bottom;
            }
            // Keep the current rotation and scale (don't change them when dragging)
            // Dragged card elevation handled by CSS isolation
          }

          return (
            <div
              key={card.id}
              className={`hearthstone-card ${isHighlighted ? "highlighted" : ""} ${isDraggedCard ? "dragged" : ""} ${isHovered ? "hovered" : ""} ${!card.playable ? "unplayable" : ""}`}
              style={
                {
                  transform: `translate(${finalX}px, ${finalY}px) rotate(${finalRotation}deg) scale(${scale})`,
                  // z-index removed - using isolation for card layering
                  "--card-type-color": getCardTypeColor(card.type),
                  "--rarity-color": getRarityColor(card.rarity),
                  "--card-x": `${finalX}px`,
                  "--card-y": `${finalY}px`,
                  "--card-rotation": `${finalRotation}deg`,
                  "--card-scale": scale,
                } as React.CSSProperties
              }
              onClick={(e) => handleCardClick(card.id, e)}
              onMouseDown={(e) => handleDragStart(card.id, e)}
              onMouseEnter={() => handleCardHover(card.id)}
              onMouseLeave={() => handleCardLeave(card.id)}
            >
              <div className="card-inner">
                <div className="card-cost">{card.cost}</div>
                <div className="card-header">
                  <div
                    className="card-name"
                    style={{ color: getRarityColor(card.rarity) }}
                  >
                    {card.name}
                  </div>
                </div>

                {card.type === "minion" && (
                  <div className="card-stats">
                    <div className="attack">{card.attack}</div>
                    <div className="health">{card.health}</div>
                  </div>
                )}

                <div className="card-description">{card.description}</div>

                <div className="card-type-indicator" />
              </div>
            </div>
          );
        })}
      </div>

      <style>{`
        .hearthstone-hand-overlay {
          position: fixed;
          bottom: 120px; /* Position above BottomResourceBar (120px height) to avoid overlap */
          left: 0;
          right: 0;
          height: 280px; /* Cards area height */
          /* z-index removed - natural DOM order places this appropriately */
          /* Positioned above BottomResourceBar so cards don't interfere with buttons */
          pointer-events: none;
        }

        .card-hand-container {
          position: absolute;
          bottom: 0px;
          left: 50%;
          transform: translateX(-50%);
          width: 100%;
          height: 200px;
          pointer-events: auto;
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        .card-hand-container.expanded {
          height: 280px; /* Increase height when expanded to accommodate larger cards */
        }

        .card-hand-container:not(.expanded) {
          cursor: pointer;
        }

        .card-hand-container:not(.expanded) .hearthstone-card {
          box-shadow: 
            0 0 15px rgba(100, 150, 255, 0.3),
            0 4px 20px rgba(0, 0, 0, 0.4);
        }

        .card-hand-container:not(.expanded):hover .hearthstone-card {
          transform: translate(var(--card-x), calc(var(--card-y) - 10px)) rotate(var(--card-rotation)) scale(var(--card-scale));
          box-shadow: 
            0 0 20px rgba(100, 150, 255, 0.5),
            0 6px 25px rgba(0, 0, 0, 0.5);
        }

        .hearthstone-card {
          position: absolute;
          width: 120px;
          height: 180px;
          bottom: 0;
          left: 50%;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
          transform-origin: bottom center;
          pointer-events: auto;
          user-select: none;
          /* Use isolation to create natural stacking contexts */
          isolation: isolate;
        }

        .hearthstone-card.hovered {
          box-shadow: 
            0 0 10px rgba(255, 255, 255, 0.2),
            0 4px 16px rgba(0, 0, 0, 0.3);
        }

        .hearthstone-card.highlighted {
          box-shadow: 
            0 0 20px var(--card-type-color),
            0 8px 32px rgba(0, 0, 0, 0.6);
        }

        .hearthstone-card.dragged {
          transition: none;
          cursor: grabbing;
        }

        .hearthstone-card:not(.dragged) {
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        .hearthstone-card.unplayable {
          opacity: 1.0;
          cursor: not-allowed;
          filter: grayscale(50%); /* Use grayscale instead of transparency for unplayable cards */
        }

        .card-inner {
          width: 100%;
          height: 100%;
          background: linear-gradient(
            135deg,
            rgb(20, 25, 40) 0%,
            rgb(35, 40, 55) 50%,
            rgb(25, 30, 45) 100%
          );
          border: 3px solid var(--card-type-color);
          border-radius: 12px;
          padding: 8px;
          position: relative;
          box-shadow: 
            0 4px 12px rgba(0, 0, 0, 0.5),
            inset 0 1px 0 rgba(255, 255, 255, 0.1);
          display: flex;
          flex-direction: column;
          pointer-events: none;
        }

        .card-cost {
          position: absolute;
          top: -2px;
          left: -2px;
          width: 32px;
          height: 32px;
          background: linear-gradient(135deg, #4a90e2 0%, #357abd 100%);
          color: white;
          font-size: 16px;
          font-weight: bold;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          border: 2px solid #ffffff;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
          pointer-events: none;
        }

        .card-header {
          margin-bottom: 8px;
          pointer-events: none;
        }

        .card-name {
          font-size: 11px;
          font-weight: bold;
          text-align: center;
          line-height: 1.2;
          color: var(--rarity-color);
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          pointer-events: none;
        }

        .card-stats {
          position: absolute;
          bottom: 4px;
          left: 4px;
          right: 4px;
          display: flex;
          justify-content: space-between;
          pointer-events: none;
        }

        .attack, .health {
          width: 24px;
          height: 24px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 12px;
          font-weight: bold;
          color: white;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          border: 2px solid #ffffff;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4);
        }

        .attack {
          background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
          pointer-events: none;
        }

        .health {
          background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
          pointer-events: none;
        }

        .card-description {
          flex: 1;
          font-size: 8px;
          color: rgb(255, 255, 255);
          text-align: center;
          line-height: 1.3;
          margin: 8px 0;
          padding: 4px;
          background: rgba(0, 0, 0, 0.3);
          border-radius: 4px;
          overflow: hidden;
          display: -webkit-box;
          -webkit-line-clamp: 6;
          -webkit-box-orient: vertical;
          pointer-events: none;
        }

        .card-type-indicator {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          height: 3px;
          background: var(--card-type-color);
          border-radius: 8px 8px 0 0;
          pointer-events: none;
        }

        /* Responsive Design */
        @media (max-width: 1200px) {
          .hearthstone-card {
            width: 100px;
            height: 150px;
          }

          .card-name {
            font-size: 10px;
          }

          .card-description {
            font-size: 7px;
            -webkit-line-clamp: 5;
          }

          .card-cost {
            width: 28px;
            height: 28px;
            font-size: 14px;
          }

          .attack, .health {
            width: 20px;
            height: 20px;
            font-size: 10px;
          }
        }

        @media (max-width: 768px) {
          .card-hand-container {
            bottom: 0px;
          }

          .hearthstone-card {
            width: 80px;
            height: 120px;
          }

          .card-name {
            font-size: 8px;
          }

          .card-description {
            font-size: 6px;
            -webkit-line-clamp: 4;
          }

          .card-cost {
            width: 24px;
            height: 24px;
            font-size: 12px;
          }

          .attack, .health {
            width: 18px;
            height: 18px;
            font-size: 9px;
          }
        }

        @media (max-width: 600px) {
          .hearthstone-card {
            width: 70px;
            height: 100px;
          }

          .card-inner {
            padding: 6px;
          }

          .card-name {
            font-size: 7px;
          }

          .card-description {
            font-size: 5px;
            -webkit-line-clamp: 3;
            margin: 4px 0;
          }

          .card-cost {
            width: 20px;
            height: 20px;
            font-size: 10px;
          }

          .attack, .health {
            width: 16px;
            height: 16px;
            font-size: 8px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardsHandOverlay;
