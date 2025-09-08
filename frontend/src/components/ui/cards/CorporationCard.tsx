import React from "react";

interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingResources?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  expansion?: string;
  logoPath?: string;
}

interface CorporationCardProps {
  corporation: Corporation;
  isSelected: boolean;
  onSelect: (corporationId: string) => void;
}

const CorporationCard: React.FC<CorporationCardProps> = ({
  corporation,
  isSelected,
  onSelect,
}) => {
  const renderResourceIcon = (
    type: string,
    amount: number,
    isProduction: boolean = false,
  ) => {
    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/energy.png",
      heat: "/assets/resources/heat.png",
    };

    const icon = iconMap[type];
    if (!icon) return null;

    // For credits, use larger custom display
    if (type === "credits") {
      return (
        <div className="resource-item-large">
          {isProduction && (
            <div className="production-display-custom">
              <div className="production-icon-container">
                <img
                  src="/assets/misc/production.png"
                  alt="Production"
                  className="production-icon-xlarge"
                />
                <img
                  src={icon}
                  alt={type}
                  className="resource-icon-in-production"
                />
              </div>
              <span className="resource-amount-xlarge">{amount}</span>
            </div>
          )}
          {!isProduction && (
            <div className="resource-credits-display">
              <img src={icon} alt={type} className="resource-icon-xlarge" />
              <span className="resource-amount-xlarge">{amount}</span>
            </div>
          )}
        </div>
      );
    }

    return (
      <div className="resource-item-large">
        {isProduction && (
          <div className="production-display-custom">
            <div className="production-icon-container">
              <img
                src="/assets/misc/production.png"
                alt="Production"
                className="production-icon-xlarge"
              />
              <img
                src={icon}
                alt={type}
                className="resource-icon-in-production"
              />
            </div>
            <span className="resource-amount-xlarge">{amount}</span>
          </div>
        )}
        {!isProduction && (
          <img src={icon} alt={type} className="resource-icon-xlarge" />
        )}
        {!isProduction && (
          <span className="resource-amount-xlarge">{amount}</span>
        )}
      </div>
    );
  };

  return (
    <div
      className={`corporation-card ${isSelected ? "selected" : ""}`}
      onClick={() => onSelect(corporation.id)}
    >
      <div className="corporation-header">
        {corporation.logoPath && (
          <img
            src={corporation.logoPath}
            alt={corporation.name}
            className="corporation-logo"
          />
        )}
        <div className="corporation-info">
          <h3 className="corporation-name">{corporation.name}</h3>
          <div className="starting-credits">
            <div className="mega-credits-display">
              <img
                src="/assets/resources/megacredit.png"
                alt="Megacredits"
                className="mega-credits-icon"
              />
              <span className="mega-credits-amount">
                {corporation.startingMegaCredits}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="corporation-description">{corporation.description}</div>

      {(corporation.startingProduction || corporation.startingResources) && (
        <div className="starting-resources">
          {corporation.startingProduction && (
            <div className="production-resources">
              <h4>Starting Production:</h4>
              <div className="resources-list">
                {Object.entries(corporation.startingProduction).map(
                  ([type, amount]) =>
                    amount > 0 ? renderResourceIcon(type, amount, true) : null,
                )}
              </div>
            </div>
          )}

          {corporation.startingResources && (
            <div className="initial-resources">
              <h4>Starting Resources:</h4>
              <div className="resources-list">
                {Object.entries(corporation.startingResources).map(
                  ([type, amount]) =>
                    amount > 0 ? renderResourceIcon(type, amount, false) : null,
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {corporation.expansion && (
        <div className="expansion-badge">{corporation.expansion}</div>
      )}

      <style>{`
        .corporation-card {
          background: linear-gradient(
            135deg,
            rgba(30, 50, 80, 0.6) 0%,
            rgba(20, 40, 70, 0.5) 100%
          );
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 12px;
          padding: 20px;
          cursor: pointer;
          transition: all 0.3s ease;
          position: relative;
        }

        .corporation-card:hover {
          transform: translateY(-2px);
          box-shadow:
            0 8px 25px rgba(0, 0, 0, 0.4),
            0 0 20px rgba(100, 150, 255, 0.3);
          border-color: rgba(100, 150, 255, 0.5);
        }

        .corporation-card.selected {
          border-color: rgba(150, 255, 150, 0.8);
          box-shadow:
            0 8px 25px rgba(0, 0, 0, 0.4),
            0 0 30px rgba(150, 255, 150, 0.4);
          background: linear-gradient(
            135deg,
            rgba(30, 60, 30, 0.6) 0%,
            rgba(20, 50, 20, 0.5) 100%
          );
        }

        .corporation-header {
          display: flex;
          align-items: center;
          margin-bottom: 15px;
          gap: 15px;
        }

        .corporation-logo {
          width: 60px;
          height: 60px;
          border-radius: 8px;
          object-fit: cover;
        }

        .corporation-info {
          flex: 1;
        }

        .corporation-name {
          font-size: 20px;
          font-weight: bold;
          color: #ffffff;
          margin: 0 0 8px 0;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .starting-credits {
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(241, 196, 15, 0.2);
          padding: 12px 16px;
          border-radius: 12px;
        }

        .mega-credits-display {
          position: relative;
          display: inline-flex;
          align-items: center;
          justify-content: center;
        }

        .mega-credits-icon {
          width: 56px;
          height: 56px;
        }

        .mega-credits-amount {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          color: #000000;
          font-weight: bold;
          font-size: 18px;
          font-family: Arial, sans-serif;
          text-shadow: 0.5px 0.5px 1px rgba(255, 255, 255, 0.8);
          line-height: 1;
        }

        .corporation-description {
          font-size: 14px;
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.5;
          margin-bottom: 15px;
        }

        .starting-resources {
          margin-top: 15px;
          padding-top: 15px;
          border-top: 1px solid rgba(255, 255, 255, 0.1);
        }

        .starting-resources h4 {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.8);
          margin: 0 0 8px 0;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .resources-list {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        .resource-item {
          display: flex;
          align-items: center;
          gap: 6px;
          background: rgba(255, 255, 255, 0.1);
          padding: 4px 8px;
          border-radius: 6px;
          font-size: 12px;
          color: #ffffff;
        }

        .resource-item-large {
          display: flex;
          align-items: center;
          gap: 10px;
          background: rgba(255, 255, 255, 0.15);
          padding: 8px 12px;
          border-radius: 8px;
          font-size: 14px;
          color: #ffffff;
        }

        .resource-credits-display {
          position: relative;
          display: inline-flex;
          align-items: center;
          justify-content: center;
        }

        .resource-icon {
          width: 16px;
          height: 16px;
        }

        .resource-icon-large {
          width: 24px;
          height: 24px;
        }

        .resource-icon-xlarge {
          width: 32px;
          height: 32px;
        }

        .resource-amount {
          font-weight: bold;
        }

        .resource-amount-large {
          font-weight: bold;
          font-size: 14px;
        }

        .resource-amount-xlarge {
          font-weight: bold;
          font-size: 16px;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .production-display-custom {
          display: inline-flex;
          align-items: center;
          gap: 10px;
        }

        .production-icon-container {
          position: relative;
          display: inline-flex;
          align-items: center;
          justify-content: center;
        }

        .production-icon-xlarge {
          width: 32px;
          height: 32px;
        }

        .resource-icon-in-production {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          width: 20px;
          height: 20px;
        }

        .expansion-badge {
          position: absolute;
          top: 10px;
          right: 10px;
          background: rgba(100, 150, 255, 0.3);
          color: rgba(255, 255, 255, 0.8);
          padding: 4px 8px;
          border-radius: 4px;
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }
      `}</style>
    </div>
  );
};

export default CorporationCard;
