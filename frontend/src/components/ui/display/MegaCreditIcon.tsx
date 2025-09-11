import React from "react";
import styles from "./MegaCreditIcon.module.css";

interface MegaCreditIconProps {
  value: number;
  size?: "small" | "medium" | "large";
}

const MegaCreditIcon: React.FC<MegaCreditIconProps> = ({
  value,
  size = "medium",
}) => {
  return (
    <div className={`${styles.container} ${styles[size]}`}>
      <img
        src="/assets/resources/megacredit.png"
        alt="MC"
        className={styles.icon}
      />
      <span className={styles.value}>{value}</span>
    </div>
  );
};

export default MegaCreditIcon;
