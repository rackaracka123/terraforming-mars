import React from "react";
import { useMainContent } from "../../../contexts/MainContentContext.tsx";
import styles from "./TopMenuBar.module.css";

const TopMenuBar: React.FC = () => {
  const { setContentType, setContentData } = useMainContent();

  const menuItems = [
    { id: "milestones" as const, label: "MILESTONES", color: "#ff6b35" },
    { id: "projects" as const, label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "awards" as const, label: "AWARDS", color: "#f39c12" },
    { id: "debug" as const, label: "DEBUG", color: "#9b59b6" },
  ];

  // Mock data for different content types - normally this would come from game state
  const getMockData = (
    type: "milestones" | "projects" | "awards" | "debug",
  ) => {
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

  const handleTabClick = (
    tabId: "milestones" | "projects" | "awards" | "debug",
  ) => {
    // For debug, we'll emit a custom event instead of using content context
    if (tabId === "debug") {
      window.dispatchEvent(new CustomEvent("toggle-debug-dropdown"));
      return;
    }
    const data = getMockData(tabId);
    setContentData(data);
    setContentType(tabId);
  };

  return (
    <div className={styles.topMenuBar}>
      <div className={styles.menuContainer}>
        <div className={styles.menuItems}>
          {menuItems.map((item) => (
            <button
              key={item.id}
              className={styles.menuItem}
              onClick={() => handleTabClick(item.id)}
              style={{ "--item-color": item.color } as React.CSSProperties}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div className={styles.menuActions}>
          <button className={styles.actionBtn}>⚙️ Settings</button>
          <button className={styles.actionBtn}>📊 Stats</button>
        </div>
      </div>
    </div>
  );
};

export default TopMenuBar;
