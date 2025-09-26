import React, { useState, useEffect, useRef, useCallback } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";

interface CardFanOverlayProps {
  cards: CardDto[];
  hideWhenModalOpen?: boolean;
  onCardSelect?: (cardId: string) => void;
  onPlayCard?: (cardId: string) => Promise<void>;
}

const CardFanOverlay: React.FC<CardFanOverlayProps> = ({
  cards,
  hideWhenModalOpen = false,
  onCardSelect,
  onPlayCard,
}) => {
  const [highlightedCard, setHighlightedCard] = useState<string | null>(null);
  const [draggedCard, setDraggedCard] = useState<string | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [dragPosition, setDragPosition] = useState({ x: 0, y: 0 });
  const [dragStartPosition, setDragStartPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [justDragged, setJustDragged] = useState(false);
  const [isInThrowZone, setIsInThrowZone] = useState(false);
  const [hoveredCard, setHoveredCard] = useState<string | null>(null);
  const [cardScales, setCardScales] = useState<Record<string, number>>({});
  const [cardRotations, setCardRotations] = useState<Record<string, number>>(
    {},
  );
  const [cardsExpanded, setCardsExpanded] = useState(false);
  const handRef = useRef<HTMLDivElement>(null);

  // Throw detection constants
  const THROW_DISTANCE_THRESHOLD = 120; // pixels - minimum distance to trigger throw
  const THROW_Y_THRESHOLD = -80; // pixels - minimum upward movement to trigger throw

  // Calculate card positions with neighbor spreading for hovered or highlighted cards
  const calculateCardPosition = (
    index: number,
    totalCards: number,
    spreadIndex: number | null = null,
    expanded: boolean = false,
  ) => {
    const handWidth = expanded ? 800 : 400;
    const handCurve = expanded ? 0.15 : 0.3;
    const cardWidth = 160; // Wider for SimpleGameCard
    const baseY = -20;

    // Base spacing - much wider when expanded
    const baseSpacing = expanded ? cardWidth * 0.8 : cardWidth * 0.3;
    const spacing = Math.min(
      baseSpacing,
      handWidth / Math.max(totalCards - 1, 1),
    );

    // Calculate neighbor spreading offset
    let spreadOffset = 0;
    if (spreadIndex !== null) {
      const distanceFromSpread = Math.abs(index - spreadIndex);
      if (distanceFromSpread === 1) {
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 40;
      } else if (distanceFromSpread === 2) {
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 20;
      } else if (distanceFromSpread === 3) {
        const direction = index > spreadIndex ? 1 : -1;
        spreadOffset = direction * 10;
      }
    }

    // Center the hand
    const totalWidth = spacing * (totalCards - 1);
    const startX = -totalWidth / 2;
    const x = startX + index * spacing + spreadOffset;

    // Create arc curve
    const normalizedX = x / (handWidth / 2);
    const curveY = Math.pow(Math.abs(normalizedX), 2) * handCurve * 60;
    const y = baseY + curveY;

    // Compact rotation
    const rotation = normalizedX * 8;

    return { x, y, rotation };
  };

  // Helper functions to manage card scales
  const getCardScale = (cardId: string) => {
    return cardScales[cardId] || 1;
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
      setCardScale(cardId, 1.2);
      setCardRotation(cardId, 0);
    }
  };

  const handleCardLeave = (cardId: string) => {
    setHoveredCard(null);
    if (!highlightedCard || highlightedCard !== cardId) {
      resetCardScale(cardId);
      resetCardRotation(cardId);
    }
  };

  // Handle card click (highlight and expand cards)
  const handleCardClick = (cardId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    if (isDragging || justDragged) return;

    // If cards are collapsed, just expand them without selecting a specific card
    if (!cardsExpanded) {
      setCardsExpanded(true);
      return;
    }

    // Handle card selection when expanded
    if (highlightedCard === cardId) {
      setHighlightedCard(null);
      resetCardScale(cardId);
      resetCardRotation(cardId);
    } else {
      setHighlightedCard(cardId);
      setCardScale(cardId, 1.4);
      setCardRotation(cardId, 0);
      // Call parent callback if provided
      onCardSelect?.(cardId);
    }
  };

  // Handle drag start
  const handleDragStart = (cardId: string, event: React.MouseEvent) => {
    event.preventDefault();

    // Expand cards when drag starts if not already expanded
    if (!cardsExpanded) {
      setCardsExpanded(true);
    }

    const cardIndex = cards.findIndex((c) => c.id === cardId);
    const spreadCard = highlightedCard || hoveredCard;
    const spreadIndex = spreadCard
      ? cards.findIndex((c) => c.id === spreadCard)
      : null;
    const cardPosition = calculateCardPosition(
      cardIndex,
      cards.length,
      spreadIndex,
      cardsExpanded,
    );
    const containerRect = handRef.current?.getBoundingClientRect();

    if (containerRect) {
      let adjustedY = cardPosition.y;
      if (!cardsExpanded) {
        adjustedY += 90;
      }

      const cardScreenX =
        containerRect.left + containerRect.width / 2 + cardPosition.x;
      const cardScreenY = containerRect.bottom + adjustedY;

      setDragOffset({
        x: cardScreenX - event.clientX,
        y: cardScreenY - event.clientY,
      });
    }

    setDraggedCard(cardId);
    setIsDragging(true);
    setDragPosition({ x: event.clientX, y: event.clientY });
    setDragStartPosition({ x: event.clientX, y: event.clientY });
    setHighlightedCard(null);
    setIsInThrowZone(false);
  };

  // Stable event handlers using useCallback
  const handleDragEnd = useCallback(async () => {
    const draggedCardId = draggedCard;

    // Calculate drag distance and direction for throw detection
    const deltaX = dragPosition.x - dragStartPosition.x;
    const deltaY = dragPosition.y - dragStartPosition.y;
    const dragDistance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
    const isUpwardThrow = deltaY < THROW_Y_THRESHOLD;
    const isThrowDetected =
      dragDistance > THROW_DISTANCE_THRESHOLD && isUpwardThrow;

    // Reset drag states
    setDraggedCard(null);
    setIsDragging(false);
    setDragPosition({ x: 0, y: 0 });
    setDragStartPosition({ x: 0, y: 0 });
    setDragOffset({ x: 0, y: 0 });
    setHighlightedCard(null);
    setIsInThrowZone(false);

    if (draggedCardId) {
      resetCardScale(draggedCardId);
      resetCardRotation(draggedCardId);

      // Handle throw action
      if (isThrowDetected && onPlayCard) {
        try {
          await onPlayCard(draggedCardId);
        } catch (error) {
          console.error("Failed to play card:", error);
          // Could add error feedback here
        }
      }
    }

    setJustDragged(true);
    setTimeout(() => {
      setJustDragged(false);
    }, 100);
  }, [
    draggedCard,
    dragPosition,
    dragStartPosition,
    onPlayCard,
    resetCardScale,
    resetCardRotation,
  ]);

  const handleDocumentClick = useCallback((event: MouseEvent) => {
    if (handRef.current && !handRef.current.contains(event.target as Node)) {
      setHighlightedCard(null);
      setCardsExpanded(false);
    }
  }, []);

  const handleDocumentMouseMove = useCallback(
    (event: MouseEvent) => {
      if (isDragging && draggedCard) {
        setDragPosition({ x: event.clientX, y: event.clientY });

        // Check if we're in throw zone for visual feedback
        const deltaX = event.clientX - dragStartPosition.x;
        const deltaY = event.clientY - dragStartPosition.y;
        const dragDistance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
        const isUpwardThrow = deltaY < THROW_Y_THRESHOLD;
        const inThrowZone =
          dragDistance > THROW_DISTANCE_THRESHOLD && isUpwardThrow;

        if (inThrowZone !== isInThrowZone) {
          setIsInThrowZone(inThrowZone);
        }
      }
    },
    [isDragging, draggedCard, dragStartPosition, isInThrowZone],
  );

  const handleDocumentMouseUp = useCallback(() => {
    if (isDragging && draggedCard) {
      handleDragEnd();
    }
  }, [isDragging, draggedCard, handleDragEnd]);

  // Add document event listeners for drag and click outside
  useEffect(() => {
    document.addEventListener("click", handleDocumentClick);
    document.addEventListener("mousemove", handleDocumentMouseMove);
    document.addEventListener("mouseup", handleDocumentMouseUp);

    return () => {
      document.removeEventListener("click", handleDocumentClick);
      document.removeEventListener("mousemove", handleDocumentMouseMove);
      document.removeEventListener("mouseup", handleDocumentMouseUp);
    };
  }, [handleDocumentClick, handleDocumentMouseMove, handleDocumentMouseUp]);

  // Hide the overlay when modals are open or no cards
  if (hideWhenModalOpen || cards.length === 0) {
    return null;
  }

  return (
    <div className="card-fan-overlay">
      <div
        className={`card-hand-container ${cardsExpanded ? "expanded" : ""}`}
        ref={handRef}
      >
        {cards.map((card, index) => {
          const spreadCard = highlightedCard || hoveredCard;
          const spreadIndex = spreadCard
            ? cards.findIndex((c) => c.id === spreadCard)
            : null;
          const position = calculateCardPosition(
            index,
            cards.length,
            spreadIndex,
            cardsExpanded,
          );
          const isHighlighted = highlightedCard === card.id;
          const isDraggedCard = draggedCard === card.id;
          const isHovered = hoveredCard === card.id;

          let finalX = position.x;
          let finalY = position.y;
          const finalRotation = getCardRotation(card.id, position.rotation);
          let scale = getCardScale(card.id);

          // Apply expanded state offset and scaling
          if (!cardsExpanded) {
            finalY += 160; // Push cards down to show only top portion
            scale = 0.8; // Smaller scale when collapsed
          } else {
            scale = Math.max(scale, 1.0); // Normal scale when expanded
          }

          if (isHovered && !isDragging && !isHighlighted && cardsExpanded) {
            finalY -= 60;
          }

          if (isHighlighted && !isDragging && cardsExpanded) {
            finalY -= 80;
          }

          if (isDraggedCard) {
            const containerRect = handRef.current?.getBoundingClientRect();
            if (containerRect) {
              const targetScreenX = dragPosition.x + dragOffset.x;
              const targetScreenY = dragPosition.y + dragOffset.y;

              finalX =
                targetScreenX - (containerRect.left + containerRect.width / 2);
              finalY = targetScreenY - containerRect.bottom;
            }
          }

          return (
            <div
              key={card.id}
              className={`terraforming-card ${isHighlighted ? "highlighted" : ""} ${isDraggedCard ? "dragged" : ""} ${isHovered ? "hovered" : ""} ${isDraggedCard && isInThrowZone ? "throw-zone" : ""}`}
              style={
                {
                  transform: `translate(${finalX}px, ${finalY}px) rotate(${finalRotation}deg) scale(${scale})`,
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
              <SimpleGameCard
                card={card}
                isSelected={isHighlighted}
                onSelect={() => {}} // Handled by parent div click
                animationDelay={0}
              />
            </div>
          );
        })}
      </div>

      <style>{`
        .card-fan-overlay {
          position: fixed;
          bottom: 48px; /* Position above BottomResourceBar */
          left: 0;
          right: 0;
          height: 300px; /* Cards area height */
          z-index: 1100; /* Above bottom bar (1000) and its content (1001) */
          pointer-events: none;
        }

        .card-hand-container {
          position: absolute;
          bottom: 48px; /* Position above BottomResourceBar (48px height) */
          left: 50%;
          transform: translateX(-50%);
          width: 100%;
          height: 240px;
          pointer-events: auto;
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        .card-hand-container.expanded {
          height: 300px;
        }

        .card-hand-container:not(.expanded) {
          cursor: pointer;
        }

        .card-hand-container:not(.expanded) .terraforming-card {
          box-shadow:
            0 0 15px rgba(100, 150, 255, 0.3),
            0 4px 20px rgba(0, 0, 0, 0.4);
        }

        .card-hand-container:not(.expanded):hover .terraforming-card {
          transform: translate(var(--card-x), calc(var(--card-y) - 10px)) rotate(var(--card-rotation)) scale(var(--card-scale));
          box-shadow:
            0 0 20px rgba(100, 150, 255, 0.5),
            0 6px 25px rgba(0, 0, 0, 0.5);
        }

        .terraforming-card {
          position: absolute;
          bottom: 0;
          left: 50%;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
          transform-origin: bottom center;
          pointer-events: auto;
          user-select: none;
          isolation: isolate;
        }

        .terraforming-card.hovered {
          box-shadow:
            0 0 10px rgba(255, 255, 255, 0.2),
            0 4px 16px rgba(0, 0, 0, 0.3);
        }

        .terraforming-card.highlighted {
          box-shadow:
            0 0 20px rgba(255, 165, 0, 0.8),
            0 8px 32px rgba(0, 0, 0, 0.6);
        }

        .terraforming-card.dragged {
          transition: none;
          cursor: grabbing;
          z-index: 1000;
        }

        .terraforming-card.throw-zone {
          box-shadow:
            0 0 30px rgba(0, 255, 100, 0.8),
            0 8px 40px rgba(0, 255, 100, 0.4);
          filter: brightness(1.2);
        }

        .terraforming-card:not(.dragged) {
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        /* Responsive Design */
        @media (max-width: 1200px) {
          .card-fan-overlay {
            height: 250px;
          }

          .card-hand-container {
            height: 200px;
          }

          .card-hand-container.expanded {
            height: 250px;
          }
        }

        @media (max-width: 768px) {
          .card-fan-overlay {
            height: 200px;
          }

          .card-hand-container {
            height: 160px;
          }

          .card-hand-container.expanded {
            height: 200px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardFanOverlay;
