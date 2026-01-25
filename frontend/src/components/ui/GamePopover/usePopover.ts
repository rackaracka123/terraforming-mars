import { useEffect, RefObject } from "react";

export function usePopover({
  isVisible,
  onClose,
  popoverRef,
  anchorRef,
}: {
  isVisible: boolean;
  onClose: () => void;
  popoverRef: RefObject<HTMLElement | null>;
  anchorRef?: RefObject<HTMLElement | null>;
}) {
  useEffect(() => {
    if (!isVisible) return;

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };

    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as Node;
      const outsidePopover = popoverRef.current && !popoverRef.current.contains(target);
      const outsideAnchor = !anchorRef?.current || !anchorRef.current.contains(target);
      if (outsidePopover && outsideAnchor) onClose();
    };

    document.addEventListener("keydown", handleEscape);
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, popoverRef, anchorRef]);
}
