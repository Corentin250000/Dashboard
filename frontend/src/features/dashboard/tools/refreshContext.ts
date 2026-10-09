import { createContext } from 'react';

export type WidgetRefresh = {
  nonce: number;
  setBusy: (busy: boolean) => void;
};

export const WidgetRefreshContext = createContext<WidgetRefresh>({
  nonce: 0,
  setBusy: () => {},
});