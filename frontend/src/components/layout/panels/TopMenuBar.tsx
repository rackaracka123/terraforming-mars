import React, { useState } from 'react';
import ModalPopup from '../../ui/overlay/ModalPopup.tsx';

interface TopMenuBarProps {}

const TopMenuBar: React.FC<TopMenuBarProps> = () => {
  const [activeModal, setActiveModal] = useState<'milestones' | 'projects' | 'awards' | null>(null);

  const menuItems = [
    { id: 'milestones' as const, label: 'MILESTONES', color: '#ff6b35' },
    { id: 'projects' as const, label: 'STANDARD PROJECTS', color: '#4a90e2' },
    { id: 'awards' as const, label: 'AWARDS', color: '#f39c12' }
  ];

  const handleTabClick = (tabId: 'milestones' | 'projects' | 'awards') => {
    setActiveModal(tabId);
  };

  const handleCloseModal = () => {
    setActiveModal(null);
  };

  const handleModalAction = (actionType: string, itemId: string) => {
    console.log(`Modal Action: ${actionType} on ${itemId}`);
    // Handle the action here - for now just log it
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
      
      {/* Modal Popup */}
      <ModalPopup 
        type={activeModal}
        onClose={handleCloseModal}
        onAction={handleModalAction}
      />
      
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
      `}</style>
    </div>
  );
};

export default TopMenuBar;