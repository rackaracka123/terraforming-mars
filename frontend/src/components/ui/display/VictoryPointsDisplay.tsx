import React from 'react';

interface VictoryPointsDisplayProps {
  victoryPoints: number;
  size?: 'small' | 'medium' | 'large';
  className?: string;
}

const VictoryPointsDisplay: React.FC<VictoryPointsDisplayProps> = ({ 
  victoryPoints, 
  size = 'medium', 
  className = '' 
}) => {
  const sizeMap = {
    small: { fontSize: '16px', padding: '8px 12px' },
    medium: { fontSize: '24px', padding: '12px 16px' },
    large: { fontSize: '32px', padding: '16px 20px' }
  };

  const dimensions = sizeMap[size];

  return (
    <div 
      className={`victory-points-display ${className}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, rgba(20, 40, 60, 0.9) 0%, rgba(30, 50, 70, 0.8) 100%)',
        border: '2px solid rgba(255, 215, 0, 0.6)',
        borderRadius: '8px',
        padding: dimensions.padding,
        boxShadow: '0 4px 15px rgba(0, 0, 0, 0.5), 0 0 20px rgba(255, 215, 0, 0.3)',
        backdropFilter: 'blur(10px)',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}
      >
        <img 
          src="/assets/resources/tr.png" 
          alt="Victory Points" 
          style={{
            width: '20px',
            height: '20px',
            filter: 'brightness(1.2)',
          }}
        />
        <span 
          style={{
            color: '#ffffff',
            fontWeight: 'bold',
            fontSize: dimensions.fontSize,
            fontFamily: 'Courier New, monospace',
            textShadow: '2px 2px 4px rgba(0, 0, 0, 0.8)',
            lineHeight: '1',
          }}
        >
          {victoryPoints}
        </span>
      </div>
    </div>
  );
};

export default VictoryPointsDisplay;