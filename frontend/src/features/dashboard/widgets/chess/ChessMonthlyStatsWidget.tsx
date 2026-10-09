import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

const ERROR_MESSAGES: Record<number, string> = {
  404: 'Player not found.',
  400: 'Invalid settings.',
  502: 'Chess.com is unavailable.',
};

function ChessMonthlyStatsWidget({ config, refreshSeconds }: WidgetProps) {
  const username = String(config.username ?? '').trim();
  const year = Number(config.year);
  const month = Number(config.month);
  const timeClass = String(config.time_class || 'rapid');
  const { data, error, errorStatus } = useTimer(
    () => api.getChessMonthlyStats(username, year, month, timeClass),
    refreshSeconds * 1000,
    `${username}|${year}|${month}|${timeClass}`,
  );

  const games = data
    ? [...data.games].sort((a, b) => b.date.localeCompare(a.date))
    : [];

  return (
    <section className="flex h-full min-h-0 flex-col overflow-hidden">
      {!data && !error && (
        <p role="status" className="text-sm text-(--text-secondary)">Loading…</p>
      )}

      {error && (
        <p role="alert" className="mb-2 shrink-0 text-sm text-(--accent-secondary)">
          {ERROR_MESSAGES[errorStatus ?? 0] ?? 'Data unavailable'}
        </p>
      )}

      {data && (
        <div className="min-h-0 flex-1 overflow-x-hidden overflow-y-auto">
          {data.rating !== undefined && (
            <div className="mb-3">
              <p className="text-xs text-(--text-secondary)">Current rating ({timeClass})</p>
              <p className="text-3xl font-semibold text-(--text-primary)">{data.rating}</p>
              <p className="text-sm text-(--text-secondary)">Best: {data.best_rating}</p>
            </div>
          )}
          <p dir="auto" className="truncate font-medium text-(--text-primary)">
            {data.username} · {String(data.month).padStart(2, '0')}/{data.year}
          </p>
          <p className="text-sm text-(--text-secondary)">
            {data.total_games} games: {data.wins}W / {data.draws}D / {data.losses}L
          </p>
          {data.average_accuracy != null && (
            <p className="text-sm text-(--text-secondary)">Average accuracy: {data.average_accuracy.toFixed(1)}%</p>
          )}

          {data.openings.length > 0 && (
            <>
              <h3 className="mt-3 text-xs font-medium uppercase text-(--text-secondary)">Top openings</h3>
              <ol className="m-0 list-none p-0">
                {data.openings.map((opening) => (
                  <li key={opening.name} className="flex justify-between gap-2 text-sm text-(--text-primary)">
                    <span className="truncate">{opening.name}</span>
                    <span className="shrink-0 text-(--text-secondary)">{opening.games}</span>
                  </li>
                ))}
              </ol>
            </>
          )}

          {games.length > 0 && (
            <>
              <h3 className="mt-3 text-xs font-medium uppercase text-(--text-secondary)">Games</h3>
              <ul className="m-0 list-none p-0">
                {games.map((game) => (
                  <li key={game.url} className="border-b border-(--border) py-2 text-sm last:border-b-0">
                    <p className="flex justify-between gap-2">
                      <span className="font-medium capitalize text-(--text-primary)">{game.result}</span>
                      <span className="text-xs capitalize text-(--text-secondary)">{game.color}</span>
                    </p>
                    <p className="text-xs text-(--text-secondary)">
                      {new Date(game.date).toLocaleDateString(undefined, { dateStyle: 'medium' })} · {game.termination}
                      {game.accuracy != null && ` · ${game.accuracy.toFixed(1)}%`}
                    </p>
                    <a
                      href={game.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-xs text-(--accent) underline"
                    >
                      View game
                    </a>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </section>
  );
}

export default ChessMonthlyStatsWidget;
