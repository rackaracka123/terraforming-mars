import React, { createContext, useContext } from 'react';
import { createPortal } from 'react-dom';
import useModalStack, { ModalLevel } from '../../../hooks/useModalStack.ts';

interface ModalContextValue {
  openModal: (
    id: string, 
    component: React.ComponentType<any>, 
    props?: any,
    level?: ModalLevel
  ) => void;
  closeModal: (id: string) => void;
  closeTopModal: () => void;
  closeAllModals: () => void;
  hasModals: boolean;
}

const ModalContext = createContext<ModalContextValue | null>(null);

/**
 * Context provider for the modal system
 * 
 * This provider manages all modals in the application using DOM ordering
 * instead of z-index values. Modals are rendered in separate portal containers
 * that are organized by level in the DOM hierarchy.
 */
export const ModalProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const modalStack = useModalStack();

  // Ensure portal containers exist and are properly ordered
  React.useEffect(() => {
    const createPortalContainer = (id: string, className: string) => {
      if (!document.getElementById(id)) {
        const container = document.createElement('div');
        container.id = id;
        container.className = className;
        document.body.appendChild(container);
      }
    };

    // Create containers in DOM order (stacking order)
    createPortalContainer('modal-primary', 'modal-container primary');
    createPortalContainer('modal-secondary', 'modal-container secondary'); 
    createPortalContainer('modal-system', 'modal-container system');

    // Add global styles for modal containers
    const style = document.createElement('style');
    style.textContent = `
      .modal-container {
        position: relative;
      }
      
      .modal-container.primary {
        /* Base layer for primary modals */
      }
      
      .modal-container.secondary {
        /* Secondary modals appear above primary */
      }
      
      .modal-container.system {
        /* System modals appear on top */
      }
    `;
    document.head.appendChild(style);

    return () => {
      document.head.removeChild(style);
    };
  }, []);

  // Render modals grouped by level
  const renderModalsByLevel = (level: ModalLevel) => {
    const modals = modalStack.getModalsByLevel(level);
    const container = document.getElementById(`modal-${level}`);
    
    if (!container || modals.length === 0) return null;

    return createPortal(
      <div className={`modal-level-container ${level}`}>
        {modals.map(modal => {
          const ModalComponent = modal.component;
          return (
            <div key={modal.id} className="modal-overlay">
              <div className="modal-backdrop" onClick={() => modalStack.closeModal(modal.id)}>
                <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                  <ModalComponent 
                    {...modal.props}
                    onClose={() => modalStack.closeModal(modal.id)}
                    modalId={modal.id}
                  />
                </div>
              </div>
            </div>
          );
        })}

        <style jsx>{`
          .modal-level-container {
            position: relative;
          }
          
          .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            /* No z-index - DOM order provides stacking */
          }

          .modal-backdrop {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.6);
            backdrop-filter: blur(3px);
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
          }

          .modal-content {
            /* Isolation creates natural stacking context */
            isolation: isolate;
            position: relative;
          }

          /* Adjust backdrop opacity based on level */
          .primary .modal-backdrop {
            background: rgba(0, 0, 0, 0.6);
          }

          .secondary .modal-backdrop {
            background: rgba(0, 0, 0, 0.4);
          }

          .system .modal-backdrop {
            background: rgba(0, 0, 0, 0.8);
          }
        `}</style>
      </div>,
      container
    );
  };

  const contextValue: ModalContextValue = {
    openModal: modalStack.openModal,
    closeModal: modalStack.closeModal,
    closeTopModal: modalStack.closeTopModal,
    closeAllModals: modalStack.closeAllModals,
    hasModals: modalStack.hasModals
  };

  return (
    <ModalContext.Provider value={contextValue}>
      {children}
      {renderModalsByLevel('primary')}
      {renderModalsByLevel('secondary')}
      {renderModalsByLevel('system')}
    </ModalContext.Provider>
  );
};

/**
 * Hook to access the modal context
 */
export const useModal = () => {
  const context = useContext(ModalContext);
  if (!context) {
    throw new Error('useModal must be used within a ModalProvider');
  }
  return context;
};

export default ModalProvider;