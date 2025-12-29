import React, { createContext, useContext } from 'react';
import type { ServiceConfig } from './types';

const ServiceConfigContext = createContext<ServiceConfig | undefined>(undefined);

export { ServiceConfig };

export function ServiceConfigProvider({
  serviceConfig,
  children,
}: {
  serviceConfig?: ServiceConfig;
  children: React.ReactNode;
}) {
  return (
    <ServiceConfigContext.Provider value={serviceConfig}>
      {children}
    </ServiceConfigContext.Provider>
  );
}

export function useServiceConfig(): ServiceConfig | undefined {
  return useContext(ServiceConfigContext);
}
