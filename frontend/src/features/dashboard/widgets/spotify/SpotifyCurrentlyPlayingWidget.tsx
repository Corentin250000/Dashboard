import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function formatDuration(milliseconds: number) {
  const totalSeconds = Math.max(0, Math.floor(milliseconds / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = String(totalSeconds % 60).padStart(2, '0');
  return `${minutes}:${seconds}`;
}

function SpotifyCurrentlyPlayingWidget({ config, refreshSeconds }: WidgetProps) {
  const { data, error } = useTimer(
    () => api.getSpotifyCurrentlyPlaying(),
    refreshSeconds * 1000,
    'spotify-currently-playing',
  );
  const showProgress = String(config.show_progress ?? '1') !== '0';
  const progress = data?.duration_ms
    ? Math.min(100, ((data.progress_ms ?? 0) / data.duration_ms) * 100)
    : 0;

  return (
    <section className="flex h-full flex-col gap-3">
      {error && (
        <p role="alert" className="text-sm text-(--accent-secondary)">
          Unable to load Spotify playback.
        </p>
      )}

      {data && !data.playing && (
        <p className="text-sm text-(--text-secondary)">Nothing is currently playing.</p>
      )}

      {data?.playing && (
        <div className="flex flex-col gap-2">
          <div>
            <h3 className="wrap-break-word font-medium text-(--text-primary)">{data.track_name}</h3>
            <p className="wrap-break-word text-sm text-(--text-secondary)">
              {data.artists?.join(', ')}
            </p>
            <p className="wrap-break-word text-xs text-(--text-secondary)">{data.album_name}</p>
          </div>

          {showProgress && data.duration_ms !== undefined && (
            <div className="flex flex-col gap-1">
              <div
                className="h-1.5 overflow-hidden rounded-full bg-(--border)"
                role="progressbar"
                aria-label="Track progress"
                aria-valuemin={0}
                aria-valuemax={data.duration_ms}
                aria-valuenow={data.progress_ms ?? 0}
              >
                <div
                  className="h-full rounded-full bg-(--accent)"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <div className="flex justify-between text-xs text-(--text-secondary)">
                <span>{formatDuration(data.progress_ms ?? 0)}</span>
                <span>{formatDuration(data.duration_ms)}</span>
              </div>
            </div>
          )}

          {data.spotify_url && (
            <a
              href={data.spotify_url}
              target="_blank"
              rel="noreferrer"
              className="text-sm text-(--accent) underline"
            >
              Open in Spotify
            </a>
          )}
        </div>
      )}
    </section>
  );
}

export default SpotifyCurrentlyPlayingWidget;
