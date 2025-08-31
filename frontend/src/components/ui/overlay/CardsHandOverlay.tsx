import React, { useState, useEffect, useRef } from 'react';

interface HearthstoneCard {
  id: string;
  name: string;
  cost: number;
  type: 'spell' | 'minion' | 'weapon';
  attack?: number;
  health?: number;
  description: string;
  rarity: 'common' | 'rare' | 'epic' | 'legendary';
  playable: boolean;
}

interface CardsHandOverlayProps {}

const CardsHandOverlay: React.FC<CardsHandOverlayProps> = () => {
  const [highlightedCard, setHighlightedCard] = useState<string | null>(null);
  const [draggedCard, setDraggedCard] = useState<string | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 }); // Offset from initial click to card position
  const [dragPosition, setDragPosition] = useState({ x: 0, y: 0 }); // Current mouse position
  const [isDragging, setIsDragging] = useState(false);
  const [justDragged, setJustDragged] = useState(false); // Track if we just finished dragging
  const [isHandSpread, setIsHandSpread] = useState(false); // Track if hand should be spread
  const handRef = useRef<HTMLDivElement>(null);

  // Mock Hearthstone cards data
  const hearthstoneCards: HearthstoneCard[] = [
    {
      id: '1',
      name: 'Fireball',
      cost: 4,
      type: 'spell',
      description: 'Deal 6 damage.',
      rarity: 'common',
      playable: true
    },
    {
      id: '2',
      name: 'Chillwind Yeti',
      cost: 4,
      type: 'minion',
      attack: 4,
      health: 5,
      description: 'A solid minion.',
      rarity: 'common',
      playable: true
    },
    {
      id: '3',
      name: 'Archmage Antonidas',
      cost: 7,
      type: 'minion',
      attack: 5,
      health: 7,
      description: 'Whenever you cast a spell, add a Fireball to your hand.',
      rarity: 'legendary',
      playable: false
    },
    {
      id: '4',
      name: 'Lightning Bolt',
      cost: 1,
      type: 'spell',
      description: 'Deal 3 damage. Overload: (1)',
      rarity: 'common',
      playable: true
    },
    {
      id: '5',
      name: 'Boulderfist Ogre',
      cost: 6,
      type: 'minion',
      attack: 6,
      health: 7,
      description: 'A big, dumb creature.',
      rarity: 'common',
      playable: true
    },
    {
      id: '6',
      name: 'Polymorph',
      cost: 4,
      type: 'spell',
      description: 'Transform a minion into a 1/1 Sheep.',
      rarity: 'common',
      playable: true
    },
    {
      id: '7',
      name: 'Ysera',
      cost: 9,
      type: 'minion',
      attack: 4,
      health: 12,
      description: 'At the end of your turn, add a Dream Card to your hand.',
      rarity: 'legendary',
      playable: false
    }
  ];

  // Calculate card positions in arc - compact or spread based on hover
  const calculateCardPosition = (index: number, totalCards: number, isSpread: boolean = false) => {
    const handWidth = isSpread ? 600 : 400; // Spread vs compact width
    const handCurve = 0.3; // How curved the hand is
    const cardWidth = 120;
    const baseY = -20; // Base Y position
    
    // Adjust spacing based on spread state
    const maxSpacing = cardWidth * (isSpread ? 0.7 : 0.4);
    const spacing = Math.min(maxSpacing, handWidth / Math.max(totalCards - 1, 1));
    
    // Center the hand
    const totalWidth = spacing * (totalCards - 1);
    const startX = -totalWidth / 2;
    const x = startX + (index * spacing);
    
    // Create arc curve
    const normalizedX = x / (handWidth / 2); // -1 to 1
    const curveY = Math.pow(Math.abs(normalizedX), 2) * handCurve * 60;
    const y = baseY + curveY;
    
    // Adjust rotation based on spread state
    const maxRotation = isSpread ? 12 : 8;
    const rotation = normalizedX * maxRotation;
    
    return { x, y, rotation };
  };

  const getCardTypeColor = (type: string) => {
    switch (type) {
      case 'spell': return '#8B5BFF';
      case 'minion': return '#FFB800';
      case 'weapon': return '#FF6B6B';
      default: return '#95a5a6';
    }
  };

  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case 'common': return '#ffffff';
      case 'rare': return '#0099cc';
      case 'epic': return '#cc00ff';
      case 'legendary': return '#ff8000';
      default: return '#ffffff';
    }
  };

  // Handle card click (highlight)
  const handleCardClick = (cardId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    if (isDragging || justDragged) return; // Ignore clicks right after dragging
    
    // Don't allow interaction with unplayable cards
    const card = hearthstoneCards.find(c => c.id === cardId);
    if (card && !card.playable) return;
    
    if (highlightedCard === cardId) {
      setHighlightedCard(null);
    } else {
      setHighlightedCard(cardId);
    }
  };

  // Handle drag start
  const handleDragStart = (cardId: string, event: React.MouseEvent) => {
    event.preventDefault(); // Prevent browser default drag behavior
    
    // Calculate initial offset from mouse to card position
    const cardIndex = hearthstoneCards.findIndex(c => c.id === cardId);
    const cardPosition = calculateCardPosition(cardIndex, hearthstoneCards.length, isHandSpread);
    const containerRect = handRef.current?.getBoundingClientRect();
    
    if (containerRect) {
      // Current card screen position
      const cardScreenX = containerRect.left + containerRect.width / 2 + cardPosition.x;
      const cardScreenY = containerRect.bottom + cardPosition.y;
      
      // Store offset from mouse to card position
      setDragOffset({
        x: cardScreenX - event.clientX,
        y: cardScreenY - event.clientY
      });
    }
    
    setDraggedCard(cardId);
    setIsDragging(true);
    setDragPosition({ x: event.clientX, y: event.clientY });
    setHighlightedCard(null); // Clear highlight when dragging
  };

  // Handle drag move - using document mouse move for better performance
  const handleDragMove = (event: MouseEvent) => {
    if (!isDragging || !draggedCard) return;
    setDragPosition({ x: event.clientX, y: event.clientY });
  };

  // Handle drag end
  const handleDragEnd = () => {
    setDraggedCard(null);
    setIsDragging(false);
    setDragPosition({ x: 0, y: 0 });
    setDragOffset({ x: 0, y: 0 }); // Reset offset
    setHighlightedCard(null); // Deselect card when released
    setJustDragged(true);
    
    // Clear the justDragged flag after a short delay to allow clicks again
    setTimeout(() => {
      setJustDragged(false);
    }, 100);
  };

  // Handle click outside to dismiss
  const handleClickOutside = (event: React.MouseEvent) => {
    if (handRef.current && !handRef.current.contains(event.target as Node)) {
      setHighlightedCard(null);
    }
  };

  // Add document event listeners for drag and click outside
  useEffect(() => {
    const handleDocumentClick = (event: MouseEvent) => {
      if (handRef.current && !handRef.current.contains(event.target as Node)) {
        setHighlightedCard(null);
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

    document.addEventListener('click', handleDocumentClick);
    document.addEventListener('mousemove', handleDocumentMouseMove);
    document.addEventListener('mouseup', handleDocumentMouseUp);
    
    return () => {
      document.removeEventListener('click', handleDocumentClick);
      document.removeEventListener('mousemove', handleDocumentMouseMove);
      document.removeEventListener('mouseup', handleDocumentMouseUp);
    };
  }, [isDragging, draggedCard]);

  return (
    <div 
      className="hearthstone-hand-overlay"
    >
      <div 
        className="card-hand-container"
        ref={handRef}
        onMouseEnter={() => setIsHandSpread(true)}
        onMouseLeave={() => setIsHandSpread(false)}
      >
        {hearthstoneCards.map((card, index) => {
          const position = calculateCardPosition(index, hearthstoneCards.length, isHandSpread);
          const isHighlighted = highlightedCard === card.id;
          const isDraggedCard = draggedCard === card.id;
          
          // Calculate final position and transform
          let finalX = position.x;
          let finalY = position.y;
          let finalRotation = position.rotation;
          let scale = 1;
          let zIndex = index;
          
          if (isHighlighted && !isDragging) {
            finalY -= 60; // Move up when highlighted/clicked
            scale = 1.2;
            zIndex = 1000;
          }
          
          if (isDraggedCard) {
            // Apply mouse position with offset to maintain relative position
            const containerRect = handRef.current?.getBoundingClientRect();
            if (containerRect) {
              // Apply mouse position + stored offset, then convert to container coordinates
              const targetScreenX = dragPosition.x + dragOffset.x;
              const targetScreenY = dragPosition.y + dragOffset.y;
              
              finalX = targetScreenX - (containerRect.left + containerRect.width / 2);
              finalY = targetScreenY - containerRect.bottom;
            }
            finalRotation = 0;
            scale = 1.1;
            zIndex = 1001;
          }
          
          return (
            <div
              key={card.id}
              className={`hearthstone-card ${isHighlighted ? 'highlighted' : ''} ${isDraggedCard ? 'dragged' : ''} ${!card.playable ? 'unplayable' : ''}`}
              style={{
                transform: `translate(${finalX}px, ${finalY}px) rotate(${finalRotation}deg) scale(${scale})`,
                zIndex: zIndex,
                '--card-type-color': getCardTypeColor(card.type),
                '--rarity-color': getRarityColor(card.rarity)
              } as React.CSSProperties}
              onClick={(e) => handleCardClick(card.id, e)}
              onMouseDown={(e) => handleDragStart(card.id, e)}
            >
              <div className="card-inner">
                <div className="card-cost">{card.cost}</div>
                <div className="card-header">
                  <div className="card-name" style={{ color: getRarityColor(card.rarity) }}>
                    {card.name}
                  </div>
                </div>
                
                {card.type === 'minion' && (
                  <div className="card-stats">
                    <div className="attack">{card.attack}</div>
                    <div className="health">{card.health}</div>
                  </div>
                )}
                
                <div className="card-description">
                  {card.description}
                </div>
                
                <div className="card-type-indicator" />
              </div>
            </div>
          );
        })}
      </div>

      <style jsx>{`
        .hearthstone-hand-overlay {
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          height: 100vh;
          z-index: 150;
          pointer-events: none;
        }

        .card-hand-container {
          position: absolute;
          bottom: 100px;
          left: 50%;
          transform: translateX(-50%);
          width: 100%;
          height: 200px;
          pointer-events: auto;
          transition: all 0.3s ease;
        }

        .hearthstone-card {
          position: absolute;
          width: 120px;
          height: 180px;
          bottom: 0;
          left: 50%;
          cursor: pointer;
          transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
          transform-origin: bottom center;
          pointer-events: auto;
          user-select: none;
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
          transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }

        .hearthstone-card.unplayable {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .card-inner {
          width: 100%;
          height: 100%;
          background: linear-gradient(
            135deg,
            rgba(20, 25, 40, 0.95) 0%,
            rgba(35, 40, 55, 0.9) 50%,
            rgba(25, 30, 45, 0.95) 100%
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
        }

        .card-header {
          margin-bottom: 8px;
        }

        .card-name {
          font-size: 11px;
          font-weight: bold;
          text-align: center;
          line-height: 1.2;
          color: var(--rarity-color);
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .card-stats {
          position: absolute;
          bottom: 4px;
          left: 4px;
          right: 4px;
          display: flex;
          justify-content: space-between;
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
        }

        .health {
          background: linear-gradient(135deg, #2ecc71 0%, #27ae60 100%);
        }

        .card-description {
          flex: 1;
          font-size: 8px;
          color: rgba(255, 255, 255, 0.9);
          text-align: center;
          line-height: 1.3;
          margin: 8px 0;
          padding: 4px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 4px;
          overflow: hidden;
          display: -webkit-box;
          -webkit-line-clamp: 6;
          -webkit-box-orient: vertical;
        }

        .card-type-indicator {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          height: 3px;
          background: var(--card-type-color);
          border-radius: 8px 8px 0 0;
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
            bottom: 80px;
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