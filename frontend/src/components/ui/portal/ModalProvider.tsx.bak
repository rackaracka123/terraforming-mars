import React, { createContext, useContext } from "react";
import { createPortal } from "react-dom";
import useModalStack, { ModalLevel } from "../../../hooks/useModalStack.ts";
import styles from "./ModalProvider.module.css";

interface ModalContextValue {
  openModal: (
    id: string,
    component: React.ComponentType<Record<string, unknown>>,
    props?: Record<string, unknown>,
    level?: ModalLevel,
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
export const ModalProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const modalStack = useModalStack();

  // Ensure portal containers exist and are properly ordered
  React.useEffect(() => {
    const createPortalContainer = (id: string, className: string) => {
      if (!document.getElementById(id)) {
        const container = document.createElement("div");
        container.id = id;
        container.className = className;
        document.body.appendChild(container);
      }
    };

    // Create containers in DOM order (stacking order)
    createPortalContainer("modal-primary", "modal-container primary");
    createPortalContainer("modal-secondary", "modal-container secondary");
    createPortalContainer("modal-system", "modal-container system");

    // Add global styles for modal containers
    const style = document.createElement("style");
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
      <div className={`${styles.modalLevelContainer} ${level}`}>
        {modals.map((modal) => {
          const ModalComponent = modal.component;
          return (
            <div key={modal.id} className={styles.modalOverlay}>
              <div
                className={styles.modalBackdrop}
                onClick={() => modalStack.closeModal(modal.id)}
              >
                <div
                  className={styles.modalContent}
                  onClick={(e) => e.stopPropagation()}
                >
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
      </div>,
      container,
    );
  };

  const contextValue: ModalContextValue = {
    openModal: modalStack.openModal,
    closeModal: modalStack.closeModal,
    closeTopModal: modalStack.closeTopModal,
    closeAllModals: modalStack.closeAllModals,
    hasModals: modalStack.hasModals,
  };

  return (
    <ModalContext.Provider value={contextValue}>
      {children}
      {renderModalsByLevel("primary")}
      {renderModalsByLevel("secondary")}
      {renderModalsByLevel("system")}
    </ModalContext.Provider>
  );
};

/**
 * Hook to access the modal context
 */
export const useModal = () => {
  const context = useContext(ModalContext);
  if (!context) {
    throw new Error("useModal must be used within a ModalProvider");
  }
  return context;
};

export default ModalProvider;
