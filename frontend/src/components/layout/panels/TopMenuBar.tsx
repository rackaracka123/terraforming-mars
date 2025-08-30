import React, { useState } from 'react';

interface TopMenuBarProps {}

const TopMenuBar: React.FC<TopMenuBarProps> = () => {
  const [activeTab, setActiveTab] = useState<string | null>(null);

  const menuItems = [
    { id: 'milestones', label: 'MILESTONES', color: '#ff6b35' },
    { id: 'projects', label: 'STANDARD PROJECTS', color: '#4a90e2' },
    { id: 'awards', label: 'AWARDS', color: '#f39c12' }
  ];

  const handleTabClick = (tabId: string) => {
    setActiveTab(activeTab === tabId ? null : tabId);
  };

  return (
    <div className="top-menu-bar">
      <div className="menu-container">
        <div className="menu-items">
          {menuItems.map((item) => (
            <button
              key={item.id}
              className={`menu-item ${activeTab === item.id ? 'active' : ''}`}
              onClick={() => handleTabClick(item.id)}
              style={{ '--item-color': item.color } as React.CSSProperties}
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
      
      {/* Dropdown Content */}
      {activeTab && (
        <div className="dropdown-content">
          <div className="dropdown-inner">
            {activeTab === 'milestones' && (
              <div className="content-grid">
                <div className="milestone-item">
                  <div className="milestone-title">Terraformer</div>
                  <div className="milestone-desc">Raise terraform rating to 35</div>
                  <div className="milestone-reward">5 VP</div>
                </div>
                <div className="milestone-item">
                  <div className="milestone-title">Mayor</div>
                  <div className="milestone-desc">Own 3 cities</div>
                  <div className="milestone-reward">5 VP</div>
                </div>
                <div className="milestone-item">
                  <div className="milestone-title">Gardener</div>
                  <div className="milestone-desc">Own 3 greeneries</div>
                  <div className="milestone-reward">5 VP</div>
                </div>
              </div>
            )}
            
            {activeTab === 'projects' && (
              <div className="content-grid">
                <div className="project-item">
                  <div className="project-title">Sell Patents</div>
                  <div className="project-cost">0 M€</div>
                </div>
                <div className="project-item">
                  <div className="project-title">Power Plant</div>
                  <div className="project-cost">11 M€</div>
                </div>
                <div className="project-item">
                  <div className="project-title">Asteroid</div>
                  <div className="project-cost">14 M€</div>
                </div>
              </div>
            )}
            
            {activeTab === 'awards' && (
              <div className="content-grid">
                <div className="award-item">
                  <div className="award-title">Landlord</div>
                  <div className="award-desc">Most tiles in play</div>
                  <div className="award-reward">5 VP / 2 VP</div>
                </div>
                <div className="award-item">
                  <div className="award-title">Banker</div>
                  <div className="award-desc">Highest M€ production</div>
                  <div className="award-reward">5 VP / 2 VP</div>
                </div>
                <div className="award-item">
                  <div className="award-title">Scientist</div>
                  <div className="award-desc">Most science tags</div>
                  <div className="award-reward">5 VP / 2 VP</div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
      
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
        }
        
        .menu-item.active {
          border-color: var(--item-color);
          background: var(--item-color);
          color: white;
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
        
        .dropdown-content {
          position: absolute;
          top: 100%;
          left: 0;
          right: 0;
          background: rgba(0, 0, 0, 0.95);
          border-bottom: 1px solid #333;
          max-height: 300px;
          overflow-y: auto;
        }
        
        .dropdown-inner {
          padding: 20px;
        }
        
        .content-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
          gap: 15px;
        }
        
        .milestone-item,
        .project-item,
        .award-item {
          background: rgba(255, 255, 255, 0.05);
          border-radius: 8px;
          padding: 15px;
          border-left: 4px solid #ff6b35;
        }
        
        .project-item {
          border-left-color: #4a90e2;
        }
        
        .award-item {
          border-left-color: #f39c12;
        }
        
        .milestone-title,
        .project-title,
        .award-title {
          font-weight: bold;
          margin-bottom: 5px;
          font-size: 16px;
        }
        
        .milestone-desc,
        .award-desc {
          font-size: 14px;
          color: #ccc;
          margin-bottom: 8px;
        }
        
        .milestone-reward,
        .project-cost,
        .award-reward {
          font-weight: bold;
          color: #4a90e2;
        }
      `}</style>
    </div>
  );
};

export default TopMenuBar;