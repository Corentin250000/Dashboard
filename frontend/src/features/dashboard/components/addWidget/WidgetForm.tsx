import type { FormEvent } from 'react';
import type { WidgetSize } from '../../../shared/api';
import type { WidgetDefinition } from '../../widgets/registry';
import ParamField from './ParamField';

type Props = {
  definition: WidgetDefinition;
  values: Record<string, string>;
  refreshSeconds: string;
  size: WidgetSize;
  editing: boolean;
  adding: boolean;
  onValueChange: (name: string, value: string) => void;
  onRefreshChange: (value: string) => void;
  onSizeChange: (size: WidgetSize) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onBack: () => void;
  onCancel: () => void;
};

const SIZES_ORDER: WidgetSize[] = ['small', 'wide', 'tall', 'big'];

export default function WidgetForm({
  definition,
  values,
  refreshSeconds,
  size,
  editing,
  adding,
  onValueChange,
  onRefreshChange,
  onSizeChange,
  onSubmit,
  onBack,
  onCancel,
}: Props) {
  return (
    <form className="space-y-4" onSubmit={onSubmit}>
      {definition.params.map((param) => (
        <ParamField key={param.name} param={param} values={values} onChange={onValueChange} />
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
          onChange={(event) => onRefreshChange(event.target.value)}
          className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
        />
        <p className="mt-1 text-xs text-(--text-secondary)">Between 60 and 21600 seconds</p>
      </div>
      {definition.allowedSizes.length > 1 && (
        <div>
          <label htmlFor="widget-size" className="mb-1 block text-sm font-medium">Size</label>
          <select
            id="widget-size"
            value={size}
            onChange={(event) => onSizeChange(event.target.value as WidgetSize)}
            className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
          >
            {SIZES_ORDER.filter((candidate) => definition.allowedSizes.includes(candidate)).map((candidate) => (
              <option key={candidate} value={candidate}>{candidate}</option>
            ))}
          </select>
        </div>
      )}
      <div className="flex justify-between gap-2 pt-2">
        {editing ? <span /> : (
          <button type="button" disabled={adding} onClick={onBack} className="rounded-md border border-(--border-input) px-4 py-2 text-sm hover:bg-(--accent-soft) disabled:opacity-50">
            Back
          </button>
        )}
        <div className="flex gap-2">
          <button type="button" disabled={adding} onClick={onCancel} className="rounded-md border border-(--border-input) px-4 py-2 text-sm hover:bg-(--accent-soft) disabled:opacity-50">
            Cancel
          </button>
          <button type="submit" disabled={adding} className="rounded-md bg-(--accent) px-4 py-2 text-sm font-medium text-(--text-on-accent) disabled:opacity-50">
            {editing ? (adding ? 'Saving…' : 'Save') : (adding ? 'Adding…' : 'Add')}
          </button>
        </div>
      </div>
    </form>
  );
}
