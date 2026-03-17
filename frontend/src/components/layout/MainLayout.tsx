import React, { type ReactNode } from 'react';

interface LayoutProps {
  children: ReactNode;
}

const MainLayout: React.FC<LayoutProps> = ({ children }) => {
  return <>{children}</>;
};

export default MainLayout;

