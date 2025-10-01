import React, { useState, useEffect } from "react";
import styles from "./LoadingOverlay.module.css";

interface LoadingOverlayProps {
  isLoading: boolean;
  message?: string;
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({
  isLoading,
  message = "Loading...",
}) => {
  const [isVisible, setIsVisible] = useState(isLoading);
  const [isFadingOut, setIsFadingOut] = useState(false);

  useEffect(() => {
    if (isLoading) {
      setIsVisible(true);
      setIsFadingOut(false);
    } else if (isVisible) {
      // Start fade out animation
      setIsFadingOut(true);
      // Remove overlay after animation completes
      const timeout = setTimeout(() => {
        setIsVisible(false);
      }, 400); // Match CSS transition duration
      return () => clearTimeout(timeout);
    }
  }, [isLoading, isVisible]);

  if (!isVisible) return null;

  return (
    <div className={`${styles.overlay} ${isFadingOut ? styles.fadeOut : ""}`}>
      <div className={styles.spinnerContainer}>
        <div className={styles.spinner} />
        <div className={styles.message}>{message}</div>
      </div>
    </div>
  );
};

export default LoadingOverlay;
