import React, { useState } from "react";

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
      className={`bg-space-black-darker/90 border-2 border-space-blue-800 rounded-lg py-3 px-5 text-white cursor-pointer transition-all duration-300 text-sm font-semibold backdrop-blur-space min-w-[120px] hover:bg-space-black-darker/95 hover:border-space-blue-600 hover:shadow-glow hover:-translate-y-0.5 disabled:cursor-default disabled:transform-none ${className}`}
      onClick={handleCopy}
      disabled={isCopied}
    >
      <span
        className={`inline-block transition-opacity duration-300 ${isCopied ? "opacity-70" : "opacity-100"}`}
      >
        {isCopied ? copiedText : defaultText}
      </span>
    </button>
  );
};

export default CopyLinkButton;
