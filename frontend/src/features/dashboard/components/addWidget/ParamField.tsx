import type { WidgetDefinition } from '../../widgets/registry';
import DatetimeField from './fields/DatetimeField';
import SelectField from './fields/SelectField';
import SwitchField from './fields/SwitchField';
import TextField from './fields/TextField';

type Props = {
  param: WidgetDefinition['params'][number];
  values: Record<string, string>;
  onChange: (name: string, value: string) => void;
};

export default function ParamField({ param, values, onChange }: Props) {
  if (param.field === 'switch') {
    return <SwitchField param={param} value={values[param.name] ?? ''} onChange={onChange} />;
  }
  if (param.field === 'select') {
    return <SelectField param={param} value={values[param.name] ?? ''} onChange={onChange} />;
  }
  if (param.field === 'datetime') {
    return <DatetimeField values={values} onChange={onChange} />;
  }
  return <TextField param={param} value={values[param.name] ?? ''} onChange={onChange} />;
}
