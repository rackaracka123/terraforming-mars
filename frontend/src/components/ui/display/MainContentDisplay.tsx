import React from "react";
import Game3DView from "../../game/view/Game3DView.tsx";
import { useMainContent } from "@/contexts/MainContentContext.tsx";
import CostDisplay from "./CostDisplay.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface Milestone {
  id: string;
  name: string;
  description: string;
  reward: string;
  cost: number;
  claimed: boolean;
  claimedBy?: string;
  available: boolean;
}

interface Award {
  id: string;
  name: string;
  description: string;
  fundingCost: number;
  funded: boolean;
  fundedBy?: string;
  winner?: string;
  available: boolean;
}

interface MainContentDisplayProps {
  gameState: GameDto;
}

const MainContentDisplay: React.FC<MainContentDisplayProps> = ({ gameState }) => {
  const { contentType, contentData, setContentType } = useMainContent();

  const renderMilestones = () => {
    const milestones: Milestone[] = (contentData?.milestones as Milestone[]) || [];

    return (
      <div className="main-content-container">
        <div className="content-header">
          <button className="back-button" onClick={() => setContentType("game")}>
            ← Back to Game
          </button>
          <h2>Milestones</h2>
          <div className="subtitle">Claim milestones to earn victory points</div>
        </div>

        <div className="items-grid">
          {milestones.map((milestone) => (
            <div
              key={milestone.id}
              className={`item-card milestone-card ${milestone.claimed ? "claimed" : ""} ${!milestone.available ? "unavailable" : ""}`}
            >
              <div className="item-header">
                <div className="item-name">{milestone.name}</div>
                <div className="item-cost">
                  <CostDisplay cost={milestone.cost} size="small" />
                </div>
              </div>
              <div className="item-description">{milestone.description}</div>
              <div className="item-reward">Reward: {milestone.reward}</div>
              {milestone.claimed && milestone.claimedBy && (
                <div className="claimed-by">Claimed by {milestone.claimedBy}</div>
              )}
              <div className="item-actions">
                <button
                  className="action-btn claim-btn"
                  disabled={milestone.claimed || !milestone.available}
                >
                  {milestone.claimed ? "Claimed" : "Claim"}
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const renderAwards = () => {
    const awards: Award[] = (contentData?.awards as Award[]) || [];

    return (
      <div className="main-content-container">
        <div className="content-header">
          <button className="back-button" onClick={() => setContentType("game")}>
            ← Back to Game
          </button>
          <h2>Awards</h2>
          <div className="subtitle">Fund awards and compete for victory points</div>
        </div>

        <div className="items-grid">
          {awards.map((award) => (
            <div
              key={award.id}
              className={`item-card award-card ${award.funded ? "funded" : ""} ${!award.available ? "unavailable" : ""}`}
            >
              <div className="item-header">
                <div className="item-name">{award.name}</div>
                <div className="item-cost">
                  <CostDisplay cost={award.fundingCost} size="small" />
                </div>
              </div>
              <div className="item-description">{award.description}</div>
              <div className="award-info">
                <div className="award-rewards">1st place: 5 VP, 2nd place: 2 VP</div>
                {award.funded && award.fundedBy && (
                  <div className="funded-by">Funded by {award.fundedBy}</div>
                )}
                {award.winner && <div className="current-winner">Leading: {award.winner}</div>}
              </div>
              <div className="item-actions">
                <button className="action-btn fund-btn" disabled={award.funded || !award.available}>
                  {award.funded ? "Funded" : "Fund"}
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  if (contentType === "game") {
    return <Game3DView gameState={gameState} />;
  }

  return (
    <div className="main-content-wrapper">
      {contentType === "milestones" && renderMilestones()}
      {contentType === "awards" && renderAwards()}

      <style>{`
        .main-content-wrapper {
          width: 100%;
          height: 100%;
          background: linear-gradient(
            135deg,
            rgba(5, 10, 25, 0.95) 0%,
            rgba(10, 20, 35, 0.9) 50%,
            rgba(5, 15, 30, 0.95) 100%
          );
          overflow-y: auto;
          position: relative;
        }

        .main-content-container {
          padding: 20px;
          max-width: 1200px;
          margin: 0 auto;
          height: 100%;
        }

        .content-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 30px;
          padding-bottom: 20px;
          border-bottom: 2px solid rgba(100, 150, 255, 0.3);
          flex-wrap: wrap;
          gap: 10px;
        }

        .back-button {
          background: linear-gradient(
            135deg,
            rgba(100, 150, 255, 0.8) 0%,
            rgba(50, 100, 200, 0.9) 100%
          );
          border: 2px solid rgba(100, 150, 255, 0.6);
          border-radius: 8px;
          color: #ffffff;
          font-size: 14px;
          font-weight: bold;
          cursor: pointer;
          padding: 10px 16px;
          transition: all 0.3s ease;
          box-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
        }

        .back-button:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 25px rgba(100, 150, 255, 0.5);
        }

        .content-header h2 {
          margin: 0;
          color: #ffffff;
          font-size: 28px;
          font-weight: bold;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          flex: 1;
          text-align: center;
        }

        .cards-count,
        .actions-count {
          color: rgba(255, 255, 255, 0.8);
          font-size: 16px;
          font-weight: 500;
          background: rgba(100, 150, 255, 0.2);
          padding: 8px 16px;
          border-radius: 20px;
          border: 1px solid rgba(100, 150, 255, 0.3);
        }

        .subtitle {
          color: rgba(255, 255, 255, 0.7);
          font-size: 16px;
          text-align: center;
          flex-basis: 100%;
        }

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
          gap: 20px;
          padding: 20px;
          justify-items: center;
        }

        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 60px 20px;
          text-align: center;
          grid-column: 1 / -1;
        }

        .empty-icon {
          font-size: 64px;
          margin-bottom: 20px;
          opacity: 0.6;
        }

        .empty-state h3 {
          color: #ffffff;
          font-size: 24px;
          margin-bottom: 10px;
        }

        .empty-state p {
          color: rgba(255, 255, 255, 0.7);
          font-size: 16px;
          margin: 0;
        }

        .card {
          width: 100%;
          max-width: 200px;
          aspect-ratio: 5 / 7; /* Playing card aspect ratio */
          border: 2px solid;
          border-radius: 12px;
          padding: 16px;
          position: relative;
          backdrop-filter: blur(10px);
          animation: cardSlideIn 0.6s ease-out both;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          display: flex;
          flex-direction: column;
          justify-content: space-between;
          overflow: hidden;
        }

        .card:hover {
          transform: translateY(-8px) scale(1.02);
        }

        .card-type-badge {
          position: absolute;
          top: 8px;
          right: 8px;
          background: rgba(0, 0, 0, 0.8);
          color: #ffffff;
          padding: 4px 8px;
          border-radius: 8px;
          font-size: 9px;
          font-weight: bold;
          text-transform: uppercase;
          letter-spacing: 0.3px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .card-cost {
          position: absolute;
          top: 8px;
          left: 8px;
        }

        .card-content {
          margin-top: 20px;
          flex: 1;
          display: flex;
          flex-direction: column;
          justify-content: space-between;
        }

        .card-name {
          color: #ffffff;
          font-size: 14px;
          font-weight: bold;
          margin: 0 0 8px 0;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          line-height: 1.2;
          text-align: center;
        }

        .card-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 11px;
          line-height: 1.3;
          margin: 0;
          background: rgba(0, 0, 0, 0.3);
          padding: 8px;
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          flex: 1;
          overflow-y: auto;
          text-align: left;
        }

        .actions-content {
          display: flex;
          flex-direction: column;
          gap: 40px;
        }

        .actions-section {
          margin-bottom: 40px;
        }

        .section-title {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 20px 0;
          padding-bottom: 10px;
          border-bottom: 2px solid rgba(100, 150, 255, 0.3);
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .actions-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
          gap: 20px;
          justify-items: center;
        }

        .action-card {
          width: 100%;
          max-width: 320px;
          min-height: 200px;
          background: linear-gradient(
            145deg,
            rgba(30, 50, 80, 0.6) 0%,
            rgba(20, 40, 70, 0.5) 100%
          );
          border: 2px solid rgba(100, 150, 255, 0.3);
          border-radius: 15px;
          padding: 20px;
          position: relative;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
          animation: actionSlideIn 0.6s ease-out both;
        }

        .action-card.available {
          cursor: pointer;
        }

        .action-card.available:hover {
          transform: translateY(-8px) scale(1.02);
          box-shadow:
            0 12px 40px rgba(0, 0, 0, 0.4),
            0 0 50px rgba(100, 150, 255, 0.4);
        }

        .action-card.unavailable {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .action-type-badge {
          position: absolute;
          top: 15px;
          right: 15px;
          background: rgba(0, 0, 0, 0.8);
          color: #ffffff;
          padding: 6px 12px;
          border-radius: 12px;
          font-size: 11px;
          font-weight: bold;
          letter-spacing: 0.5px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .action-cost {
          position: absolute;
          top: 15px;
          left: 15px;
        }

        .action-cost.unavailable-cost {
          opacity: 0.6;
        }

        .action-content {
          margin-top: 35px;
          margin-bottom: 15px;
        }

        .action-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 8px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          line-height: 1.3;
        }

        .action-source {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          font-style: italic;
          margin-bottom: 10px;
        }

        .action-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.5;
          margin: 0 0 10px 0;
          background: rgba(0, 0, 0, 0.3);
          padding: 12px;
          border-radius: 8px;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .action-requirement {
          color: rgba(255, 200, 100, 0.9);
          font-size: 12px;
          line-height: 1.4;
          background: rgba(255, 200, 100, 0.1);
          padding: 8px 12px;
          border-radius: 6px;
          border: 1px solid rgba(255, 200, 100, 0.3);
        }

        .action-cost-reward {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          margin: 10px 0;
          padding: 8px;
          background: rgba(0, 0, 0, 0.2);
          border-radius: 8px;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .cost-section,
        .reward-section {
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .resource-item {
          display: flex;
          align-items: center;
          gap: 3px;
          background: rgba(255, 255, 255, 0.1);
          padding: 4px 6px;
          border-radius: 4px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .resource-icon {
          width: 16px;
          height: 16px;
          object-fit: contain;
          filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.8));
        }

        .resource-amount {
          color: #ffffff;
          font-size: 12px;
          font-weight: bold;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .arrow-section {
          display: flex;
          align-items: center;
          padding: 0 4px;
        }

        .arrow-icon {
          width: 20px;
          height: 12px;
          object-fit: contain;
          filter: brightness(1.2) drop-shadow(0 1px 3px rgba(0, 0, 0, 0.8));
        }

        .action-button {
          width: 100%;
          background: linear-gradient(
            135deg,
            rgba(100, 150, 255, 0.8) 0%,
            rgba(50, 100, 200, 0.9) 100%
          );
          border: 2px solid rgba(100, 150, 255, 0.6);
          border-radius: 8px;
          color: #ffffff;
          font-size: 14px;
          font-weight: bold;
          cursor: pointer;
          padding: 10px 16px;
          transition: all 0.3s ease;
          box-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
        }

        .action-button:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 25px rgba(100, 150, 255, 0.5);
        }

        .items-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
          gap: 20px;
        }

        .item-card {
          background: linear-gradient(
            135deg,
            rgba(30, 50, 80, 0.6) 0%,
            rgba(20, 40, 70, 0.5) 100%
          );
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 12px;
          padding: 20px;
          transition: all 0.3s ease;
          position: relative;
        }

        .milestone-card {
          border-left-color: #ff6b35;
        }

        .project-card {
          border-left-color: #4a90e2;
        }

        .award-card {
          border-left-color: #f39c12;
        }

        .item-card:hover:not(.unavailable) {
          transform: translateY(-2px);
          box-shadow:
            0 8px 25px rgba(0, 0, 0, 0.4),
            0 0 20px rgba(100, 150, 255, 0.3);
        }

        .item-card.claimed,
        .item-card.funded {
          border-color: rgba(150, 255, 150, 0.5);
          background: linear-gradient(
            135deg,
            rgba(30, 60, 30, 0.6) 0%,
            rgba(20, 50, 20, 0.5) 100%
          );
        }

        .item-card.unavailable {
          opacity: 0.5;
          border-color: rgba(255, 150, 150, 0.3);
        }

        .item-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
        }

        .project-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
        }

        .project-icon-name {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .project-icon {
          width: 24px;
          height: 24px;
          filter: drop-shadow(0 1px 3px rgba(0, 0, 0, 0.8));
        }

        .item-name {
          font-size: 18px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .item-cost {
          font-size: 16px;
          font-weight: bold;
          color: #f1c40f;
          background: rgba(241, 196, 15, 0.2);
          padding: 4px 8px;
          border-radius: 6px;
        }

        .item-description {
          font-size: 14px;
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.5;
          margin-bottom: 12px;
        }

        .item-reward {
          font-size: 12px;
          color: rgba(150, 255, 150, 0.9);
          margin-bottom: 8px;
          font-weight: 500;
        }

        .award-info {
          margin-bottom: 12px;
        }

        .award-rewards {
          font-size: 12px;
          color: rgba(150, 255, 150, 0.9);
          margin-bottom: 4px;
        }

        .claimed-by,
        .funded-by,
        .current-winner {
          font-size: 11px;
          color: rgba(100, 200, 255, 0.8);
          font-style: italic;
        }

        .item-actions {
          display: flex;
          justify-content: flex-end;
          margin-top: 15px;
        }

        .action-btn {
          padding: 8px 16px;
          border: none;
          border-radius: 6px;
          font-weight: bold;
          cursor: pointer;
          transition: all 0.2s ease;
          font-size: 14px;
        }

        .claim-btn {
          background: linear-gradient(135deg, #ff6b35 0%, #ff8c42 100%);
          color: white;
        }

        .claim-btn:hover:not(:disabled) {
          background: linear-gradient(135deg, #e55a2b 0%, #ff6b35 100%);
          transform: translateY(-1px);
        }

        .play-btn {
          background: linear-gradient(135deg, #4a90e2 0%, #5ba0f2 100%);
          color: white;
        }

        .play-btn:hover:not(:disabled) {
          background: linear-gradient(135deg, #357abd 0%, #4a90e2 100%);
          transform: translateY(-1px);
        }

        .fund-btn {
          background: linear-gradient(135deg, #f39c12 0%, #f1c40f 100%);
          color: white;
        }

        .fund-btn:hover:not(:disabled) {
          background: linear-gradient(135deg, #d68910 0%, #f39c12 100%);
          transform: translateY(-1px);
        }

        .action-btn:disabled {
          background: rgba(100, 100, 100, 0.5) !important;
          color: rgba(255, 255, 255, 0.5);
          cursor: not-allowed;
          transform: none !important;
        }

        @keyframes cardSlideIn {
          from {
            opacity: 0;
            transform: translateY(30px) scale(0.9);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        @keyframes actionSlideIn {
          from {
            opacity: 0;
            transform: translateY(30px) scale(0.9);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        @media (max-width: 768px) {
          .main-content-container {
            padding: 15px;
          }

          .content-header {
            margin-bottom: 20px;
            padding-bottom: 15px;
          }

          .content-header h2 {
            font-size: 22px;
            text-align: left;
          }

          .cards-grid {
            grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
            gap: 15px;
            padding: 15px;
          }

          .card {
            max-width: 160px;
            padding: 12px;
          }

          .card-name {
            font-size: 12px;
          }

          .card-description {
            font-size: 10px;
            padding: 6px;
          }

          .actions-grid,
          .items-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .action-card,
          .item-card {
            max-width: 100%;
            min-height: auto;
            padding: 15px;
          }

          .back-button {
            font-size: 12px;
            padding: 8px 12px;
          }
        }

        @media (max-width: 480px) {
          .cards-grid {
            grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
            gap: 12px;
            padding: 12px;
          }

          .card {
            max-width: 140px;
            padding: 10px;
          }

          .card-name {
            font-size: 11px;
          }

          .card-description {
            font-size: 9px;
            padding: 5px;
          }

          .card-type-badge {
            font-size: 8px;
            padding: 3px 6px;
          }
        }
      `}</style>
    </div>
  );
};

export default MainContentDisplay;
