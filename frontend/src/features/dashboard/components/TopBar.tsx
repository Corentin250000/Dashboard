import { useEffect, useState } from 'react';
import {
  Dialog, DialogBackdrop, DialogPanel, DialogTitle,
  Menu, MenuButton, MenuItem, MenuItems,
} from '@headlessui/react';
import { Plus, UserRound } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { api, ApiError } from '../../shared/api';

type Props = {
  onAddWidget: () => void;
  onNotify?: (message: string) => void;
};

const iconButton =
  'flex size-10 items-center justify-center rounded-full text-(--text-on-accent) transition-colors hover:bg-white/15 focus-visible:ring-2 focus-visible:ring-(--text-on-accent)';
const itemClass =
  'block w-full rounded-md px-3 py-2 text-left text-sm data-focus:bg-(--accent-soft)';

function TopBar({ onAddWidget, onNotify }: Props) {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    api.me()
      .then((user) => setEmail(user.email))
      .catch(() => setEmail(''));
  }, []);

  const handleLogout = async () => {
    try {
      await api.logout();
      navigate('/login', { replace: true });
    } catch {
      onNotify?.('Logout failed. Please try again.');
    }
  };

  const handleDelete = async () => {
    setError('');
    setDeleting(true);
    try {
      await api.deleteAccount();
      navigate('/login', { replace: true });
    } catch (cause: unknown) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate('/login', { replace: true });
        return;
      }
      setError('Unable to delete your account. Please try again.');
      setDeleting(false);
    }
  };

  return (
    <header className="flex h-16 w-full items-center justify-end gap-2 border-b border-(--border) bg-(--accent) px-6 text-(--text-on-accent)">
      <button type="button" aria-label="Add widget" onClick={onAddWidget} className={iconButton}>
        <Plus size={20} aria-hidden="true" />
      </button>

      <Menu>
        <MenuButton aria-label="Open account menu" className={iconButton}>
          <UserRound size={20} aria-hidden="true" />
        </MenuButton>
        <MenuItems
          anchor="bottom end"
          className="z-40 min-w-56 rounded-lg border border-(--border) bg-(--card) p-2 shadow-lg [--anchor-gap:0.5rem] focus:outline-none"
        >
          <p className="truncate px-3 py-2 text-sm text-(--text-secondary)">
            {email || 'User account'}
          </p>
          <MenuItem>
            <Link to="/subscription" className={`${itemClass} text-(--text-primary)`}>
              Manage services
            </Link>
          </MenuItem>
          <MenuItem>
            <button type="button" onClick={() => void handleLogout()} className={`${itemClass} text-(--text-primary)`}>
              Log out
            </button>
          </MenuItem>
          <MenuItem>
            <button
              type="button"
              onClick={() => { setError(''); setConfirmOpen(true); }}
              className={`${itemClass} text-(--accent-secondary)`}
            >
              Delete account
            </button>
          </MenuItem>
        </MenuItems>
      </Menu>

      <Dialog open={confirmOpen} onClose={() => { if (!deleting) setConfirmOpen(false); }} className="relative z-50">
        <DialogBackdrop className="fixed inset-0 bg-black/40" />
        <div className="fixed inset-0 flex items-center justify-center p-4">
          <DialogPanel className="w-full max-w-sm rounded-xl border border-(--border) bg-(--card) p-6 text-(--text-primary) shadow-xl">
            <DialogTitle className="mb-3 text-lg font-semibold">Delete account</DialogTitle>
            <p className="mb-4 text-sm">
              This will permanently delete your account and all your data. This cannot be undone.
            </p>
            {error && (
              <p role="alert" className="mb-3 text-sm text-(--accent-secondary)">{error}</p>
            )}
            <div className="flex justify-end gap-2">
              <button
                type="button"
                disabled={deleting}
                onClick={() => setConfirmOpen(false)}
                className="rounded-md border border-(--border) px-4 py-2 text-sm hover:bg-(--accent-soft) disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={deleting}
                onClick={() => void handleDelete()}
                className="rounded-md bg-(--accent-secondary) px-4 py-2 text-sm font-medium text-(--text-on-accent) disabled:opacity-50"
              >
                {deleting ? 'Deleting…' : 'Delete'}
              </button>
            </div>
          </DialogPanel>
        </div>
      </Dialog>
    </header>
  );
}

export default TopBar;
