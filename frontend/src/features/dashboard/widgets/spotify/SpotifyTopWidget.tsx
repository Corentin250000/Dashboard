import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function SpotifyTopWidget({ config, refreshSeconds }: WidgetProps) {
  const itemType = config.item_type ?? 'tracks';
  const timeRange = config.time_range ?? 'medium_term';
  const { data, error } = useTimer(
    () => api.getSpotifyTopItems(itemType, timeRange),
    refreshSeconds * 1000,
    `${itemType}-${timeRange}`,
  );

  return (
    <section className="flex h-full flex-col gap-2">
      {error && (
        <p role="alert" className="text-sm text-(--accent-secondary)">
          Unable to load Spotify top items.
        </p>
      )}

      {data && data.items.length === 0 && (
        <p className="text-sm text-(--text-secondary)">No Spotify data available.</p>
      )}

      <ol className="m-0 list-none p-0">
        {data?.items.map((item) => (
          <li
            key={item.rank}
            className="flex items-center gap-3 border-b border-(--border) py-2 last:border-b-0"
          >
            <span className="w-5 shrink-0 text-sm font-medium text-(--text-secondary)">
              {item.rank}
            </span>
            <div className="min-w-0 flex-1">
              {item.spotify_url ? (
                <a
                  href={item.spotify_url}
                  target="_blank"
                  rel="noreferrer"
                  className="block truncate font-medium text-(--text-primary) hover:underline"
                >
                  {item.name}
                </a>
              ) : (
                <p className="truncate font-medium text-(--text-primary)">{item.name}</p>
              )}
              {item.artists && item.artists.length > 0 && (
                <p className="truncate text-xs text-(--text-secondary)">
                  {item.artists.join(', ')}
                </p>
              )}
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}

export default SpotifyTopWidget;
