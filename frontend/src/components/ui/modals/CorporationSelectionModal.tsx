import React, { useState, useRef } from 'react';
import CorporationCard from '../cards/CorporationCard.tsx';

interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingResources?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  expansion?: string;
  logoPath?: string;
}

interface CorporationSelectionModalProps {
  corporations: Corporation[];
  onSelectCorporation: (corporationId: string) => void;
  isVisible: boolean;
}

const CorporationSelectionModal: React.FC<CorporationSelectionModalProps> = ({ 
  corporations, 
  onSelectCorporation, 
  isVisible 
}) => {
  const [selectedCorporation, setSelectedCorporation] = useState<string | null>(null);
  const [isFlashing, setIsFlashing] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);

  if (!isVisible) return null;

  const handleOverlayClick = (e: React.MouseEvent) => {
    // Prevent click from reaching overlay if clicking on modal content
    if (modalRef.current && modalRef.current.contains(e.target as Node)) {
      return;
    }
    
    // Flash animation when trying to dismiss
    setIsFlashing(true);
    setTimeout(() => setIsFlashing(false), 600);
  };

  const handleCorporationSelect = (corporationId: string) => {
    setSelectedCorporation(corporationId);
  };

  const handleConfirmSelection = () => {
    if (selectedCorporation) {
      onSelectCorporation(selectedCorporation);
    }
  };


  return (
    <div className={`modal-overlay ${isFlashing ? 'flashing' : ''}`} onClick={handleOverlayClick}>
      <div className="modal-content" ref={modalRef}>
        <div className="modal-header">
          <h2>Choose Your Corporation</h2>
          <p>Select a corporation to begin your Mars terraforming journey</p>
        </div>
        
        <div className="corporations-grid">
          {corporations.map((corp) => (
            <CorporationCard
              key={corp.id}
              corporation={corp}
              isSelected={selectedCorporation === corp.id}
              onSelect={handleCorporationSelect}
            />
          ))}
        </div>

        <div className="modal-actions">
          <button
            className="confirm-btn"
            disabled={!selectedCorporation}
            onClick={handleConfirmSelection}
          >
            Confirm Selection
          </button>
        </div>

        <style jsx>{`
          .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.9);
            backdrop-filter: blur(10px);
            display: flex;
            align-items: center;
            justify-content: center;
            /* z-index removed - natural DOM order places modal above other elements */
            /* Use isolation to ensure proper stacking above all other content */
            isolation: isolate;
            padding: 20px;
            transition: all 0.3s ease;
          }

          .modal-overlay.flashing {
            animation: flashBorder 0.6s ease;
          }

          @keyframes flashBorder {
            0%, 100% { box-shadow: inset 0 0 0 0px rgba(255, 0, 0, 0); }
            25% { box-shadow: inset 0 0 0 8px rgba(255, 0, 0, 0.8); }
            50% { box-shadow: inset 0 0 0 4px rgba(255, 0, 0, 0.6); }
            75% { box-shadow: inset 0 0 0 6px rgba(255, 0, 0, 0.4); }
          }

          .modal-content {
            background: linear-gradient(
              135deg,
              rgba(10, 20, 40, 0.98) 0%,
              rgba(20, 30, 50, 0.96) 50%,
              rgba(15, 25, 45, 0.98) 100%
            );
            border: 2px solid rgba(100, 150, 255, 0.5);
            border-radius: 20px;
            max-width: 1200px;
            max-height: 90vh;
            width: 100%;
            overflow-y: auto;
            backdrop-filter: blur(20px);
            box-shadow: 
              0 20px 60px rgba(0, 0, 0, 0.8),
              0 0 40px rgba(100, 150, 255, 0.3);
            animation: modalPulse 2s ease-in-out infinite;
          }

          @keyframes modalPulse {
            0%, 100% { 
              border-color: rgba(100, 150, 255, 0.5);
              box-shadow: 
                0 20px 60px rgba(0, 0, 0, 0.8),
                0 0 40px rgba(100, 150, 255, 0.3);
            }
            50% { 
              border-color: rgba(255, 200, 100, 0.8);
              box-shadow: 
                0 20px 60px rgba(0, 0, 0, 0.8),
                0 0 60px rgba(255, 200, 100, 0.5),
                0 0 100px rgba(255, 200, 100, 0.2);
            }
          }

          .modal-header {
            text-align: center;
            padding: 30px 30px 20px 30px;
            border-bottom: 1px solid rgba(100, 150, 255, 0.3);
          }

          .modal-header h2 {
            font-size: 32px;
            color: #ffffff;
            margin-bottom: 8px;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
            animation: headerPulse 3s ease-in-out infinite;
          }

          @keyframes headerPulse {
            0%, 100% { 
              color: #ffffff;
              text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
            }
            50% { 
              color: #ffcc66;
              text-shadow: 
                0 2px 4px rgba(0, 0, 0, 0.8),
                0 0 20px rgba(255, 204, 102, 0.6);
            }
          }

          .modal-header p {
            font-size: 16px;
            color: rgba(255, 255, 255, 0.7);
            margin: 0;
          }

          .corporations-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
            padding: 30px;
          }


          .modal-actions {
            padding: 20px 30px 30px 30px;
            text-align: center;
            border-top: 1px solid rgba(100, 150, 255, 0.3);
          }

          .confirm-btn {
            background: linear-gradient(135deg, #4a90e2 0%, #5ba0f2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            padding: 12px 30px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.2s ease;
            position: relative;
            overflow: hidden;
          }

          .confirm-btn:not(:disabled) {
            animation: buttonPulse 2.5s ease-in-out infinite;
          }

          @keyframes buttonPulse {
            0%, 100% { 
              transform: scale(1);
              box-shadow: 0 4px 12px rgba(74, 144, 226, 0.3);
            }
            50% { 
              transform: scale(1.05);
              box-shadow: 
                0 6px 20px rgba(74, 144, 226, 0.5),
                0 0 30px rgba(91, 160, 242, 0.4);
            }
          }

          .confirm-btn:hover:not(:disabled) {
            background: linear-gradient(135deg, #357abd 0%, #4a90e2 100%);
            transform: translateY(-1px);
          }

          .confirm-btn:disabled {
            background: rgba(100, 100, 100, 0.5);
            color: rgba(255, 255, 255, 0.5);
            cursor: not-allowed;
            transform: none;
          }

          @media (max-width: 800px) {
            .corporations-grid {
              grid-template-columns: 1fr;
              padding: 20px;
            }
            
            .modal-header h2 {
              font-size: 24px;
            }
          }
        `}</style>
      </div>
    </div>
  );
};

export default CorporationSelectionModal;