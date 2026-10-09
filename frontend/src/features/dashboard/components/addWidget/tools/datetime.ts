import type { WidgetDefinition } from '../../../widgets/registry';

export function splitDatetime(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}:\d{2})$/.exec(value);
  return match
    ? { datetime_date: `${match[3]}/${match[2]}/${match[1]}`, datetime_time: match[4] }
    : { datetime_date: '', datetime_time: '' };
}

export function formatDatetime(date: string, time: string) {
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

export function valuesForParams(
  definition: WidgetDefinition,
  config: Record<string, string> = {},
) {
  const values = Object.fromEntries(definition.params.map((param) => [
    param.name,
    String(config[param.name] ?? param.defaultValue ?? ''),
  ]));
  if (definition.params.some((param) => param.field === 'datetime')) {
    Object.assign(values, splitDatetime(config.datetime ?? ''));
  }
  return values;
}
