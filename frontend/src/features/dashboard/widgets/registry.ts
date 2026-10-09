import type { ComponentType } from 'react';
import type { WidgetSize } from '../../shared/api';
import SncfDeparturesWidget from './sncf/SncfDeparturesWidget';
import SncfJourneysWidget from './sncf/SncfJourneysWidget';
import SpotifyCurrentlyPlayingWidget from './spotify/SpotifyCurrentlyPlayingWidget';
import SpotifyTopWidget from './spotify/SpotifyTopWidget';
import GoogleCalendarWidget from './google/GoogleCalendarWidget';

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
  params: {
    name: string;
    type: 'string' | 'integer';
    label: string;
    defaultValue?: string;
    field?: 'text' | 'select' | 'switch' | 'datetime';
    options?: { value: string; label: string }[];
  }[];
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
      { name: 'datetime', type: 'string', label: 'Departure date and time (DD/MM/YYYY, HH:mm)', field: 'datetime' },
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
      { name: 'show_progress', type: 'integer', label: 'Show progress', defaultValue: '1', field: 'switch' },
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
      {
        name: 'item_type', type: 'string', label: 'Show top', defaultValue: 'tracks', field: 'select',
        options: [{ value: 'tracks', label: 'Tracks' }, { value: 'artists', label: 'Artists' }],
      },
      {
        name: 'time_range', type: 'string', label: 'Time range', defaultValue: 'medium_term', field: 'select',
        options: [
          { value: 'short_term', label: 'Last 4 weeks' },
          { value: 'medium_term', label: 'Last 6 months' },
          { value: 'long_term', label: 'All time' },
        ],
      },
    ],
  },
  'calendar.upcoming_events': {
    label: 'Calendar events',
    service: 'calendar',
    requiresAccount: true,
    component: GoogleCalendarWidget,
    allowedSizes: ['tall', 'big'],
    defaultRefreshSeconds: 300,
    params: [
      { name: 'maxResults', type: 'integer', label: 'Number of events (1-50)', defaultValue: '10' },
      {
        name: 'period', type: 'string', label: 'Period', defaultValue: 'day', field: 'select',
        options: [{ value: 'day', label: 'Next 24 hours' }, { value: 'week', label: 'Next 7 days' }],
      },
    ],
  },
};
