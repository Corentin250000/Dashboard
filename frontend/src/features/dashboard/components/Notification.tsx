import { useEffect } from 'react';
import { X } from 'lucide-react';

type Props = { message: string; onClose: () => void };

export default function Notifications({ message, onClose }: Props) {
  useEffect(() => {
    if (!message) return;
    const timer = setTimeout(onClose, 5000);
    return () => clearTimeout(timer);
  }, [message, onClose]);

  if (!message) return null;

  return (
    <div
      role="alert"
      className="fixed bottom-4 left-1/2 z-50 flex max-w-[90vw] -translate-x-1/2 items-center gap-3 rounded-lg border border-(--border) bg-(--card) px-4 py-3 text-sm text-(--text-primary) shadow-lg"
    >
      <span>{message}</span>
      <button
        type="button"
        aria-label="Dismiss notification"
        onClick={onClose}
        className="rounded p-1 text-(--text-secondary) hover:bg-(--background) focus-visible:ring-2 focus-visible:ring-(--accent)"
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  );
}