import { useEffect, useState } from 'react';

export function useTimer<T>(fetcher: () => Promise<T>, intervalMs: number, restartKey: string)
{
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    const load = () => {
      fetcher()
        .then((res) => {
          setData(res);
          setError(false);
        })
        .catch((err) => {
  console.error(err);
  setError(true);
});
    };

    load();
    const id = setInterval(load, intervalMs);
    return () => clearInterval(id);
  }, [restartKey, intervalMs]);

  return { data, error };
}