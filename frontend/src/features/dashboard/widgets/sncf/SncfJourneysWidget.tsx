import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function SncfJourneysWidget({ config, refreshSeconds }: WidgetProps) {
  const from = config.from ?? '';
  const to = config.to ?? '';
  const datetime = config.datetime ?? '';

  const { data, error } = useTimer(
    () => api.getSncfJourneys(from, to, datetime),
    refreshSeconds * 1000,
    `${from}-${to}-${datetime}`,
  );

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium text-(--text-primary)">
        {from} <span className="text-(--text-secondary)">→</span> {to}
      </h3>

      {error && (
        <p role="alert" className="text-sm text-(--accent-secondary)">
          Data unavailable
        </p>
      )}

      {data && data.journeys.length === 0 && (
        <p className="text-sm text-(--text-secondary)">No journey found.</p>
      )}

      <ul className="m-0 list-none p-0">
        {data?.journeys.map((journey, index) => (
          <li
            key={`${journey.departure_time}-${journey.arrival_time}-${index}`}
            className="border-b border-(--border) py-2 last:border-b-0"
          >
            <div className="flex flex-col gap-1">
              <span className="font-medium text-(--text-primary)">
                {journey.departure_time} → {journey.arrival_time}
              </span>
              <span className="text-xs text-(--text-secondary)">
                {journey.transfers === 0
                  ? 'Direct journey'
                  : `${journey.transfers} transfer${journey.transfers > 1 ? 's' : ''}`}
              </span>
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
}

export default SncfJourneysWidget;
