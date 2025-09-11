import React from "react";
import styles from "./ProductionIcon.module.css";

interface ProductionIconProps {
  resourceIcon: string;
  value?: string | number;
  size?: "small" | "medium" | "large";
}

const ProductionIcon: React.FC<ProductionIconProps> = ({
  resourceIcon,
  value,
  size = "medium",
}) => {
  return (
    <div className={`${styles.container} ${styles[size]}`}>
      <img
        src="/assets/misc/production.png"
        alt="Production"
        className={styles.productionBackground}
      />
      <img src={resourceIcon} alt="Resource" className={styles.resourceIcon} />
      {value && <span className={styles.value}>{value}</span>}
    </div>
  );
};

export default ProductionIcon;
