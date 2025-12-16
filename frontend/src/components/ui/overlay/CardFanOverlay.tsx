import React, { useState, useEffect, useRef, useCallback } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import { PlayerCardDto, StateErrorDto } from "@/types/generated/api-types.ts";

/**
 * Convert StateErrorDto to a user-friendly message
 */
function formatErrorMessage(errors: StateErrorDto[]): string {
  if (errors.length === 0) return "";
  if (errors.length === 1) return errors[0].message;
  return `${errors[0].message} (+${errors.length - 1} more)`;
}

interface CardFanOverlayProps {
  cards: PlayerCardDto[];
  hideWhenModalOpen?: boolean;
  onCardSelect?: (cardId: string) => void;
  onPlayCard?: (cardId: string) => Promise<void>;
  onUnplayableCard?: (
    card: PlayerCardDto | null,
    errorMessage: string | null,
  ) => void;
}

const CardFanOverlay: React.FC<CardFanOverlayProps> = ({
  cards,
  hideWhenModalOpen = false,
  onCardSelect,
  onPlayCard,
  onUnplayableCard,
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
  const [isHoveringMars, setIsHoveringMars] = useState(false);
  const [returningCard, setReturningCard] = useState<string | null>(null);
  const handRef = useRef<HTMLDivElement>(null);
  const cardsRef = useRef(cards);

  // Update refs when props change
  useEffect(() => {
    cardsRef.current = cards;
  }, [cards]);

  // Throw detection constants
  const THROW_DISTANCE_THRESHOLD = 120; // pixels - minimum distance to trigger throw
  const THROW_Y_THRESHOLD = -80; // pixels - minimum upward movement to trigger throw

  // Mars hover detection - define the Mars play area (center portion of screen)
  const isCursorOverMars = (x: number, y: number): boolean => {
    const centerX = window.innerWidth / 2;

    // Define Mars area as the center portion of the screen
    // Exclude bottom area where cards are (bottom 300px) and top area (top 100px)
    const marsAreaWidth = window.innerWidth * 0.8; // 80% of screen width
    const marsAreaHeight = window.innerHeight - 400; // Exclude top 100px and bottom 300px

    const leftBound = centerX - marsAreaWidth / 2;
    const rightBound = centerX + marsAreaWidth / 2;
    const topBound = 100; // Top margin
    const bottomBound = topBound + marsAreaHeight;

    return (
      x >= leftBound && x <= rightBound && y >= topBound && y <= bottomBound
    );
  };

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
    if (!highlightedCard || highlightedCard !== cardId) {
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

  // Handle card click (highlight card)
  const handleCardClick = (cardId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    if (isDragging || justDragged) return;

    // Toggle card selection
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

    // Always allow drag to start - we'll check playability when hovering over Mars
    // Clear hover state when drag starts
    setHoveredCard(null);

    const cardIndex = cards.findIndex((c) => c.id === cardId);
    const spreadCard = highlightedCard || hoveredCard;
    const spreadIndex = spreadCard
      ? cards.findIndex((c) => c.id === spreadCard)
      : null;
    const cardPosition = calculateCardPosition(
      cardIndex,
      cards.length,
      spreadIndex,
      false,
    );
    const containerRect = handRef.current?.getBoundingClientRect();

    if (containerRect) {
      let adjustedY = cardPosition.y;
      // Always add offset for collapsed state
      adjustedY += 90;

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

    if (draggedCardId) {
      // Handle throw action first - but only if card is playable
      if (isThrowDetected && onPlayCard) {
        // Check if card is playable before attempting to play it
        const draggedCardData = cardsRef.current.find(
          (c) => c.id === draggedCardId,
        );

        if (draggedCardData?.available) {
          // Card is playable, proceed with playing it
          try {
            await onPlayCard(draggedCardId);
            // If card is played successfully, clear drag states immediately
            setDraggedCard(null);
            setIsDragging(false);
            setDragPosition({ x: 0, y: 0 });
            setDragStartPosition({ x: 0, y: 0 });
            setDragOffset({ x: 0, y: 0 });
            setHighlightedCard(null);
            setIsInThrowZone(false);
            setIsHoveringMars(false);
            resetCardScale(draggedCardId);
            resetCardRotation(draggedCardId);
          } catch (error) {
            console.error("Failed to play card:", error);
            // Card failed to play, animate return to hand
            setReturningCard(draggedCardId);
            setIsDragging(false);
            setIsInThrowZone(false);
            setIsHoveringMars(false);
            setHoveredCard(null);

            // After animation completes, reset all states
            setTimeout(() => {
              setDraggedCard(null);
              setDragPosition({ x: 0, y: 0 });
              setDragStartPosition({ x: 0, y: 0 });
              setDragOffset({ x: 0, y: 0 });
              setHighlightedCard(null);
              setReturningCard(null);
              resetCardScale(draggedCardId);
              resetCardRotation(draggedCardId);
            }, 400); // Match CSS transition duration
          }
        } else {
          // Card is not playable, don't send to backend - animate return to hand
          // Card is not playable, blocking play attempt
          setReturningCard(draggedCardId);
          setIsDragging(false);
          setIsInThrowZone(false);
          setIsHoveringMars(false);
          setHoveredCard(null);

          // After animation completes, reset all states
          setTimeout(() => {
            setDraggedCard(null);
            setDragPosition({ x: 0, y: 0 });
            setDragStartPosition({ x: 0, y: 0 });
            setDragOffset({ x: 0, y: 0 });
            setHighlightedCard(null);
            setReturningCard(null);
            resetCardScale(draggedCardId);
            resetCardRotation(draggedCardId);
          }, 400); // Match CSS transition duration
        }
      } else {
        // No throw detected, animate return to hand
        setReturningCard(draggedCardId);
        setIsDragging(false);
        setIsInThrowZone(false);
        setIsHoveringMars(false);
        setHoveredCard(null);

        // After animation completes, reset all states
        setTimeout(() => {
          setDraggedCard(null);
          setDragPosition({ x: 0, y: 0 });
          setDragStartPosition({ x: 0, y: 0 });
          setDragOffset({ x: 0, y: 0 });
          setHighlightedCard(null);
          setReturningCard(null);
          resetCardScale(draggedCardId);
          resetCardRotation(draggedCardId);
        }, 400); // Match CSS transition duration
      }
    } else {
      // No dragged card, clear states immediately
      setDraggedCard(null);
      setIsDragging(false);
      setDragPosition({ x: 0, y: 0 });
      setDragStartPosition({ x: 0, y: 0 });
      setDragOffset({ x: 0, y: 0 });
      setHighlightedCard(null);
      setIsInThrowZone(false);
      setIsHoveringMars(false);
    }

    // Clear unplayable card feedback
    if (onUnplayableCard) {
      onUnplayableCard(null, null);
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
    }
  }, []);

  const handleDocumentMouseMove = useCallback(
    (event: MouseEvent) => {
      if (isDragging && draggedCard) {
        setDragPosition({ x: event.clientX, y: event.clientY });

        // Check if we're hovering over Mars
        const hoveringMars = isCursorOverMars(event.clientX, event.clientY);
        if (hoveringMars !== isHoveringMars) {
          setIsHoveringMars(hoveringMars);
        }

        // Update feedback based on Mars hover state (do this every time, not just on state change)
        if (hoveringMars && onUnplayableCard && draggedCard) {
          // Get current card data using refs to avoid dependency issues
          const currentCard = cardsRef.current.find(
            (c) => c.id === draggedCard,
          );

          if (
            currentCard &&
            !currentCard.available &&
            currentCard.errors.length > 0
          ) {
            // Card is not playable, show error message
            const errorMessage = formatErrorMessage(currentCard.errors);
            onUnplayableCard(currentCard, errorMessage);
          } else if (currentCard?.available) {
            // Clear feedback if card is actually playable
            onUnplayableCard(null, null);
          }
        } else if (!hoveringMars && onUnplayableCard) {
          // Clear feedback when not hovering Mars
          onUnplayableCard(null, null);
        }

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
    [
      isDragging,
      draggedCard,
      dragStartPosition,
      isInThrowZone,
      isHoveringMars,
      onUnplayableCard,
      isCursorOverMars,
    ],
  );

  const handleDocumentMouseUp = useCallback(() => {
    if (isDragging && draggedCard) {
      void handleDragEnd();
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
    <div className="card-fan-overlay" ref={handRef}>
      {cards.map((card, index) => {
        const spreadCard = highlightedCard || hoveredCard;
        const spreadIndex = spreadCard
          ? cards.findIndex((c) => c.id === spreadCard)
          : null;
        const position = calculateCardPosition(
          index,
          cards.length,
          spreadIndex,
          false,
        );
        const isHighlighted = highlightedCard === card.id;
        const isDraggedCard = draggedCard === card.id;
        const isHovered = hoveredCard === card.id;
        const isReturning = returningCard === card.id;
        const isUnplayableInThrowZone =
          isDraggedCard && isInThrowZone && !card.available;
        const isUnplayableOverMars =
          isDraggedCard && isHoveringMars && !card.available;

        let finalX = position.x;
        let finalY = position.y;
        const finalRotation = getCardRotation(card.id, position.rotation);
        let scale = getCardScale(card.id);

        // Cards are always in collapsed state - show only top portion
        finalY += 160;
        scale = Math.max(scale * 0.8, 0.8); // Smaller scale when collapsed

        if (isHovered && !isDragging && !isHighlighted) {
          finalY -= 60;
        }

        if (isHighlighted && !isDragging) {
          finalY -= 80;
        }

        if (isDraggedCard && !isReturning) {
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
            className={`terraforming-card ${isHighlighted ? "highlighted" : ""} ${isDraggedCard && !isReturning ? "dragged" : ""} ${isHovered ? "hovered" : ""} ${isDraggedCard && isInThrowZone && card.available ? "throw-zone" : ""} ${isUnplayableInThrowZone ? "unplayable-throw-zone" : ""} ${isUnplayableOverMars ? "unplayable-over-mars" : ""} ${isReturning ? "returning" : ""}`}
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
              isSelected={
                isHighlighted ||
                isHovered ||
                (isDraggedCard && isInThrowZone && card.available === true)
              }
              onSelect={() => {}} // Handled by parent div click
              animationDelay={0}
            />
          </div>
        );
      })}

      <style>{`
        .card-fan-overlay {
          position: fixed;
          bottom: 48px; /* Position above BottomResourceBar */
          left: 50%;
          transform: translateX(-50%);
          width: 0; /* No blocking width */
          height: 300px; /* Cards area height */
          z-index: 1100; /* Above bottom bar (1000) and its content (1001) */
          pointer-events: none; /* Don't block clicks - cards handle their own events */
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
          /* Glow is now handled by SimpleGameCard's isSelected prop */
        }

        .terraforming-card.highlighted {
          /* Glow is now handled by SimpleGameCard's isSelected prop */
        }

        .terraforming-card.dragged {
          transition: none;
          cursor: grabbing;
          z-index: 1000;
        }

        .terraforming-card.throw-zone {
          /* Glow is now handled by SimpleGameCard's isSelected prop */
          /* Keeping only the brightness filter for extra visual feedback */
          filter: brightness(1.15);
        }

        .terraforming-card.unplayable-throw-zone {
          /* No special styling for unplayable cards in throw zone */
        }

        .terraforming-card.unplayable-over-mars {
          opacity: 0.5;
          filter: grayscale(70%) brightness(0.6);
        }

        .terraforming-card:not(.dragged) {
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        .terraforming-card.returning {
          transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
          cursor: pointer;
        }


        /* Responsive Design */
        @media (max-width: 1200px) {
          .card-fan-overlay {
            height: 250px;
          }
        }

        @media (max-width: 768px) {
          .card-fan-overlay {
            height: 200px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardFanOverlay;
