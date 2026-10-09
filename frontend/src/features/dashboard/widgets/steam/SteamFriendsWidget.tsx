import { Link } from 'react-router-dom';
import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function SteamFriendsWidget({ config, refreshSeconds }: WidgetProps) {
  const limit = Number(config.limit) || 5;
  const inGameOnly = String(config.inGameOnly) === '1';
  const { data, error, errorStatus } = useTimer(
    () => api.getSteamFriends(limit, inGameOnly),
    refreshSeconds * 1000,
    `${limit}|${inGameOnly}`,
  );

  return (
    <section className="flex h-full min-h-0 flex-col overflow-hidden">
      {!data && !error && (
        <p role="status" className="text-sm text-(--text-secondary)">Loading…</p>
      )}

      {error && (
        <p role="alert" className="mb-2 shrink-0 text-sm text-(--accent-secondary)">
          {errorStatus === 403 ? 'This friend list is private.' : errorStatus === 400 ? (
            <>Connect Steam first. <Link to="/subscription" className="text-(--accent) underline">Manage services</Link></>
          ) : 'Data unavailable'}
        </p>
      )}

      {data && (
        <>
          <p className="mb-2 shrink-0 text-xs text-(--text-secondary)">
            {data.returned} of {data.totalOnline} online
          </p>
          {data.friends.length === 0 ? (
            !error && <p className="text-sm text-(--text-secondary)">{inGameOnly ? 'No friends in game.' : 'No friends online.'}</p>
          ) : (
            <ul className="m-0 min-h-0 flex-1 list-none overflow-x-hidden overflow-y-auto p-0">
              {data.friends.map((friend) => (
                <li key={friend.steamId} className="min-w-0 border-b border-(--border) py-2 last:border-b-0">
                  <p dir="auto" className="truncate font-medium text-(--text-primary)">{friend.pseudo}</p>
                  <p className="wrap-break-word text-sm text-(--text-secondary)">{friend.status}</p>
                  {friend.currentGame && (
                    <p className="truncate text-xs text-(--text-secondary)">Playing {friend.currentGame.name}</p>
                  )}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  );
}

export default SteamFriendsWidget;
