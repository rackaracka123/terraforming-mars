import React from "react";
import { createPortal } from "react-dom";
import styles from "./ModalPortal.module.css";

interface ModalPortalProps {
  children: React.ReactNode;
  isOpen: boolean;
  onClose: () => void;
  level?: "primary" | "secondary" | "system";
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
  level = "primary",
}) => {
  // Ensure portal containers exist
  React.useEffect(() => {
    const createPortalContainer = (id: string) => {
      if (!document.getElementById(id)) {
        const container = document.createElement("div");
        container.id = id;
        document.body.appendChild(container);
      }
    };

    createPortalContainer("modal-primary");
    createPortalContainer("modal-secondary");
    createPortalContainer("modal-system");
  }, []);

  // Handle ESC key
  React.useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      // Prevent body scroll
      document.body.style.overflow = "hidden";

      return () => {
        document.removeEventListener("keydown", handleKeyDown);
        document.body.style.overflow = "";
      };
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const portalId = `modal-${level}`;
  const container = document.getElementById(portalId);

  if (!container) return null;

  return createPortal(
    <div className={styles.modalOverlay} onClick={onClose}>
      <div
        className={styles.modalContainer}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>,
    container,
  );
};

export default ModalPortal;
