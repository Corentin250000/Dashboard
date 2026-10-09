import { useEffect, useState, type FormEvent } from 'react';
import { Dialog, DialogBackdrop, DialogPanel, DialogTitle } from '@headlessui/react';
import { Link, useNavigate } from 'react-router-dom';
import { api, ApiError, type WidgetInstance, type WidgetSize } from '../../../shared/api';
import { findFreeSpot } from '../../tools/findFreeSpot';
import { WIDGETS, type WidgetDefinition } from '../../widgets/registry';
import { formatDatetime, valuesForParams } from './tools/datetime';
import WidgetForm from './WidgetForm';
import WidgetPicker from './WidgetPicker';

type Props = {
  open: boolean; onClose: () => void; widgets: WidgetInstance[]; onCreated: (widget: WidgetInstance) => void;
  connectedServices: string[]; editing?: WidgetInstance; onSaved?: (widget: WidgetInstance) => void; onGone?: (id: string) => void;
};

export default function AddWidgetModal({ open, onClose, widgets, onCreated, connectedServices, editing, onSaved, onGone }: Props) {
  const navigate = useNavigate();
  const [selected, setSelected] = useState<[string, WidgetDefinition] | null>(null);
  const [values, setValues] = useState<Record<string, string>>({});
  const [refreshSeconds, setRefreshSeconds] = useState('');
  const [size, setSize] = useState<WidgetSize>('small');
  const [error, setError] = useState('');
  const [adding, setAdding] = useState(false);
  const [configurationUnavailable, setConfigurationUnavailable] = useState(false);
  const failureMessage = editing ? 'Unable to save this widget. Please try again.' : 'Unable to add this widget. Please try again.';

  useEffect(() => {
    if (!open) return;
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
    if (definition.params.some((param) => param.field !== 'datetime' && !values[param.name]?.trim())) {
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
    if (definition.params.some((param) => param.field === 'datetime')) {
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
        const updated = await api.updateWidget(editing.id, { size, config, refresh_rate_seconds: refresh });
        onSaved?.(updated);
      } else {
        const spot = findFreeSpot(widgets, size);
        if (!spot) {
          setError('No space left on your dashboard.');
          setAdding(false);
          return;
        }
        const created = await api.createWidget({ type, size, config, refresh_rate_seconds: refresh, position: spot });
        onCreated(created);
      }
      onClose();
    } catch (cause) {
      if (cause instanceof ApiError) {
        if (cause.status === 401) { navigate('/login'); return; }
        if (cause.status === 409) setError('This spot is already taken.');
        else if (cause.status === 422) setError(cause.message);
        else if (cause.status === 404) {
          if (editing) { onGone?.(editing.id); onClose(); return; }
          setError('This widget is no longer available.');
        } else if (cause.status === 403) setError('Connect this service first.');
        else setError(failureMessage);
      } else setError(failureMessage);
    } finally {
      setAdding(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} className="relative z-50">
      <DialogBackdrop className="fixed inset-0 bg-black/40" />
      <div className="fixed inset-0 flex items-center justify-center overflow-y-auto p-4">
        <DialogPanel className="w-full max-w-lg rounded-xl border border-(--border) bg-(--card) p-6 text-(--text-primary) shadow-xl">
          <DialogTitle className="mb-5 text-xl font-semibold">{editing ? 'Configure widget' : 'Add widget'}</DialogTitle>
          {error && <p role="alert" className="mb-4 text-sm text-(--accent-secondary)">{error}{error === 'Connect this service first.' && <> <Link className="underline" to="/subscription" onClick={onClose}>Manage services</Link></>}</p>}
          {configurationUnavailable ? null : !selected ? (
            <WidgetPicker connectedServices={connectedServices} onSelect={chooseWidget} onClose={onClose} />
          ) : (
            <WidgetForm
              definition={selected[1]}
              values={values}
              refreshSeconds={refreshSeconds}
              size={size}
              editing={Boolean(editing)}
              adding={adding}
              onValueChange={(name, value) => setValues((current) => ({ ...current, [name]: value }))}
              onRefreshChange={setRefreshSeconds}
              onSizeChange={setSize}
              onSubmit={(event) => void handleSubmit(event)}
              onBack={() => { setSelected(null); setError(''); }}
              onCancel={onClose}
            />
          )}
        </DialogPanel>
      </div>
    </Dialog>
  );
}
