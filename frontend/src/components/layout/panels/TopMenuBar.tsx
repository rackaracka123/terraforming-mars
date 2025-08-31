import React, { useState } from 'react';
import ModalPopup from '../../ui/overlay/ModalPopup.tsx';

interface TopMenuBarProps {}

const TopMenuBar: React.FC<TopMenuBarProps> = () => {
  const [activeModal, setActiveModal] = useState<'milestones' | 'projects' | 'awards' | null>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

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
        {/* Mobile Menu Toggle */}
        <button 
          className="mobile-menu-toggle"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
        >
          ☰ Menu
        </button>

        <div className={`menu-items ${mobileMenuOpen ? 'mobile-open' : ''}`}>
          {menuItems.map((item) => (
            <button
              key={item.id}
              className="menu-item"
              onClick={() => {
                handleTabClick(item.id);
                setMobileMenuOpen(false);
              }}
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

        .mobile-menu-toggle {
          display: none;
          background: none;
          border: 1px solid #333;
          color: white;
          padding: 8px 12px;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
        }

        .mobile-menu-toggle:hover {
          background: rgba(255, 255, 255, 0.1);
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
            position: relative;
          }

          .mobile-menu-toggle {
            display: block;
          }

          .menu-items {
            display: none;
            position: absolute;
            top: 100%;
            left: 0;
            right: 0;
            background: rgba(0, 0, 0, 0.98);
            border: 1px solid #333;
            border-top: none;
            flex-direction: column;
            gap: 0;
            z-index: 1000;
          }

          .menu-items.mobile-open {
            display: flex;
          }

          .menu-item {
            width: 100%;
            text-align: left;
            border-bottom: 1px solid #333;
            border-radius: 0;
            padding: 12px 20px;
            font-size: 14px;
          }

          .menu-item:last-child {
            border-bottom: none;
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

          .mobile-menu-toggle {
            font-size: 12px;
            padding: 6px 10px;
          }

          .menu-item {
            padding: 10px 15px;
            font-size: 12px;
          }
        }
      `}</style>
    </div>
  );
};

export default TopMenuBar;