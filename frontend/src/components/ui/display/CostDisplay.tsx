import React from 'react';
import { Z_INDEX } from '../../../constants/zIndex.ts';

interface CostDisplayProps {
  cost: number;
  size?: 'small' | 'medium' | 'large';
  className?: string;
}

const CostDisplay: React.FC<CostDisplayProps> = ({ cost, size = 'medium', className = '' }) => {
  const sizeMap = {
    small: { container: 24, icon: 24, fontSize: '10px' },
    medium: { container: 32, icon: 32, fontSize: '12px' },
    large: { container: 40, icon: 40, fontSize: '14px' }
  };

  const dimensions = sizeMap[size];

  return (
    <div 
      className={`cost-display ${className}`}
      style={{
        position: 'relative',
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: `${dimensions.container}px`,
        height: `${dimensions.container}px`,
        borderRadius: '4px',
      }}
    >
      <img 
        src="/assets/resources/megacredit.png" 
        alt="Megacredits" 
        style={{
          width: `${dimensions.icon}px`,
          height: `${dimensions.icon}px`,
          display: 'block',
        }}
      />
      <span 
        style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          zIndex: Z_INDEX.COST_DISPLAY,
          color: '#000000',
          fontWeight: 'bold',
          fontSize: dimensions.fontSize,
          textAlign: 'center',
          fontFamily: 'Arial, sans-serif',
          lineHeight: '1',
          textShadow: '0.5px 0.5px 1px rgba(255, 255, 255, 0.8)',
          whiteSpace: 'nowrap',
        }}
      >
        {cost}
      </span>
    </div>
  );
};

export default CostDisplay;