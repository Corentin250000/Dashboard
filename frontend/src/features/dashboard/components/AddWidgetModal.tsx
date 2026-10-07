import { useEffect, useState, type FormEvent } from 'react';
import { Dialog, DialogBackdrop, DialogPanel, DialogTitle } from '@headlessui/react';
import { Link, useNavigate } from 'react-router-dom';
import { api, ApiError, type WidgetInstance, type WidgetSize } from '../../shared/api';
import { findFreeSpot } from '../tools/findFreeSpot';
import { WIDGETS, type WidgetDefinition } from '../widgets/registry';

type Props = {
  open: boolean;
  onClose: () => void;
  widgets: WidgetInstance[];
  onCreated: (widget: WidgetInstance) => void;
  connectedServices: string[];
};

const SIZES_ORDER: WidgetSize[] = ['small', 'wide', 'tall', 'big'];

export default function AddWidgetModal({ open, onClose, widgets, onCreated, connectedServices }: Props) {
  const navigate = useNavigate();
  const [selected, setSelected] = useState<[string, WidgetDefinition] | null>(null);
  const [values, setValues] = useState<Record<string, string>>({});
  const [refreshSeconds, setRefreshSeconds] = useState('');
  const [size, setSize] = useState<WidgetSize>('small');
  const [error, setError] = useState('');
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    if (open) {
      setSelected(null);
      setValues({});
      setRefreshSeconds('');
      setSize('small');
      setError('');
      setAdding(false);
    }
  }, [open]);

  const chooseWidget = (key: string, definition: WidgetDefinition) => {
    setSelected([key, definition]);
    setValues(Object.fromEntries(definition.params.map((param) => [param.name, ''])));
    setRefreshSeconds(String(definition.defaultRefreshSeconds));
    setSize(definition.allowedSizes[0]);
    setError('');
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected) return;
    const [type, definition] = selected;
    if (definition.params.some((param) => !values[param.name]?.trim())) {
      setError('Please complete all required fields.');
      return;
    }
    const refresh = Number(refreshSeconds);
    if (!Number.isInteger(refresh) || refresh < 60 || refresh > 21600) {
      setError('Refresh interval must be between 60 and 21600 seconds.');
      return;
    }

    const spot = findFreeSpot(widgets, size);
    if (!spot) {
      setError('No space left on your dashboard.');
      return;
    }

    // The backend config contract uses strings, including integer parameters; confirm this with the backend.
    const config = Object.fromEntries(definition.params.map((param) => [
      param.name,
      param.type === 'integer' ? String(Number(values[param.name])) : values[param.name].trim(),
    ]));

    setError('');
    setAdding(true);
    try {
      const created = await api.createWidget({
        type,
        size,
        config,
        refresh_rate_seconds: refresh,
        position: spot,
      });
      onCreated(created);
      onClose();
    } catch (cause) {
      if (cause instanceof ApiError) {
        if (cause.status === 401) {
          navigate('/login');
          return;
        }
        if (cause.status === 409) setError('This spot is already taken.');
        else if (cause.status === 422) setError(cause.message);
        else if (cause.status === 404) setError('This widget is no longer available.');
        else if (cause.status === 403) setError('Connect this service first.');
        else setError('Unable to add this widget. Please try again.');
      } else {
        setError('Unable to add this widget. Please try again.');
      }
    } finally {
      setAdding(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/40" />
      <div className="fixed inset-0 flex items-center justify-center overflow-y-auto p-4">
        <DialogPanel className="w-full max-w-lg rounded-xl border border-(--border) bg-(--card) p-6 text-(--text-primary) shadow-xl">
          <DialogTitle className="mb-5 text-xl font-semibold">Add widget</DialogTitle>

          {error && (
            <p role="alert" className="mb-4 text-sm text-(--accent-secondary)">
              {error}
              {error === 'Connect this service first.' && (
                <> <Link className="underline" to="/subscription" onClick={onClose}>Manage services</Link></>
              )}
            </p>
          )}

          {!selected ? (
            <div className="space-y-3">
              <ul className="space-y-2">
                {Object.entries(WIDGETS).map(([key, definition]) => {
                  const locked = definition.requiresAccount && !connectedServices.includes(definition.service);
                  return (
                    <li key={key} className="rounded-lg border border-(--border) p-3">
                      <button
                        type="button"
                        disabled={locked}
                        onClick={() => chooseWidget(key, definition)}
                        className="w-full text-left text-sm font-medium disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {definition.label} <span className="text-(--text-secondary)">({definition.service})</span>
                      </button>
                      {locked && (
                        <p className="mt-1 text-sm text-(--text-secondary)">
                          Connect this service first.{' '}
                          <Link className="text-(--accent) underline" to="/subscription" onClick={onClose}>Connect service</Link>
                        </p>
                      )}
                    </li>
                  );
                })}
              </ul>
              <Link className="inline-block text-sm text-(--accent) underline" to="/subscription" onClick={onClose}>
                Manage services
              </Link>
              <div className="flex justify-end">
                <button type="button" onClick={onClose} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--background)">
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <form className="space-y-4" onSubmit={(event) => void handleSubmit(event)}>
              {selected[1].params.map((param) => (
                <div key={param.name}>
                  <label htmlFor={`widget-param-${param.name}`} className="mb-1 block text-sm font-medium">{param.label}</label>
                  <input
                    id={`widget-param-${param.name}`}
                    required
                    type={param.type === 'integer' ? 'number' : 'text'}
                    value={values[param.name] ?? ''}
                    onChange={(event) => setValues((current) => ({ ...current, [param.name]: event.target.value }))}
                    className="w-full rounded-md border border-(--border) bg-(--card) px-3 py-2 text-sm"
                  />
                </div>
              ))}

              <div>
                <label htmlFor="widget-refresh" className="mb-1 block text-sm font-medium">Refresh interval (seconds)</label>
                <input
                  id="widget-refresh"
                  required
                  type="number"
                  min={60}
                  max={21600}
                  value={refreshSeconds}
                  onChange={(event) => setRefreshSeconds(event.target.value)}
                  className="w-full rounded-md border border-(--border) bg-(--card) px-3 py-2 text-sm"
                />
                <p className="mt-1 text-xs text-(--text-secondary)">Between 60 and 21600 seconds</p>
              </div>

              {selected[1].allowedSizes.length > 1 && (
                <div>
                  <label htmlFor="widget-size" className="mb-1 block text-sm font-medium">Size</label>
                  <select
                    id="widget-size"
                    value={size}
                    onChange={(event) => setSize(event.target.value as WidgetSize)}
                    className="w-full rounded-md border border-(--border) bg-(--card) px-3 py-2 text-sm"
                  >
                    {SIZES_ORDER.filter((candidate) => selected[1].allowedSizes.includes(candidate)).map((candidate) => (
                      <option key={candidate} value={candidate}>{candidate}</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex justify-between gap-2 pt-2">
                <button type="button" disabled={adding} onClick={() => { setSelected(null); setError(''); }} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--background) disabled:opacity-50">
                  Back
                </button>
                <div className="flex gap-2">
                  <button type="button" disabled={adding} onClick={onClose} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--background) disabled:opacity-50">
                    Cancel
                  </button>
                  <button type="submit" disabled={adding} className="rounded-md bg-(--accent) px-4 py-2 text-sm font-medium text-(--text-on-accent) disabled:opacity-50">
                    {adding ? 'Adding…' : 'Add'}
                  </button>
                </div>
              </div>
            </form>
          )}
        </DialogPanel>
      </div>
    </Dialog>
  );
}
