import React from "react";
import { CardDto } from "../../../types/generated/api-types.ts";
import { UnplayableReason } from "../../../utils/cardPlayabilityUtils.ts";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";

interface UnplayableCardOverlayProps {
  card: CardDto | null;
  reason: UnplayableReason | null;
  isVisible: boolean;
}

const UnplayableCardOverlay: React.FC<UnplayableCardOverlayProps> = ({
  card,
  reason,
  isVisible,
}) => {
  // Don't render if not visible or missing data
  if (!isVisible || !card || !reason) {
    return null;
  }

  // Get the appropriate icon for the requirement type
  const getRequirementIcon = (reason: UnplayableReason) => {
    const iconMap: { [key: string]: string } = {
      // Cost
      cost: "/assets/resources/megacredit.png",
      credits: "/assets/resources/megacredit.png",
      megacredits: "/assets/resources/megacredit.png",

      // Global parameters
      temperature: "/assets/global-parameters/temperature.png",
      oxygen: "/assets/global-parameters/oxygen.png",
      oceans: "/assets/tiles/ocean.png",

      // Resources
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",

      // Tags
      science: "/assets/tags/science.png",
      "power-tag": "/assets/tags/power.png",
      space: "/assets/tags/space.png",
      building: "/assets/tags/building.png",
      city: "/assets/tags/city.png",
      earth: "/assets/tags/earth.png",
      jovian: "/assets/tags/jovian.png",
      venus: "/assets/tags/venus.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      plant: "/assets/tags/plant.png",
      event: "/assets/tags/event.png",
      mars: "/assets/tags/mars.png",
      moon: "/assets/tags/moon.png",
    };

    // For cost type, return megacredit icon
    if (reason.type === "cost") {
      return iconMap.cost;
    }

    // For global-param type, use the specific parameter type
    if (reason.type === "global-param" && reason.requirement?.type) {
      return iconMap[reason.requirement.type];
    }

    // For tag type, use the tag name
    if (reason.type === "tag" && reason.requirement?.tag) {
      const tagName = reason.requirement.tag.toLowerCase();
      // Handle special case for power tag to avoid duplicate key conflict
      if (tagName === "power") {
        return iconMap["power-tag"];
      }
      return iconMap[tagName];
    }

    // For production type, use the resource icon
    if (reason.type === "production" && reason.requirement?.resource) {
      const resourceName = reason.requirement.resource.toLowerCase();
      // For production, power refers to energy resource, not power tag
      if (resourceName === "power") {
        return iconMap["energy"];
      }
      // For production, plant refers to plants resource
      if (resourceName === "plant") {
        return iconMap["plants"];
      }
      return iconMap[resourceName];
    }

    return null;
  };

  const icon = getRequirementIcon(reason);

  // Determine which global parameters are affected
  const getAffectedParameters = (reason: UnplayableReason): string[] => {
    if (reason.type === "multiple" && reason.failedRequirements) {
      // Get all global parameters from failed requirements
      return reason.failedRequirements
        .filter((req) => req.type === "global-param" && req.requirement?.type)
        .map((req) => req.requirement.type);
    } else if (reason.type === "global-param" && reason.requirement?.type) {
      return [reason.requirement.type];
    }
    return [];
  };

  const affectedParameters = reason ? getAffectedParameters(reason) : [];

  // Create CSS class string for highlighting multiple parameters
  const createHighlightClasses = () => {
    if (affectedParameters.length === 0) return "";
    if (affectedParameters.length === 1)
      return `highlight-${affectedParameters[0]}`;

    // Multiple parameters - create multiple classes
    return affectedParameters.map((param) => `highlight-${param}`).join(" ");
  };

  return (
    <div className={`unplayable-card-overlay ${createHighlightClasses()}`}>
      {/* Top popup notification */}
      <div className="popup-notification">
        <div className="popup-content">
          <div className="popup-header">
            <span className="card-name">{card.name}</span>
            <span className="cannot-play">Cannot be played</span>
          </div>

          {reason.type === "multiple" && reason.failedRequirements ? (
            <div className="multiple-requirements">
              {reason.failedRequirements.map((req, index) => {
                const reqIcon = getRequirementIcon(req);
                return (
                  <div key={index} className="requirement-info">
                    <div className="requirement-icon-container">
                      {reqIcon && (
                        <img
                          src={reqIcon}
                          alt="requirement"
                          className="requirement-icon"
                        />
                      )}
                      {req.type === "production" && (
                        <img
                          src="/assets/misc/production.png"
                          alt="production"
                          className="production-background"
                        />
                      )}
                    </div>
                    <div className="requirement-text">
                      <p className="requirement-message">{req.message}</p>
                    </div>
                    {req.type === "cost" && (
                      <div className="cost-display">
                        <MegaCreditIcon value={card.cost} size="medium" />
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          ) : (
            <>
              <div className="requirement-info">
                <div className="requirement-icon-container">
                  {icon && (
                    <img
                      src={icon}
                      alt="requirement"
                      className="requirement-icon"
                    />
                  )}
                  {reason.type === "production" && (
                    <img
                      src="/assets/misc/production.png"
                      alt="production"
                      className="production-background"
                    />
                  )}
                </div>
                <div className="requirement-text">
                  <p className="requirement-message">{reason.message}</p>
                </div>
              </div>

              {reason.type === "cost" && (
                <div className="cost-display">
                  <MegaCreditIcon value={card.cost} size="medium" />
                </div>
              )}
            </>
          )}
        </div>
      </div>

      <style>{`
        .unplayable-card-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          z-index: 1500; /* Above Mars but below card hand */
          pointer-events: none;
          display: flex;
          justify-content: center;
        }

        .popup-notification {
          animation: slideDown 0.3s ease-out;
          margin-top: 16px;
        }

        .popup-content {
          background: linear-gradient(
            135deg,
            rgba(40, 40, 50, 0.95) 0%,
            rgba(60, 60, 70, 0.95) 100%
          );
          border: 2px solid #dc1414;
          border-radius: 12px;
          padding: 16px 20px;
          max-width: 600px;
          width: max-content;
          box-shadow:
            0 8px 32px rgba(0, 0, 0, 0.5),
            0 4px 25px rgba(220, 20, 20, 0.4),
            0 0 30px rgba(220, 20, 20, 0.6),
            0 0 60px rgba(180, 0, 0, 0.2);
          backdrop-filter: blur(8px);
        }

        .popup-header {
          display: flex;
          align-items: flex-start;
          justify-content: space-between;
          margin-bottom: 12px;
          gap: 16px;
          flex-wrap: wrap;
        }

        .card-name {
          color: #ffffff;
          font-size: 16px;
          font-weight: bold;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          flex: 1;
          min-width: 0;
          word-break: break-word;
        }

        .cannot-play {
          color: #dc1414;
          font-size: 12px;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 1px;
          white-space: nowrap;
          flex-shrink: 0;
        }

        .multiple-requirements {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .requirement-info {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .requirement-icon-container {
          position: relative;
          flex-shrink: 0;
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .requirement-icon {
          max-width: 32px;
          max-height: 32px;
          width: auto;
          height: auto;
          border-radius: 6px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
          object-fit: contain;
        }

        .production-background {
          position: absolute;
          top: 0;
          left: 0;
          width: 32px;
          height: 32px;
          z-index: -1;
        }

        .requirement-text {
          flex: 1;
        }

        .requirement-message {
          color: #ffffff;
          font-size: 14px;
          font-weight: 500;
          margin: 0;
          line-height: 1.3;
        }

        .cost-display {
          display: flex;
          justify-content: center;
          margin-top: 8px;
        }

        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        /* Responsive design */
        @media (max-width: 768px) {
          .popup-content {
            max-width: 90vw;
            padding: 14px 18px;
          }

          .card-name {
            font-size: 14px;
          }

          .cannot-play {
            font-size: 11px;
          }

          .requirement-message {
            font-size: 13px;
          }

          .requirement-icon-container {
            width: 28px;
            height: 28px;
          }

          .requirement-icon {
            max-width: 28px;
            max-height: 28px;
          }

          .production-background {
            width: 28px;
            height: 28px;
          }
        }

        @media (max-width: 480px) {
          .popup-content {
            max-width: 95vw;
            padding: 12px 16px;
          }

          .popup-header {
            flex-direction: column;
            align-items: flex-start;
            gap: 6px;
            margin-bottom: 10px;
          }

          .requirement-info {
            gap: 10px;
          }

          .requirement-icon-container {
            width: 24px;
            height: 24px;
          }

          .requirement-icon {
            max-width: 24px;
            max-height: 24px;
          }

          .production-background {
            width: 24px;
            height: 24px;
          }
        }

        /* Temperature requirement not met - highlight temperature meter, dim others */
        body:has(.unplayable-card-overlay.highlight-temperature) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-temperature) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-temperature) .currentOxygen,
        body:has(.unplayable-card-overlay.highlight-temperature) .currentOceans {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.unplayable-card-overlay.highlight-temperature) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature) .currentTemp {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Oxygen requirement not met - highlight oxygen meter, dim others */
        body:has(.unplayable-card-overlay.highlight-oxygen) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-oxygen) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-oxygen) .currentTemp,
        body:has(.unplayable-card-overlay.highlight-oxygen) .currentOceans {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.unplayable-card-overlay.highlight-oxygen) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-oxygen) .currentOxygen {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Oceans requirement not met - highlight oceans meter, dim others */
        body:has(.unplayable-card-overlay.highlight-oceans) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-oceans) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-oceans) .currentTemp,
        body:has(.unplayable-card-overlay.highlight-oceans) .currentOxygen {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.unplayable-card-overlay.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Multiple parameter highlighting support */
        /* When multiple parameters are affected, highlight all of them without dimming others */
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .currentTemp,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .currentOxygen {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        /* Temperature + Oceans combination */
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .currentTemp,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }

        /* Oxygen + Oceans combination */
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .currentOxygen,
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-oxygen.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }

        /* All three parameters combination */
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentTemp,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOxygen,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .temperatureMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.unplayable-card-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }
      `}</style>
    </div>
  );
};

export default UnplayableCardOverlay;
