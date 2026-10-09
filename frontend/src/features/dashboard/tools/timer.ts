import { useContext, useEffect, useRef, useState } from 'react';
import { ApiError } from '../../shared/api';
import { WidgetRefreshContext } from './refreshContext';

export function useTimer<T>(fetcher: () => Promise<T>, intervalMs: number, restartKey: string)
{
  const { nonce, setBusy } = useContext(WidgetRefreshContext);
  const currentKey = useRef(restartKey);
  const handledNonce = useRef(nonce);
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
    const manual = handledNonce.current !== nonce;
    handledNonce.current = nonce;
    setStoredError({ key: requestKey, error: false, errorStatus: null });

    if (!requestKey) return () => {
      active = false;
      setBusy(false);
    };

    const load = (showBusy: boolean) => {
      if (showBusy) setBusy(true);
      fetcher()
        .then((value) => {
          if (!active || currentKey.current !== requestKey) return;
          setStoredData({ key: requestKey, value });
          setStoredError({ key: requestKey, error: false, errorStatus: null });
        })
        .catch((cause: unknown) => {
          if (!active || currentKey.current !== requestKey) return;
          setStoredError({
            key: requestKey,
            error: true,
            errorStatus: cause instanceof ApiError ? cause.status : null,
          });
        })
        .finally(() => {
          if (showBusy && active) setBusy(false);
        });
    };

    load(manual);
    const id = setInterval(() => load(false), intervalMs);
    return () => {
      active = false;
      clearInterval(id);
      setBusy(false);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [restartKey, intervalMs, nonce]);

  const currentError = storedError?.key === restartKey ? storedError : null;
  return {
    data: storedData?.key === restartKey ? storedData.value : null,
    error: currentError?.error ?? false,
    errorStatus: currentError?.errorStatus ?? null,
  };
}
