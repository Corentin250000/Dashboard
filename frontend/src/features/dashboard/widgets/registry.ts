import type { ComponentType } from 'react';
import type { WidgetSize } from '../../shared/api';
import SncfDeparturesWidget from './sncf/SncfDeparturesWidget';
import SncfJourneysWidget from './sncf/SncfJourneysWidget';

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
    component: SncfDeparturesWidget,
    allowedSizes: ['big'],
    defaultRefreshSeconds: 60,
    params: [{ name: 'station', type: 'string', label: 'Station' }],
  },
  'sncf.journeys': {
    label: 'Train journeys',
    service: 'sncf',
    requiresAccount: false,
    component: SncfJourneysWidget,
    allowedSizes: ['tall', 'big'],
    defaultRefreshSeconds: 60,
    params: [
      { name: 'from', type: 'string', label: 'Departure station' },
      { name: 'to', type: 'string', label: 'Arrival station' },
      { name: 'datetime', type: 'string', label: 'Departure date and time (DD/MM/YYYY, HH:mm)' },
    ],
  },
};
