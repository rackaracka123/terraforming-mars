import React from "react";
import { createPortal } from "react-dom";

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

    return undefined;
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const portalId = `modal-${level}`;
  const container = document.getElementById(portalId);

  if (!container) return null;

  return createPortal(
    <div
      className="fixed top-0 left-0 right-0 bottom-0 bg-black/80 [backdrop-filter:blur(5px)] flex items-center justify-center p-5 max-[800px]:p-2.5"
      onClick={onClose}
    >
      <div
        className="bg-[linear-gradient(135deg,rgba(10,20,40,0.98)_0%,rgba(20,30,50,0.96)_50%,rgba(15,25,45,0.98)_100%)] border-2 border-[rgba(100,150,255,0.5)] rounded-[20px] max-w-[1000px] max-h-[80vh] w-full overflow-y-auto [backdrop-filter:blur(20px)] shadow-[0_20px_60px_rgba(0,0,0,0.8),0_0_40px_rgba(100,150,255,0.3)] relative isolate max-[800px]:max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>,
    container,
  );
};

export default ModalPortal;
