import { useState } from 'react';
import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';
import './sncf.css';

function SncfWidget() {
  const [station, setStation] = useState('');
  const [input, setInput] = useState('');

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    const cleanedInput = input.trim();
    if (cleanedInput) setStation(cleanedInput);
  };

  const { data, error } = useTimer(
    () => api.getSncfDepartures(station),
    300000,
    station
  );

  return (
    <section className="sncf-widget">
      <h2 className="sncf-title">{station || 'SNCF'}</h2>

      <form className="sncf-form" onSubmit={handleSubmit}>
        <input
          className="sncf-input"
          type="text"
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder="Station name"
          aria-label="Station name"
        />
        <button className="sncf-button" type="submit">Search</button>
      </form>

      {!station && <p>Enter a station name to see departures.</p>}
      {error && <p role="alert" className="sncf-error">Data unavailable</p>}

      <ul className="sncf-list">
        {data?.departures.map((depart, index) => (
          <li
            key={index}
            className={`sncf-row ${depart.cancelled ? 'sncf-row--cancelled' : ''}`}
          >
            {depart.delay_minutes > 0 ? (
              <span className="sncf-time-delay">{depart.departure}</span>
            ) : (
              <span className="sncf-time">{depart.departure}</span>
            )}
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