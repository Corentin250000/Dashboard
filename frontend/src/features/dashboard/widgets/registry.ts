import type { ComponentType } from 'react';
import type { WidgetSize } from '../../shared/api';
import SncfDeparturesWidget from './sncf/SncfDeparturesWidget';
import SncfJourneysWidget from './sncf/SncfJourneysWidget';
import SpotifyCurrentlyPlayingWidget from './spotify/SpotifyCurrentlyPlayingWidget';
import SpotifyTopWidget from './spotify/SpotifyTopWidget';
import GoogleCalendarWidget from './google/GoogleCalendarWidget';
import SteamFriendsWidget from './steam/SteamFriendsWidget';
import ChessStatsWidget from './chess/ChessStatsWidget';
import ChessMonthlyStatsWidget from './chess/ChessMonthlyStatsWidget';

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
  'steam.friends': {
    label: 'Steam friends online',
    service: 'steam',
    requiresAccount: true,
    component: SteamFriendsWidget,
    allowedSizes: ['tall', 'big'],
    defaultRefreshSeconds: 60,
    params: [
      { name: 'limit', type: 'integer', label: 'Number of friends (1-100)', defaultValue: '5' },
      { name: 'inGameOnly', type: 'integer', label: 'Only friends in game', defaultValue: '0', field: 'switch' },
    ],
  },
  'chess.stats': {
    label: 'Chess rating',
    service: 'chess',
    requiresAccount: false,
    component: ChessStatsWidget,
    allowedSizes: ['wide', 'tall', 'big'],
    defaultRefreshSeconds: 300,
    params: [
      { name: 'username', type: 'string', label: 'Chess.com username' },
      {
        name: 'time_class', type: 'string', label: 'Time class', defaultValue: 'rapid', field: 'select',
        options: [
          { value: 'rapid', label: 'Rapid' },
          { value: 'blitz', label: 'Blitz' },
          { value: 'bullet', label: 'Bullet' },
        ],
      },
    ],
  },
  'chess.monthly_stats': {
    label: 'Chess monthly stats',
    service: 'chess',
    requiresAccount: false,
    component: ChessMonthlyStatsWidget,
    allowedSizes: ['tall', 'big'],
    defaultRefreshSeconds: 600,
    params: [
      { name: 'username', type: 'string', label: 'Chess.com username' },
      {
        name: 'time_class', type: 'string', label: 'Time class', defaultValue: 'rapid', field: 'select',
        options: [
          { value: 'rapid', label: 'Rapid' },
          { value: 'blitz', label: 'Blitz' },
          { value: 'bullet', label: 'Bullet' },
        ],
      },
      { name: 'year', type: 'integer', label: 'Year', defaultValue: String(new Date().getFullYear()) },
      {
        name: 'month', type: 'integer', label: 'Month', defaultValue: String(new Date().getMonth() + 1), field: 'select',
        options: [
          'January', 'February', 'March', 'April', 'May', 'June',
          'July', 'August', 'September', 'October', 'November', 'December',
        ].map((label, index) => ({ value: String(index + 1), label })),
      },
    ],
  },
};
