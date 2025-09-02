import React from 'react';
import { createPortal } from 'react-dom';

interface ModalPortalProps {
  children: React.ReactNode;
  isOpen: boolean;
  onClose: () => void;
  level?: 'primary' | 'secondary' | 'system';
}

/**
 * Portal-based modal system that uses DOM ordering instead of z-index
 * 
 * Modals are rendered in separate DOM nodes in order:
 * - #modal-primary: Main modals (game actions, card details)
 * - #modal-secondary: Confirmation dialogs, sub-modals
 * - #modal-system: Critical system messages, errors
 * 
 * This eliminates the need for z-index management while maintaining
 * proper stacking order through DOM hierarchy.
 */
const ModalPortal: React.FC<ModalPortalProps> = ({ 
  children, 
  isOpen, 
  onClose,
  level = 'primary' 
}) => {
  // Ensure portal containers exist
  React.useEffect(() => {
    const createPortalContainer = (id: string) => {
      if (!document.getElementById(id)) {
        const container = document.createElement('div');
        container.id = id;
        document.body.appendChild(container);
      }
    };

    createPortalContainer('modal-primary');
    createPortalContainer('modal-secondary');
    createPortalContainer('modal-system');
  }, []);

  // Handle ESC key
  React.useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      // Prevent body scroll
      document.body.style.overflow = 'hidden';
      
      return () => {
        document.removeEventListener('keydown', handleKeyDown);
        document.body.style.overflow = '';
      };
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const portalId = `modal-${level}`;
  const container = document.getElementById(portalId);
  
  if (!container) return null;

  return createPortal(
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-container" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
      
      <style jsx>{`
        .modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.8);
          backdrop-filter: blur(5px);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
          /* No z-index needed - DOM order handles stacking */
        }

        .modal-container {
          background: linear-gradient(
            135deg,
            rgba(10, 20, 40, 0.98) 0%,
            rgba(20, 30, 50, 0.96) 50%,
            rgba(15, 25, 45, 0.98) 100%
          );
          border: 2px solid rgba(100, 150, 255, 0.5);
          border-radius: 20px;
          max-width: 1000px;
          max-height: 80vh;
          width: 100%;
          overflow-y: auto;
          backdrop-filter: blur(20px);
          box-shadow: 
            0 20px 60px rgba(0, 0, 0, 0.8),
            0 0 40px rgba(100, 150, 255, 0.3);
          position: relative;
          /* Container isolation creates stacking context naturally */
          isolation: isolate;
        }

        @media (max-width: 800px) {
          .modal-overlay {
            padding: 10px;
          }
          
          .modal-container {
            max-height: 90vh;
          }
        }
      `}</style>
    </div>,
    container
  );
};

export default ModalPortal;