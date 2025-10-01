import React, { useState } from "react";
import styles from "./CopyLinkButton.module.css";

interface CopyLinkButtonProps {
  textToCopy: string;
  defaultText: string;
  copiedText?: string;
  className?: string;
  onCopySuccess?: () => void;
  onCopyError?: (error: Error) => void;
}

const CopyLinkButton: React.FC<CopyLinkButtonProps> = ({
  textToCopy,
  defaultText,
  copiedText = "Copied!",
  className = "",
  onCopySuccess,
  onCopyError,
}) => {
  const [isCopied, setIsCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setIsCopied(true);
      onCopySuccess?.();

      // Fade back to default text after 1 second
      setTimeout(() => {
        setIsCopied(false);
      }, 1000);
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      onCopyError?.(error as Error);
    }
  };

  return (
    <button
      className={`${styles.copyLinkButton} ${className}`}
      onClick={handleCopy}
      disabled={isCopied}
    >
      <span className={`${styles.buttonText} ${isCopied ? styles.fadeOut : styles.fadeIn}`}>
        {isCopied ? copiedText : defaultText}
      </span>
    </button>
  );
};

export default CopyLinkButton;
