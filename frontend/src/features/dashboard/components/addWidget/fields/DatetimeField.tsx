type Props = {
  values: Record<string, string>;
  onChange: (name: string, value: string) => void;
};

export default function DatetimeField({ values, onChange }: Props) {
  return (
    <div>
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
            onChange={(event) => onChange('datetime_date', event.target.value)}
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
            onChange={(event) => onChange('datetime_time', event.target.value)}
            className="w-full rounded-md border border-(--border-input) bg-(--card) px-3 py-2 text-sm"
          />
        </div>
      </div>
    </div>
  );
}
