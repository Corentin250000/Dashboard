import type { WidgetDefinition } from '../../../widgets/registry';

type Props = {
  param: WidgetDefinition['params'][number];
  value: string;
  onChange: (name: string, value: string) => void;
};

export default function SelectField({ param, value, onChange }: Props) {
  const id = `widget-param-${param.name.replaceAll('_', '-')}`;

  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">{param.label}</label>
      <select
        id={id}
        value={value || param.defaultValue || param.options?.[0]?.value || ''}
        onChange={(event) => onChange(param.name, event.target.value)}
        className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
      >
        {param.options?.map((option) => (
          <option key={option.value} value={option.value}>{option.label}</option>
        ))}
      </select>
    </div>
  );
}
