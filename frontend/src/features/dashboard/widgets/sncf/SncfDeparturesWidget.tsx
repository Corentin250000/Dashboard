import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import type { WidgetProps } from '../registry';

function SncfDeparturesWidget({ config, refreshSeconds }: WidgetProps) {
  const station = config.station ?? '';

  const { data, error } = useTimer(
    () => api.getSncfDepartures(station),
    refreshSeconds * 1000,
    station
  );

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium text-(--text-primary)">{station}</h3>

      {error && (
        <p role="alert" className="text-sm text-(--accent-secondary)">
          Data unavailable
        </p>
      )}

      <ul className="m-0 list-none p-0">
        {data?.departures.map((depart, index) => (
          <li
            key={index}
            className={`grid grid-cols-[auto_1fr_auto] items-center gap-2 border-b border-(--border) py-2 last:border-b-0 ${
              depart.cancelled ? 'text-(--text-secondary) line-through' : 'text-(--text-primary)'
            }`}
          >
            <span className="flex items-center gap-1 whitespace-nowrap font-medium">
              {depart.delay_minutes > 0 && !depart.cancelled && (
                <span className="text-(--text-secondary) line-through">
                  {depart.scheduled_departure}
                </span>
              )}
              <span
                className={depart.delay_minutes > 0 && !depart.cancelled ? 'text-(--accent-secondary)' : ''}
              >
                {depart.departure}
              </span>
            </span>
            <span className="truncate">{depart.destination}</span>
            {depart.cancelled && (
              <span className="text-[13px] font-medium text-(--accent-secondary) no-underline">
                Cancelled
              </span>
            )}
            {!depart.cancelled && depart.delay_minutes > 0 && (
              <span className="text-[13px] font-medium text-(--accent-secondary)">
                +{depart.delay_minutes} min
              </span>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}

export default SncfDeparturesWidget;
