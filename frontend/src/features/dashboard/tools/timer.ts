import { useEffect, useState } from 'react';
import { ApiError } from '../../shared/api';

export function useTimer<T>(fetcher: () => Promise<T>, intervalMs: number, restartKey: string)
{
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState(false);
  const [errorStatus, setErrorStatus] = useState<number | null>(null);

  useEffect(() => {
    if (!restartKey) return;

    const load = () => {
      fetcher()
        .then((res) => {
          setData(res);
          setError(false);
          setErrorStatus(null);
        })
        .catch((cause: unknown) => {
          setError(true);
          setErrorStatus(cause instanceof ApiError ? cause.status : null);
        });
    };

    load();
    const id = setInterval(load, intervalMs);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [restartKey, intervalMs]);

  return { data, error, errorStatus };
}
