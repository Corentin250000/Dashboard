import { api } from '../../../shared/api';
import { useTimer } from '../../tools/timer';

type Props = {
  station: string;
};

function SncfWidget({ station }: Props) {
  const { data, error } = useTimer(
    () => api.getSncfDepartures(station),
    60000,
    station
  );

  return (
    <section>
      <h2>{station}</h2>
      {error && <p role="alert">Données indisponibles</p>}
      <ul>
        {data?.departures.map((d, i) => (
          <li key={i}>
            {d.departure} - {d.destination}
          </li>
        ))}
      </ul>
    </section>
  );
}

export default SncfWidget;