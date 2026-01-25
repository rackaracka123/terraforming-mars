import { useEffect } from "react";

interface UseModalOptions {
  isVisible: boolean;
  onClose: () => void;
  closeOnEscape?: boolean;
  lockScroll?: boolean;
  preventClose?: boolean;
  onPreventedClose?: () => void;
}

export function useModal({
  isVisible,
  onClose,
  closeOnEscape = true,
  lockScroll = true,
  preventClose = false,
  onPreventedClose,
}: UseModalOptions) {
  useEffect(() => {
    if (!isVisible) return;

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (preventClose) {
          onPreventedClose?.();
        } else if (closeOnEscape) {
          onClose();
        }
      }
    };

    document.addEventListener("keydown", handleEscape);

    if (lockScroll) {
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      if (lockScroll) {
        document.body.style.overflow = "unset";
      }
    };
  }, [isVisible, onClose, closeOnEscape, lockScroll, preventClose, onPreventedClose]);
}
