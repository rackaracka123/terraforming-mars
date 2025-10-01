import React, { useEffect, useRef, useState } from "react";
import {
  CardDto,
  ChoiceDto,
  CardBehaviorDto,
} from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import styles from "./ChoiceSelectionPopover.module.css";

interface ChoiceItem {
  index: number;
  choice: ChoiceDto;
}

interface ChoiceSelectionPopoverProps {
  card: CardDto;
  behaviorIndex: number;
  onChoiceSelect: (choiceIndex: number) => void;
  onCancel: () => void;
  isVisible: boolean;
}

const ChoiceSelectionPopover: React.FC<ChoiceSelectionPopoverProps> = ({
  card,
  behaviorIndex,
  onChoiceSelect,
  onCancel,
  isVisible,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  // Get the behavior with choices
  const behavior = card.behaviors?.[behaviorIndex];
  const choices: ChoiceItem[] =
    behavior?.choices?.map((choice, index) => ({
      index,
      choice,
    })) || [];

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200); // Match fadeOut animation duration
  };

  const handleChoiceClick = (choiceIndex: number) => {
    // Call immediately without delay for choice selection
    onChoiceSelect(choiceIndex);
  };

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleCancelClick();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      // Only close on left click (button 0), ignore right click (button 2) and middle click (button 1)
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        handleCancelClick();
      }
    };

    const preventScroll = (event: WheelEvent | TouchEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      // Prevent body scroll
      document.body.style.overflow = "hidden";

      // Add event listeners
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      // Restore body scroll
      document.body.style.overflow = "";

      // Remove event listeners
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel]);

  if (!isVisible || choices.length === 0) return null;

  return (
    <div className={styles.overlay}>
      <div
        className={`${styles.popover} ${isClosing ? styles.closing : ""}`}
        ref={popoverRef}
      >
        <div className={styles.header}>
          <h3 className={styles.title}>Choose One Effect</h3>
          <div className={styles.subtitle}>{card.name}</div>
        </div>

        <div className={styles.choicesContainer}>
          {choices.map(({ index, choice }) => {
            // Convert the choice into a CardBehaviorDto that BehaviorSection can render
            const behaviorForChoice: CardBehaviorDto = {
              triggers: behavior?.triggers || [],
              inputs: choice.inputs,
              outputs: choice.outputs,
              choices: undefined, // Don't show nested choices
            };

            return (
              <div
                key={index}
                className={styles.choiceItem}
                onClick={() => handleChoiceClick(index)}
              >
                <div className={styles.choiceLabel}>Choice {index + 1}</div>
                <div className={styles.choiceContent}>
                  <BehaviorSection behaviors={[behaviorForChoice]} />
                </div>
              </div>
            );
          })}
        </div>

        <div className={styles.footer}>
          <button className={styles.cancelButton} onClick={handleCancelClick}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChoiceSelectionPopover;
