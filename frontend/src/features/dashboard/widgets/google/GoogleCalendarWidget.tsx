import { useEffect, useState } from 'react';
import { MapPin } from 'lucide-react';
import { Link } from 'react-router-dom';
import { api, type CalendarEvent } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function parseEventDate(value: string) {
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (dateOnly) {
    return new Date(Number(dateOnly[1]), Number(dateOnly[2]) - 1, Number(dateOnly[3]));
  }
  return new Date(value);
}

function dateKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

function formatDuration(start: Date, end: Date) {
  const minutes = Math.max(0, Math.round((end.getTime() - start.getTime()) / 60000));
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (hours && remainingMinutes) return `${hours}h ${remainingMinutes} min`;
  if (hours) return `${hours}h`;
  return `${minutes} min`;
}

type EventRowProps = {
  event: CalendarEvent;
  now: Date;
};

function EventRow({ event, now }: EventRowProps) {
  const start = parseEventDate(event.start);
  const end = parseEventDate(event.end);
  const isCurrent = !event.allDay && start <= now && now < end;
  const isPast = end < now;
  const textColor = isPast ? 'text-(--text-secondary)' : 'text-(--text-primary)';

  return (
    <li className={`flex min-w-0 gap-3 border-b border-(--border) py-2 last:border-b-0 ${
      isCurrent ? 'rounded-lg bg-(--accent-soft) px-2' : ''
    } ${textColor}`}>
      <div className="flex shrink-0 flex-wrap items-baseline gap-x-2 gap-y-1">
        {event.allDay ? (
          <span className="whitespace-nowrap rounded-full border border-(--border) bg-(--accent-soft) px-2 py-0.5 text-xs font-medium text-(--accent)">
            All day
          </span>
        ) : (
          <span className="whitespace-nowrap text-sm font-medium">
            {`${start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} - ${end.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`}
          </span>
        )}
        {!event.allDay && (
          <span className="text-xs text-(--text-secondary)">{formatDuration(start, end)}</span>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex min-w-0 items-center gap-2">
          <p className={`min-w-0 flex-1 truncate text-sm font-medium ${textColor}`}>{event.title}</p>
          {isCurrent && (
            <span className="shrink-0 rounded-full bg-(--accent) px-2 py-0.5 text-xs font-medium text-(--text-on-accent)">
              Now
            </span>
          )}
        </div>
        {event.location && (
          <p className="flex min-w-0 items-center gap-1 truncate text-xs text-(--text-secondary)">
            <MapPin size={14} aria-hidden="true" className="shrink-0" />
            <span className="truncate">{event.location}</span>
          </p>
        )}
      </div>
    </li>
  );
}

function GoogleCalendarWidget({ config, refreshSeconds }: WidgetProps) {
  const maxResults = Number(config.maxResults) || 10;
  const period = config.period === 'week' ? 'week' : 'day';
  const { data, error, errorStatus } = useTimer(
    () => api.getCalendarEvents(maxResults, period),
    refreshSeconds * 1000,
    `${maxResults}|${period}`,
  );
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const intervalId = setInterval(() => setNow(new Date()), 30000);
    return () => clearInterval(intervalId);
  }, []);

  const groups = data?.reduce<{ key: string; date: Date; events: CalendarEvent[] }[]>((current, event) => {
    const start = parseEventDate(event.start);
    const key = dateKey(start);
    const existing = current.find((group) => group.key === key);
    if (existing) existing.events.push(event);
    else current.push({ key, date: start, events: [event] });
    return current;
  }, []);
  const showHeadings = (groups?.length ?? 0) > 1;
  const today = dateKey(now);
  const tomorrowDate = new Date(now);
  tomorrowDate.setDate(tomorrowDate.getDate() + 1);
  const tomorrow = dateKey(tomorrowDate);

  return (
    <section className="flex h-full min-h-0 flex-col overflow-hidden">
      {!data && !error && (
        <p role="status" className="text-sm text-(--text-secondary)">Loading…</p>
      )}

      {error && (
        <p role="alert" className="mb-2 shrink-0 text-sm text-(--accent-secondary)">
          {errorStatus === 403 ? (
            <>Connect Google first. <Link to="/subscription" className="text-(--accent) underline">Manage services</Link></>
          ) : 'Data unavailable'}
        </p>
      )}

      {data && data.length === 0 && !error && (
        <p className="text-sm text-(--text-secondary)">No upcoming events.</p>
      )}

      {data && data.length > 0 && groups && (
        <ul className="m-0 min-h-0 flex-1 list-none overflow-x-hidden overflow-y-auto p-0">
          {showHeadings ? groups.map((group) => (
            <li key={group.key} className="min-w-0 list-none">
              <h3 className="py-2 text-sm font-semibold text-(--text-primary)">
                {group.key === today
                  ? 'Today'
                  : group.key === tomorrow
                    ? 'Tomorrow'
                    : group.date.toLocaleDateString()}
              </h3>
              <ul className="m-0 list-none p-0">
                {group.events.map((event) => <EventRow key={event.id} event={event} now={now} />)}
              </ul>
            </li>
          )) : data.map((event) => <EventRow key={event.id} event={event} now={now} />)}
        </ul>
      )}
    </section>
  );
}

export default GoogleCalendarWidget;
