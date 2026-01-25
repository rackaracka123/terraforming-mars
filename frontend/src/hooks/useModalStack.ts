import { useState, useCallback, useEffect } from "react";

export type ModalLevel = "primary" | "secondary" | "system";

interface ModalState {
  id: string;
  level: ModalLevel;
  component: React.ComponentType<Record<string, unknown>>;
  props?: Record<string, unknown>;
}

/**
 * Hook for managing modal stack without z-index
 *
 * This hook manages a stack of modals using React state instead of CSS layering.
 * Modals are rendered in DOM order (later modals appear on top) and managed
 * through declarative state updates.
 *
 * Usage:
 * ```
 * const { openModal, closeModal, modals } = useModalStack();
 *
 * const showConfirmation = () => {
 *   openModal('confirm-delete', ConfirmationModal, {
 *     message: 'Are you sure?',
 *     level: 'secondary'
 *   });
 * };
 * ```
 */
export const useModalStack = () => {
  const [modals, setModals] = useState<ModalState[]>([]);

  const openModal = useCallback(
    (
      id: string,
      component: React.ComponentType<Record<string, unknown>>,
      props: Record<string, unknown> = {},
      level: ModalLevel = "primary",
    ) => {
      setModals((prev) => {
        const filtered = prev.filter((modal) => modal.id !== id);
        return [...filtered, { id, level, component, props }];
      });
    },
    [],
  );

  const closeModal = useCallback((id: string) => {
    setModals((prev) => prev.filter((modal) => modal.id !== id));
  }, []);

  const closeTopModal = useCallback(() => {
    setModals((prev) => prev.slice(0, -1));
  }, []);

  const closeAllModals = useCallback(() => {
    setModals([]);
  }, []);

  const getModalsByLevel = useCallback(
    (level: ModalLevel) => {
      return modals.filter((modal) => modal.level === level);
    },
    [modals],
  );

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && modals.length > 0) {
        closeTopModal();
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [modals.length, closeTopModal]);

  return {
    modals,
    openModal,
    closeModal,
    closeTopModal,
    closeAllModals,
    getModalsByLevel,
    hasModals: modals.length > 0,
    topModal: modals[modals.length - 1] || null,
  };
};

export default useModalStack;
