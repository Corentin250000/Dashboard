import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import './sncf.css';

type Props = {
  station: string;
};

function SncfWidget({ station }: Props) {
  const { data, error } = useTimer(
    () => api.getSncfDepartures(station),
    50000,
    station
  );

  return (
    <section className="sncf-widget">
      <h2 className="sncf-title">{station}</h2>
      {error && <p role="alert" className="sncf-error">Data unavailable</p>}
      <ul className="sncf-list">
        {data?.departures.map((depart, index) => (
          <li
            key={index}
            className={`sncf-row ${depart.cancelled ? 'sncf-row--cancelled' : ''}`}
          >
            {depart.delay_minutes > 0 ?  <span className="sncf-time-delay">{depart.departure}</span> :  <span className="sncf-time">{depart.departure}</span>}
            <span className="sncf-dest">{depart.destination}</span>
            {!depart.cancelled && depart.delay_minutes > 0 && (
              <span className="sncf-status sncf-status--alert">+{depart.delay_minutes} min</span>
            )}
            {depart.cancelled && (
              <span className="sncf-status sncf-status--alert">Cancelled</span>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}

export default SncfWidget;