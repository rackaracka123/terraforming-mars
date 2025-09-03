import React, { createContext, useContext, useState, ReactNode } from "react";

export type MainContentType =
  | "game"
  | "played-cards"
  | "available-actions"
  | "milestones"
  | "projects"
  | "awards";

interface MainContentContextType {
  contentType: MainContentType;
  setContentType: (type: MainContentType) => void;
  contentData: any;
  setContentData: (data: any) => void;
}

const MainContentContext = createContext<MainContentContextType | undefined>(
  undefined,
);

export const MainContentProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const [contentType, setContentType] = useState<MainContentType>("game");
  const [contentData, setContentData] = useState<any>(null);

  return (
    <MainContentContext.Provider
      value={{
        contentType,
        setContentType,
        contentData,
        setContentData,
      }}
    >
      {children}
    </MainContentContext.Provider>
  );
};

export const useMainContent = (): MainContentContextType => {
  const context = useContext(MainContentContext);
  if (!context) {
    throw new Error("useMainContent must be used within a MainContentProvider");
  }
  return context;
};
