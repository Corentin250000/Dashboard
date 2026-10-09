import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

const ERROR_MESSAGES: Record<number, string> = {
  404: 'Player not found.',
  400: 'Invalid settings.',
  502: 'Chess.com is unavailable.',
};

function ChessStatsWidget({ config, refreshSeconds }: WidgetProps) {
  const username = String(config.username ?? '').trim();
  const timeClass = String(config.time_class || 'rapid');
  const { data, error, errorStatus } = useTimer(
    () => api.getChessStats(username, timeClass),
    refreshSeconds * 1000,
    `${username}|${timeClass}`,
  );

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
          <p dir="auto" className="truncate font-medium text-(--text-primary)">{data.username}</p>
          <p className="text-xs capitalize text-(--text-secondary)">{data.time_class}</p>

          <p className="mt-2 text-3xl font-semibold text-(--text-primary)">{data.rating}</p>
          <p className="text-sm text-(--text-secondary)">Best: {data.best_rating}</p>

          <dl className="mt-3 grid grid-cols-3 gap-2 text-center text-sm">
            <div>
              <dt className="text-xs text-(--text-secondary)">Wins</dt>
              <dd className="font-medium text-(--text-primary)">{data.record.wins}</dd>
            </div>
            <div>
              <dt className="text-xs text-(--text-secondary)">Draws</dt>
              <dd className="font-medium text-(--text-primary)">{data.record.draws}</dd>
            </div>
            <div>
              <dt className="text-xs text-(--text-secondary)">Losses</dt>
              <dd className="font-medium text-(--text-primary)">{data.record.losses}</dd>
            </div>
          </dl>

          <p className="mt-2 text-sm text-(--text-secondary)">
            Win rate: {data.record.win_rate.toFixed(1)}% ({data.record.games} games)
          </p>
          <section className="mt-3">
            <h3 className="text-sm font-medium text-(--text-primary)">This month</h3>
            {data.month_record.games === 0 ? (
              <p className="text-sm text-(--text-secondary)">No {timeClass} games this month.</p>
            ) : (
              <>
                <p className="text-sm text-(--text-secondary)">
                  Win rate this month: {data.month_record.win_rate.toFixed(1)}% ({data.month_record.games} games)
                </p>
                <p className="text-sm text-(--text-secondary)">
                  {data.month_record.wins}W / {data.month_record.draws}D / {data.month_record.losses}L
                </p>
              </>
            )}
          </section>
          <p className="mt-1 text-xs text-(--text-secondary)">
            Updated {new Date(data.updated_at * 1000).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
          </p>
        </div>
      )}
    </section>
  );
}

export default ChessStatsWidget;
