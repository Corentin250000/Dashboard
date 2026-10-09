import type { KeyboardEvent, ReactNode } from 'react';
import { Settings, Trash2 } from 'lucide-react';

type Props = {
  title: string;
  selected?: boolean;
  onSelect?: () => void;
  onConfigure?: () => void;
  onRemove?: () => void;
  children: ReactNode;
};

export default function WidgetCard({
  title,
  selected = false,
  onSelect,
  onConfigure,
  onRemove,
  children,
}: Props) {
  const handleKeyDown = (event: KeyboardEvent<HTMLElement>) => {
    if (event.target !== event.currentTarget) return;
    if ((event.key === 'Enter' || event.key === ' ') && onSelect) {
      event.preventDefault();
      onSelect();
    }
  };

  return (
    <article
      tabIndex={0}
      aria-current={selected ? 'true' : undefined}
      onClick={onSelect}
      onKeyDown={handleKeyDown}
      className={`flex h-full flex-col overflow-hidden rounded-xl border border-(--border) bg-(--card) ${
        selected ? 'ring-2 ring-(--accent)' : ''
      }`}
    >
      <header className="drag-handle flex cursor-grab items-center justify-between border-b border-(--border) bg-(--accent-soft) px-3 py-2">
        <h2 className="min-w-0 truncate font-medium text-(--text-primary)">{title}</h2>
        <div className="no-drag flex shrink-0 items-center gap-1">
          {onConfigure && (
            <button
              type="button"
              aria-label={`Configure ${title}`}
              onClick={(event) => {
                event.stopPropagation();
                onConfigure();
              }}
              className="rounded-md p-1.5 text-(--text-secondary) hover:bg-(--accent) hover:text-(--text-on-accent) focus-visible:ring-2 focus-visible:ring-(--accent)"
            >
              <Settings aria-hidden="true" size={18} />
            </button>
          )}
          {onRemove && (
            <button
              type="button"
              aria-label={`Remove ${title}`}
              onClick={(event) => {
                event.stopPropagation();
                onRemove();
              }}
              className="rounded-md p-1.5 text-(--text-secondary) hover:bg-(--accent) hover:text-(--text-on-accent) focus-visible:ring-2 focus-visible:ring-(--accent)"
            >
              <Trash2 aria-hidden="true" size={18} />
            </button>
          )}
        </div>
      </header>
      <div className="flex-1 overflow-auto p-3">{children}</div>
    </article>
  );
}
