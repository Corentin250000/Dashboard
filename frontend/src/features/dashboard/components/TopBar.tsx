import { useEffect, useState } from 'react';
import { UserRound } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api, ApiError } from '../../shared/api';

function TopBar() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [menuOpen, setMenuOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    api.me()
      .then((user) => setEmail(user.email))
      .catch(() => setEmail(''));
  }, []);

  const closeMenu = () => {
    setMenuOpen(false);
    setConfirmDelete(false);
    setError('');
  };

  const handleLogout = async () => {
    setError('');
    try {
      await api.logout();
      navigate('/login', { replace: true });
    } catch {
      setError('Logout failed. Please try again.');
    }
  };

  const handleDelete = async () => {
    setError('');
    setDeleting(true);
    try {
      await api.deleteAccount();
      navigate('/login', { replace: true });
    } catch (e: unknown) {
      if (e instanceof ApiError && e.status === 401) {
        navigate('/login', { replace: true }); // session expirée
        return;
      }
      setError('Unable to delete your account. Please try again.');
      setDeleting(false);
    }
  };

  return (
    <header
      className="flex h-16 w-full items-center justify-end border-b border-(--border) bg-(--card) px-6"
      onKeyDown={(e) => { if (e.key === 'Escape') closeMenu(); }}
    >
      <div className="relative">
        <button
          type="button"
          aria-label="Open account menu"
          aria-expanded={menuOpen}
          onClick={() => (menuOpen ? closeMenu() : setMenuOpen(true))}
          className="flex size-10 items-center justify-center rounded-full text-(--text-secondary) transition-colors hover:bg-(--background) hover:text-(--text-primary)"
        >
          <UserRound size={20} aria-hidden="true" />
        </button>

        {menuOpen && (
          <div className="absolute right-0 top-full z-10 mt-2 min-w-56 rounded-lg border border-(--border) bg-(--card) p-3 shadow-lg">
            <p className="mb-3 truncate text-sm text-(--text-secondary)">
              {email || 'User account'}
            </p>

            {error && (
              <p role="alert" className="mb-3 text-sm text-(--accent-secondary)">
                {error}
              </p>
            )}

            {!confirmDelete ? (
              <>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="w-full rounded-md px-3 py-2 text-left text-sm text-(--text-primary) hover:bg-(--background)"
                >
                  Log out
                </button>
                <button
                  type="button"
                  onClick={() => setConfirmDelete(true)}
                  className="w-full rounded-md px-3 py-2 text-left text-sm text-(--accent-secondary) hover:bg-(--background)"
                >
                  Delete account
                </button>
              </>
            ) : (
              <div role="alertdialog" aria-label="Confirm account deletion">
                <p className="mb-3 text-sm text-(--text-primary)">
                  This will permanently delete your account and all your data. This cannot be undone.
                </p>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => setConfirmDelete(false)}
                    disabled={deleting}
                    className="flex-1 rounded-md border border-(--border) px-3 py-2 text-sm text-(--text-primary) hover:bg-(--background) disabled:opacity-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleDelete}
                    disabled={deleting}
                    className="flex-1 rounded-md bg-(--accent-secondary) px-3 py-2 text-sm font-medium text-white disabled:opacity-50"
                  >
                    {deleting ? 'Deleting…' : 'Delete'}
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </header>
  );
}

export default TopBar;
