import type { WidgetDefinition } from '../../../widgets/registry';

type Props = {
  param: WidgetDefinition['params'][number];
  value: string;
  onChange: (name: string, value: string) => void;
};

export default function TextField({ param, value, onChange }: Props) {
  return (
    <div>
      <label htmlFor={`widget-param-${param.name}`} className="mb-1 block text-sm font-medium">{param.label}</label>
      <input
        id={`widget-param-${param.name}`}
        required
        type={param.type === 'integer' ? 'number' : 'text'}
        value={value}
        onChange={(event) => onChange(param.name, event.target.value)}
        className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
      />
    </div>
  );
}
