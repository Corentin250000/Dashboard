import type { ComponentType } from 'react';
import type { WidgetSize } from '../../shared/api';
import SncfWidget from './sncf/SncfWidget';

export type WidgetProps = {
  config: Record<string, string>;
  refreshSeconds: number;
};

export type WidgetDefinition = {
  label: string;
  service: string;
  requiresAccount: boolean;
  component: ComponentType<WidgetProps>;
  allowedSizes: WidgetSize[];
  defaultRefreshSeconds: number;
  params: { name: string; type: 'string' | 'integer'; label: string }[];
};

export const WIDGETS: Record<string, WidgetDefinition> = {
  'sncf.departures': {
    label: 'Train departures',
    service: 'sncf',
    requiresAccount: false,
    component: SncfWidget,
    allowedSizes: ['big'],
    defaultRefreshSeconds: 60,
    params: [{ name: 'station', type: 'string', label: 'Station' }],
  },
};
