import { useEffect, useState, type FormEvent } from 'react';
import { Dialog, DialogBackdrop, DialogPanel, DialogTitle } from '@headlessui/react';
import { Link, useNavigate } from 'react-router-dom';
import { api, ApiError, type WidgetInstance, type WidgetSize } from '../../shared/api';
import { findFreeSpot } from '../tools/findFreeSpot';
import { WIDGETS, type WidgetDefinition } from '../widgets/registry';
import { SERVICE_ICONS } from '../../shared/servicesIcons'

type Props = {
  open: boolean;
  onClose: () => void;
  widgets: WidgetInstance[];
  onCreated: (widget: WidgetInstance) => void;
  connectedServices: string[];
  editing?: WidgetInstance;
  onSaved?: (widget: WidgetInstance) => void;
  onGone?: (id: string) => void;
};

const SIZES_ORDER: WidgetSize[] = ['small', 'wide', 'tall', 'big'];

function splitDatetime(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}:\d{2})$/.exec(value);
  return match
    ? { datetime_date: `${match[3]}/${match[2]}/${match[1]}`, datetime_time: match[4] }
    : { datetime_date: '', datetime_time: '' };
}

function formatDatetime(date: string, time: string) {
  const dateMatch = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(date);
  const timeMatch = /^(\d{2}):(\d{2})$/.exec(time);
  if (!dateMatch || !timeMatch) return null;

  const day = Number(dateMatch[1]);
  const month = Number(dateMatch[2]);
  const year = Number(dateMatch[3]);
  const hours = Number(timeMatch[1]);
  const minutes = Number(timeMatch[2]);
  const parsed = new Date(Date.UTC(year, month - 1, day, hours, minutes));
  if (
    parsed.getUTCFullYear() !== year ||
    parsed.getUTCMonth() !== month - 1 ||
    parsed.getUTCDate() !== day ||
    parsed.getUTCHours() !== hours ||
    parsed.getUTCMinutes() !== minutes
  ) return null;

  return `${dateMatch[3]}-${dateMatch[2]}-${dateMatch[1]}T${time}`;
}

function valuesForParams(
  definition: WidgetDefinition,
  config: Record<string, string> = {},
) {
  const values = Object.fromEntries(definition.params.map((param) => [
    param.name,
    String(config[param.name] ?? param.defaultValue ?? ''),
  ]));
  if (definition.params.some((param) => param.name === 'datetime')) {
    Object.assign(values, splitDatetime(config.datetime ?? ''));
  }
  return values;
}

