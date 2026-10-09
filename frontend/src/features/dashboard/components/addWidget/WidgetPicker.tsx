import { Link } from 'react-router-dom';
import { SERVICE_ICONS } from '../../../shared/servicesIcons';
import { WIDGETS, type WidgetDefinition } from '../../widgets/registry';

type Props = {
  connectedServices: string[];
  onSelect: (key: string, definition: WidgetDefinition) => void;
  onClose: () => void;
};

export default function WidgetPicker({ connectedServices, onSelect, onClose }: Props) {
  return (
    <div className="space-y-3">
      <ul className="space-y-2">
        {Object.entries(WIDGETS).map(([key, definition]) => {
          const locked = definition.requiresAccount && !connectedServices.includes(definition.service);
          return (
            <li key={key} className="rounded-lg border border-(--border) p-3">
              <button
                type="button"
                disabled={locked}
                onClick={() => onSelect(key, definition)}
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
        <button type="button" onClick={onClose} className="rounded-md border border-(--border-input) px-4 py-2 text-sm hover:bg-(--accent-soft)">
          Cancel
        </button>
      </div>
    </div>
  );
}
