import type { WidgetDefinition } from '../../../widgets/registry';

type Props = {
  param: WidgetDefinition['params'][number];
  value: string;
  onChange: (name: string, value: string) => void;
};

export default function SwitchField({ param, value, onChange }: Props) {
  const checked = value === '1';
  const id = `widget-param-${param.name.replaceAll('_', '-')}`;

  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">{param.label}</label>
      <div className="flex items-center gap-3">
        <button
          id={id}
          type="button"
          role="switch"
          aria-checked={checked}
          onClick={() => onChange(param.name, checked ? '0' : '1')}
          className={`relative h-6 w-11 rounded-full border border-(--border-input) transition-colors focus-visible:ring-2 focus-visible:ring-(--accent) focus-visible:ring-offset-2 ${
            checked ? 'bg-(--accent)' : 'bg-(--border)'
          }`}
        >
          <span className={`absolute top-1 left-1 h-4 w-4 rounded-full bg-white transition-transform ${checked ? 'translate-x-5' : ''}`} />
        </button>
        <span className="text-sm text-(--text-secondary)">{checked ? 'On' : 'Off'}</span>
      </div>
    </div>
  );
}
