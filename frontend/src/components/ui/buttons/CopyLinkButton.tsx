import React, { useState } from "react";

interface CopyLinkButtonProps {
  textToCopy: string;
  defaultText: string;
  copiedText?: string;
  className?: string;
  icon?: React.ReactNode;
  onCopySuccess?: () => void;
  onCopyError?: (error: Error) => void;
}

const CopyLinkButton: React.FC<CopyLinkButtonProps> = ({
  textToCopy,
  defaultText,
  copiedText = "Copied!",
  className = "",
  icon,
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
      className={`bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl py-3 px-5 text-white cursor-pointer transition-all duration-300 text-sm font-semibold font-orbitron tracking-wide backdrop-blur-space min-w-[120px] hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 disabled:cursor-default disabled:transform-none ${className}`}
      onClick={handleCopy}
      disabled={isCopied}
    >
      <span
        className={`inline-flex items-center gap-2 transition-opacity duration-300 ${isCopied ? "opacity-70" : "opacity-100"}`}
      >
        {isCopied ? copiedText : defaultText}
        {icon && !isCopied && icon}
      </span>
    </button>
  );
};

export default CopyLinkButton;
