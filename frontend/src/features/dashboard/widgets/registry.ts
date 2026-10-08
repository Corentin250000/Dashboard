import type { ComponentType } from 'react';
import type { WidgetSize } from '../../shared/api';
import SncfDeparturesWidget from './sncf/SncfDeparturesWidget';
import SncfJourneysWidget from './sncf/SncfJourneysWidget';
import SpotifyCurrentlyPlayingWidget from './spotify/SpotifyCurrentlyPlayingWidget';
import SpotifyTopWidget from './spotify/SpotifyTopWidget';

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
  params: { name: string; type: 'string' | 'integer'; label: string; defaultValue?: string }[];
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
  'spotify.currently_playing': {
    label: 'Currently playing',
    service: 'spotify',
    requiresAccount: true,
    component: SpotifyCurrentlyPlayingWidget,
    allowedSizes: ['small', 'wide', 'tall', 'big'],
    defaultRefreshSeconds: 60,
    params: [
      { name: 'show_progress', type: 'integer', label: 'Show progress', defaultValue: '1' },
    ],
  },
  'spotify.top_items': {
    label: 'Spotify top items',
    service: 'spotify',
    requiresAccount: true,
    component: SpotifyTopWidget,
    allowedSizes: ['small', 'wide', 'tall', 'big'],
    defaultRefreshSeconds: 60,
    params: [
      { name: 'item_type', type: 'string', label: 'Top type', defaultValue: 'tracks' },
      { name: 'time_range', type: 'string', label: 'Time range', defaultValue: 'medium_term' },
    ],
  },
};
