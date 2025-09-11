import React from "react";
import styles from "./VictoryPointIcon.module.css";

interface VictoryPointIconProps {
  value: number | string;
  size?: "small" | "medium" | "large";
}

const VictoryPointIcon: React.FC<VictoryPointIconProps> = ({
  value,
  size = "medium",
}) => {
  // Don't render anything if value is 0, empty, or falsy
  if (value === 0 || !value) {
    return null;
  }

  // Helper function to format victory point text
  const formatVictoryPoints = (val: number | string): string => {
    if (typeof val === "number") {
      return val.toString();
    }

    // Handle special Terraforming Mars syntax
    return val
      .replace(/1 point per animal/gi, "1/🐾")
      .replace(/1 point per jupiter card/gi, "1/♃")
      .replace(/1 point per jovian/gi, "1/♃")
      .replace(/1 point per science/gi, "1/🧪")
      .replace(/1 point per space/gi, "1/🚀")
      .replace(/1 point per microbe/gi, "1/🦠")
      .replace(/1 point per plant/gi, "1/🌱")
      .replace(/1 point per city/gi, "1/🏙️")
      .replace(/1 point per earth/gi, "1/🌍")
      .replace(/1 point per energy/gi, "1/⚡")
      .replace(/1 point per building/gi, "1/🏢")
      .replace(/1 point per venus/gi, "1/♀");
  };

  return (
    <div className={`${styles.container} ${styles[size]}`}>
      <img src="/assets/mars.png" alt="VP" className={styles.icon} />
      <span className={styles.value}>{formatVictoryPoints(value)}</span>
    </div>
  );
};

export default VictoryPointIcon;
