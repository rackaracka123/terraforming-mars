import React from "react";
import { useMainContent } from "../../../contexts/MainContentContext.tsx";

const TopMenuBar: React.FC = () => {
  const { setContentType, setContentData } = useMainContent();

  const menuItems = [
    { id: "milestones" as const, label: "MILESTONES", color: "#ff6b35" },
    { id: "projects" as const, label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "awards" as const, label: "AWARDS", color: "#f39c12" },
  ];

  // Mock data for different content types - normally this would come from game state
  const getMockData = (type: "milestones" | "projects" | "awards") => {
    if (type === "milestones") {
      return {
        milestones: [
          {
            id: "terraformer",
            name: "Terraformer",
            description: "Have a terraform rating of at least 35",
            reward: "5 VP",
            cost: 8,
            claimed: false,
            available: true,
          },
          {
            id: "mayor",
            name: "Mayor",
            description: "Own at least 3 city tiles",
            reward: "5 VP",
            cost: 8,
            claimed: true,
            claimedBy: "Alice Chen",
            available: false,
          },
          {
            id: "gardener",
            name: "Gardener",
            description: "Own at least 3 greenery tiles",
            reward: "5 VP",
            cost: 8,
            claimed: false,
            available: true,
          },
        ],
      };
    }

    if (type === "projects") {
      return {
        projects: [
          {
            id: "sell-patents",
            name: "Sell Patents",
            cost: 0,
            description:
              "Discard any number of cards from hand and gain that many M€",
            available: true,
            effects: { immediate: [{ type: "credits", amount: 1 }] },
            icon: "/assets/resources/megacredit.png",
          },
          {
            id: "power-plant",
            name: "Power Plant",
            cost: 11,
            description: "Increase your energy production 1 step",
            available: true,
            effects: { production: [{ type: "energy", amount: 1 }] },
            icon: "/assets/resources/power.png",
          },
          {
            id: "city",
            name: "City",
            cost: 25,
            description: "Place a city tile",
            available: true,
            effects: { tiles: ["city"] },
            icon: "/assets/tiles/city.png",
          },
        ],
      };
    }

    if (type === "awards") {
      return {
        awards: [
          {
            id: "landlord",
            name: "Landlord",
            description: "Most tiles on Mars",
            fundingCost: 8,
            funded: true,
            fundedBy: "Bob Martinez",
            winner: "Alice Chen",
            available: false,
          },
          {
            id: "banker",
            name: "Banker",
            description: "Highest M€ production",
            fundingCost: 8,
            funded: false,
            available: true,
          },
          {
            id: "scientist",
            name: "Scientist",
            description: "Most science tags",
            fundingCost: 8,
            funded: true,
            fundedBy: "Carol Kim",
            available: false,
          },
        ],
      };
    }

    return {};
  };

  const handleTabClick = (tabId: "milestones" | "projects" | "awards") => {
    const data = getMockData(tabId);
    setContentData(data);
    setContentType(tabId);
  };

  return (
    <div className="top-menu-bar">
      <div className="menu-container">
        <div className="menu-items">
          {menuItems.map((item) => (
            <button
              key={item.id}
              className="menu-item"
              onClick={() => handleTabClick(item.id)}
              style={{ "--item-color": item.color } as React.CSSProperties}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div className="menu-actions">
          <button className="action-btn">⚙️ Settings</button>
          <button className="action-btn">📊 Stats</button>
        </div>
      </div>

      <style jsx>{`
        .top-menu-bar {
          background: rgba(0, 0, 0, 0.95);
          border-bottom: 1px solid #333;
          position: relative;
          z-index: 100;
        }

        .menu-container {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0 20px;
          height: 60px;
        }

        .menu-items {
          display: flex;
          gap: 20px;
        }

        .menu-item {
          background: none;
          border: none;
          color: white;
          font-size: 14px;
          font-weight: bold;
          padding: 10px 20px;
          cursor: pointer;
          border-radius: 4px;
          transition: all 0.2s ease;
          border: 2px solid transparent;
        }

        .menu-item:hover {
          background: rgba(255, 255, 255, 0.1);
          border-color: var(--item-color);
        }

        .menu-actions {
          display: flex;
          gap: 10px;
        }

        .action-btn {
          background: rgba(255, 255, 255, 0.1);
          border: 1px solid #333;
          color: white;
          padding: 8px 12px;
          border-radius: 4px;
          cursor: pointer;
          font-size: 12px;
        }

        .action-btn:hover {
          background: rgba(255, 255, 255, 0.2);
        }

        @media (max-width: 1024px) {
          .menu-container {
            padding: 0 15px;
            height: 50px;
          }

          .menu-item {
            font-size: 12px;
            padding: 8px 15px;
          }

          .action-btn {
            padding: 6px 10px;
            font-size: 11px;
          }
        }

        @media (max-width: 768px) {
          .menu-container {
            padding: 0 10px;
            flex-wrap: wrap;
          }

          .menu-items {
            order: 2;
            flex-basis: 100%;
            margin-top: 10px;
          }

          .menu-item {
            padding: 8px 15px;
            font-size: 12px;
          }

          .menu-actions {
            gap: 8px;
          }

          .action-btn {
            padding: 6px 8px;
            font-size: 10px;
          }
        }

        @media (max-width: 600px) {
          .menu-actions {
            flex-direction: column;
            gap: 4px;
          }

          .action-btn {
            padding: 4px 6px;
            font-size: 9px;
          }

          .menu-item {
            padding: 6px 12px;
            font-size: 11px;
          }
        }
      `}</style>
    </div>
  );
};

export default TopMenuBar;
