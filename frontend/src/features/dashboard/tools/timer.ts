import { useEffect, useRef, useState } from 'react';
import { ApiError } from '../../shared/api';

export function useTimer<T>(fetcher: () => Promise<T>, intervalMs: number, restartKey: string)
{
  const currentKey = useRef(restartKey);
  currentKey.current = restartKey;
  const [storedData, setStoredData] = useState<{ key: string; value: T } | null>(null);
  const [storedError, setStoredError] = useState<{
    key: string;
    error: boolean;
    errorStatus: number | null;
  } | null>(null);

  useEffect(() => {
    const requestKey = restartKey;
    let active = true;
    setStoredError({ key: requestKey, error: false, errorStatus: null });

    if (!requestKey) return () => { active = false; };

    const load = () => {
      fetcher()
        .then((res) => {
          if (!active || currentKey.current !== requestKey) return;
          setStoredData({ key: requestKey, value: res });
          setStoredError({ key: requestKey, error: false, errorStatus: null });
        })
        .catch((cause: unknown) => {
          if (!active || currentKey.current !== requestKey) return;
          setStoredError({
            key: requestKey,
            error: true,
            errorStatus: cause instanceof ApiError ? cause.status : null,
          });
        });
    };

    load();
    const id = setInterval(load, intervalMs);
    return () => {
      active = false;
      clearInterval(id);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [restartKey, intervalMs]);

  const currentError = storedError?.key === restartKey ? storedError : null;
  return {
    data: storedData?.key === restartKey ? storedData.value : null,
    error: currentError?.error ?? false,
    errorStatus: currentError?.errorStatus ?? null,
  };
}