export default function AddWidgetModal({
  open,
  onClose,
  widgets,
  onCreated,
  connectedServices,
  editing,
  onSaved,
  onGone,
}: Props) {
  const navigate = useNavigate();
  const [selected, setSelected] = useState<[string, WidgetDefinition] | null>(null);
  const [values, setValues] = useState<Record<string, string>>({});
  const [refreshSeconds, setRefreshSeconds] = useState('');
  const [size, setSize] = useState<WidgetSize>('small');
  const [error, setError] = useState('');
  const [adding, setAdding] = useState(false);
  const [configurationUnavailable, setConfigurationUnavailable] = useState(false);
  const failureMessage = editing
  ? 'Unable to save this widget. Please try again.'
  : 'Unable to add this widget. Please try again.';

  useEffect(() => {
    if (open) {
      setAdding(false);
      setError('');
      setConfigurationUnavailable(false);
      if (editing) {
        const definition = WIDGETS[editing.type];
        if (!definition) {
           setSelected(null);
           setValues({});
          setRefreshSeconds('');
          setConfigurationUnavailable(true);
          setError('This widget can no longer be configured.');
        } else {
          setSelected([editing.type, definition]);
           setValues(valuesForParams(definition, editing.config));
          setRefreshSeconds(String(editing.refresh_rate_seconds));
          setSize(editing.size);
        }
      } else {
        setSelected(null);
        setValues({});
        setRefreshSeconds('');
        setSize('small');
      }
    }
  }, [open, editing]);

  const chooseWidget = (key: string, definition: WidgetDefinition) => {
    setSelected([key, definition]);
    setValues(valuesForParams(definition));
    setRefreshSeconds(String(definition.defaultRefreshSeconds));
    setSize(definition.allowedSizes[0]);
    setError('');
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected) return;
    const [type, definition] = selected;
    if (definition.params.some((param) => param.name !== 'datetime' && !values[param.name]?.trim())) {
      setError('Please complete all required fields.');
      return;
    }
    const refresh = Number(refreshSeconds);
    if (!Number.isInteger(refresh) || refresh < 60 || refresh > 21600) {
      setError('Refresh interval must be between 60 and 21600 seconds.');
      return;
    }

    const config = Object.fromEntries(definition.params.map((param) => [
      param.name,
      param.type === 'integer' ? Number(values[param.name]) : values[param.name].trim(),
    ]));
    if (definition.params.some((param) => param.name === 'datetime')) {
      const datetime = formatDatetime(values.datetime_date?.trim() ?? '', values.datetime_time?.trim() ?? '');
      if (!datetime) {
        setError('Please enter a valid date (DD/MM/YYYY) and time (HH:mm).');
        return;
      }
      config.datetime = datetime;
    }

    setError('');
    setAdding(true);
    try {
      if (editing) {
        const updated = await api.updateWidget(editing.id, {
          size,
          config,
          refresh_rate_seconds: refresh,
        });
        onSaved?.(updated);
      } else {
        const spot = findFreeSpot(widgets, size);
        if (!spot) {
          setError('No space left on your dashboard.');
          setAdding(false);
          return;
        }
        const created = await api.createWidget({
          type,
          size,
          config,
          refresh_rate_seconds: refresh,
          position: spot,
        });
        onCreated(created);
      }
      onClose();
    } catch (cause) {
      if (cause instanceof ApiError) {
        if (cause.status === 401) {
          navigate('/login');
          return;
        }
        if (cause.status === 409) setError('This spot is already taken.');
        else if (cause.status === 422) setError(cause.message);
        else if (cause.status === 404) {
        if (editing) {
          onGone?.(editing.id);
          onClose();
          return;
        }
        setError('This widget is no longer available.');
      }
        else if (cause.status === 403) setError('Connect this service first.');
        else setError(failureMessage);
      } else {
        setError(failureMessage);
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
          <DialogTitle className="mb-5 text-xl font-semibold">
            {editing ? 'Configure widget' : 'Add widget'}
          </DialogTitle>

          {error && (
            <p role="alert" className="mb-4 text-sm text-(--accent-secondary)">
              {error}
              {error === 'Connect this service first.' && (
                <> <Link className="underline" to="/subscription" onClick={onClose}>Manage services</Link></>
              )}
            </p>
          )}

          {configurationUnavailable ? null : !selected ? (
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
                        <span className="flex items-center gap-3">
                          {SERVICE_ICONS[definition.service] && (
                            <img src={SERVICE_ICONS[definition.service]} alt="" className="h-6 w-6 shrink-0 object-contain" />
                          )}
                          <span>
                            {definition.label} <span className="text-(--text-secondary)">({definition.service})</span>
                          </span>
                        </span>
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
                <button type="button" onClick={onClose} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--accent-soft)">
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <form className="space-y-4" onSubmit={(event) => void handleSubmit(event)}>
              {selected[1].params.map((param) => param.name === 'show_progress' ? (
                <div key={param.name}>
                  <label htmlFor="widget-param-show-progress" className="mb-1 block text-sm font-medium">
                    Show progress
                  </label>
                  <div className="flex items-center gap-3">
                    <button
                      id="widget-param-show-progress"
                      type="button"
                      role="switch"
                      aria-checked={values.show_progress === '1'}
                      aria-label="Show progress"
                      onClick={() => setValues((current) => ({
                        ...current,
                        show_progress: current.show_progress === '1' ? '0' : '1',
                      }))}
                      className={`relative h-6 w-11 rounded-full border border-(--border-input) transition-colors focus-visible:ring-2 focus-visible:ring-(--accent) focus-visible:ring-offset-2 ${
                        values.show_progress === '1' ? 'bg-(--accent)' : 'bg-(--border)'
                      }`}
                    >
                      <span
                        className={`absolute top-1 left-1 h-4 w-4 rounded-full bg-white transition-transform ${
                          values.show_progress === '1' ? 'translate-x-5' : ''
                        }`}
                      />
                    </button>
                    <span className="text-sm text-(--text-secondary)">
                      {values.show_progress === '1' ? 'On' : 'Off'}
                    </span>
                  </div>
                </div>
              ) : param.name === 'item_type' ? (
                <div key={param.name}>
                  <span className="mb-1 block text-sm font-medium">Show top</span>
                  <div className="flex items-center gap-3">
                    <span className="text-sm text-(--text-secondary)">Artists</span>
                    <button
                      type="button"
                      role="switch"
                      aria-checked={values.item_type === 'tracks'}
                      aria-label="Choose top tracks or artists"
                      onClick={() => setValues((current) => ({
                        ...current,
                        item_type: current.item_type === 'tracks' ? 'artists' : 'tracks',
                      }))}
                      className={`relative h-6 w-11 rounded-full border border-(--border-input) transition-colors focus-visible:ring-2 focus-visible:ring-(--accent) focus-visible:ring-offset-2 ${
                        values.item_type === 'tracks' ? 'bg-(--accent)' : 'bg-(--border)'
                      }`}
                    >
                      <span
                        className={`absolute top-1 left-1 h-4 w-4 rounded-full bg-white transition-transform ${
                          values.item_type === 'tracks' ? 'translate-x-5' : ''
                        }`}
                      />
                    </button>
                    <span className="text-sm text-(--text-secondary)">Tracks</span>
                  </div>
                </div>
              ) : param.name === 'time_range' ? (
                <div key={param.name}>
                  <label htmlFor="widget-param-time-range" className="mb-1 block text-sm font-medium">
                    Time range
                  </label>
                  <select
                    id="widget-param-time-range"
                    value={values.time_range ?? 'medium_term'}
                    onChange={(event) => setValues((current) => ({ ...current, time_range: event.target.value }))}
                    className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
                  >
                    <option value="short_term">Last 4 weeks</option>
                    <option value="medium_term">Last 6 months</option>
                    <option value="long_term">All time</option>
                  </select>
                </div>
              ) : param.name === 'datetime' ? (
                <div key={param.name}>
                  <span className="mb-1 block text-sm font-medium">Departure date and time</span>
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label htmlFor="widget-param-datetime-date" className="sr-only">Date (DD/MM/YYYY)</label>
                      <input
                        id="widget-param-datetime-date"
                        required
                        type="text"
                        inputMode="numeric"
                        placeholder="DD/MM/YYYY"
                        pattern="[0-9]{2}/[0-9]{2}/[0-9]{4}"
                        value={values.datetime_date ?? ''}
                        onChange={(event) => setValues((current) => ({ ...current, datetime_date: event.target.value }))}
                        className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
                      />
                    </div>
                    <div>
                      <label htmlFor="widget-param-datetime-time" className="sr-only">Time (HH:mm)</label>
                      <input
                        id="widget-param-datetime-time"
                        required
                        type="text"
                        inputMode="numeric"
                        placeholder="HH:mm"
                        pattern="[0-9]{2}:[0-9]{2}"
                        value={values.datetime_time ?? ''}
                        onChange={(event) => setValues((current) => ({ ...current, datetime_time: event.target.value }))}
                        className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
                      />
                    </div>
                  </div>
                </div>
              ) : (
                <div key={param.name}>
                  <label htmlFor={`widget-param-${param.name}`} className="mb-1 block text-sm font-medium">{param.label}</label>
                  <input
                    id={`widget-param-${param.name}`}
                    required
                    type={param.type === 'integer' ? 'number' : 'text'}
                    value={values[param.name] ?? ''}
                    onChange={(event) => setValues((current) => ({ ...current, [param.name]: event.target.value }))}
                    className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
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
                  className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
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
                    className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
                  >
                    {SIZES_ORDER.filter((candidate) => selected[1].allowedSizes.includes(candidate)).map((candidate) => (
                      <option key={candidate} value={candidate}>{candidate}</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex justify-between gap-2 pt-2">
                {editing ? <span /> : (
                  <button type="button" disabled={adding} onClick={() => { setSelected(null); setError(''); }} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--accent-soft) disabled:opacity-50">
                    Back
                  </button>
                )}
                <div className="flex gap-2">
                  <button type="button" disabled={adding} onClick={onClose} className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--accent-soft) disabled:opacity-50">
                    Cancel
                  </button>
                  <button type="submit" disabled={adding} className="rounded-md bg-(--accent) px-4 py-2 text-sm font-medium text-(--text-on-accent) disabled:opacity-50">
                    {editing ? (adding ? 'Saving…' : 'Save') : (adding ? 'Adding…' : 'Add')}
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
