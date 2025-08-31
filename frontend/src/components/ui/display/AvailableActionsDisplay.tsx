import React from 'react';

interface AvailableActionsDisplayProps {
  availableActions: number;
  availableEffects: number;
  size?: 'small' | 'medium' | 'large';
  className?: string;
  onClick?: () => void;
}

const AvailableActionsDisplay: React.FC<AvailableActionsDisplayProps> = ({ 
  availableActions, 
  availableEffects,
  size = 'medium', 
  className = '',
  onClick
}) => {
  const sizeMap = {
    small: { iconSize: 16, fontSize: '10px', padding: '4px 8px', gap: '8px' },
    medium: { iconSize: 20, fontSize: '12px', padding: '6px 10px', gap: '12px' },
    large: { iconSize: 24, fontSize: '14px', padding: '8px 12px', gap: '16px' }
  };

  const dimensions = sizeMap[size];

  return (
    <button
      className={`available-actions-display ${className}`}
      onClick={onClick}
      disabled={!onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: dimensions.gap,
        background: 'linear-gradient(135deg, rgba(80, 60, 40, 0.9) 0%, rgba(70, 50, 30, 0.8) 100%)',
        border: '2px solid rgba(255, 200, 100, 0.4)',
        borderRadius: '8px',
        padding: dimensions.padding,
        boxShadow: '0 2px 10px rgba(0, 0, 0, 0.4)',
        backdropFilter: 'blur(8px)',
        cursor: onClick ? 'pointer' : 'default',
        transition: 'all 0.3s ease',
        width: '100%',
      }}
      onMouseEnter={(e) => {
        if (onClick) {
          e.currentTarget.style.background = 'linear-gradient(135deg, rgba(90, 70, 50, 0.95) 0%, rgba(80, 60, 40, 0.9) 100%)';
          e.currentTarget.style.borderColor = 'rgba(255, 220, 120, 0.6)';
          e.currentTarget.style.transform = 'scale(1.02)';
          e.currentTarget.style.boxShadow = '0 4px 15px rgba(0, 0, 0, 0.5)';
        }
      }}
      onMouseLeave={(e) => {
        if (onClick) {
          e.currentTarget.style.background = 'linear-gradient(135deg, rgba(80, 60, 40, 0.9) 0%, rgba(70, 50, 30, 0.8) 100%)';
          e.currentTarget.style.borderColor = 'rgba(255, 200, 100, 0.4)';
          e.currentTarget.style.transform = 'scale(1)';
          e.currentTarget.style.boxShadow = '0 2px 10px rgba(0, 0, 0, 0.4)';
        }
      }}
    >
      {/* Actions Section */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          padding: '2px 6px',
          background: 'rgba(0, 0, 0, 0.3)',
          borderRadius: '4px',
          border: '1px solid rgba(255, 255, 255, 0.1)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '2px',
          }}
        >
          <span
            style={{
              color: 'rgba(255, 255, 255, 0.8)',
              fontSize: '9px',
              fontWeight: 'bold',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
            }}
          >
            ACTIONS
          </span>
          <span 
            style={{
              color: '#ffffff',
              fontSize: dimensions.fontSize,
              fontWeight: 'bold',
              fontFamily: 'Courier New, monospace',
              textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)',
              lineHeight: '1',
              minWidth: '16px',
              textAlign: 'center',
              background: availableActions > 0 ? 'rgba(0, 255, 100, 0.2)' : 'rgba(100, 100, 100, 0.2)',
              padding: '2px 4px',
              borderRadius: '3px',
              border: availableActions > 0 ? '1px solid rgba(0, 255, 100, 0.3)' : '1px solid rgba(100, 100, 100, 0.3)',
            }}
          >
            {availableActions}
          </span>
        </div>
      </div>

      {/* Effects Section */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          padding: '2px 6px',
          background: 'rgba(0, 0, 0, 0.3)',
          borderRadius: '4px',
          border: '1px solid rgba(255, 255, 255, 0.1)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '2px',
          }}
        >
          <span
            style={{
              color: 'rgba(255, 255, 255, 0.8)',
              fontSize: '9px',
              fontWeight: 'bold',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
            }}
          >
            EFFECTS
          </span>
          <span 
            style={{
              color: '#ffffff',
              fontSize: dimensions.fontSize,
              fontWeight: 'bold',
              fontFamily: 'Courier New, monospace',
              textShadow: '1px 1px 2px rgba(0, 0, 0, 0.8)',
              lineHeight: '1',
              minWidth: '16px',
              textAlign: 'center',
              background: availableEffects > 0 ? 'rgba(255, 100, 0, 0.2)' : 'rgba(100, 100, 100, 0.2)',
              padding: '2px 4px',
              borderRadius: '3px',
              border: availableEffects > 0 ? '1px solid rgba(255, 100, 0, 0.3)' : '1px solid rgba(100, 100, 100, 0.3)',
            }}
          >
            {availableEffects}
          </span>
        </div>
      </div>
    </button>
  );
};

export default AvailableActionsDisplay;