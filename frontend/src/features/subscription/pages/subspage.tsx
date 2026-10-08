import { useCallback, useEffect, useState } from 'react';
import { Dialog, DialogBackdrop, DialogPanel, DialogTitle } from '@headlessui/react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { api, API_URL, ApiError, type ServiceInfo } from '../../shared/api';
import { SERVICE_ICONS } from '../../shared/servicesIcons'

function SubscriptionPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [disconnectError, setDisconnectError] = useState('');
  const [serviceToDisconnect, setServiceToDisconnect] = useState<ServiceInfo | null>(null);
  const [disconnecting, setDisconnecting] = useState(false);
  const callbackError = searchParams.has('error');

  const loadServices = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const response = await api.getServices();
      setServices(response.services);
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate('/login');
      } else {
        setError('Unable to load services. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  useEffect(() => { void loadServices(); }, [loadServices]);

  const confirmDisconnect = async () => {
    if (!serviceToDisconnect) return;
    setDisconnectError('');
    setDisconnecting(true);
    try {
      await api.disconnectService(serviceToDisconnect.name);
      setServiceToDisconnect(null);
      await loadServices();
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate('/login');
      } else {
        setDisconnectError('Unable to disconnect this service. Please try again.');
      }
    } finally {
      setDisconnecting(false);
    }
  };

  const closeDisconnectDialog = () => {
    if (disconnecting) return;
    setServiceToDisconnect(null);
    setDisconnectError('');
  };

  return (
    <div className="min-h-screen px-6 py-8">
      <div className="mx-auto max-w-270">
        <Link to="/dashboard" className="text-sm text-(--accent) underline">Back to dashboard</Link>
        <h1 className="mb-6 mt-4 text-center text-2xl font-medium text-(--text-primary)">Manage services</h1>

        {callbackError && (
          <p role="alert" className="mb-4 rounded-lg border border-(--border) bg-(--card) p-3 text-sm text-(--text-secondary)">
            We could not complete the service connection. Please try again.
          </p>
        )}
        {error && (
          <div role="alert" className="mb-4 flex items-center justify-between gap-4 rounded-lg border border-(--border) bg-(--card) p-3 text-sm">
            <span>{error}</span>
            <button type="button" onClick={() => void loadServices()} className="shrink-0 rounded-md bg-(--accent) px-3 py-1.5 font-medium text-(--text-on-accent)">
              Retry
            </button>
          </div>
        )}

        {loading ? (
          <p role="status" className="text-center text-(--text-secondary)">Loading services…</p>
        ) : !error && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {services.map((service) => (
              <article key={service.name} className="flex min-h-56 flex-col items-center justify-center gap-3 rounded-xl border border-(--border) bg-(--card) p-5 text-center">
                {SERVICE_ICONS[service.name] && (
                  <img src={SERVICE_ICONS[service.name]} alt="" className="h-10 w-10 object-contain" />
                )}
                <h2 className="text-xl font-medium text-(--text-primary)">{service.display_name}</h2>
                <p className="text-sm text-(--text-secondary)">{service.description}</p>
                <p className="text-sm text-(--text-secondary)">
                  {service.widgets.length} {service.widgets.length === 1 ? 'widget' : 'widgets'}
                </p>
                {!service.requires_connection ? (
                  <span className="rounded-full border border-(--border) px-3 py-1 text-sm text-(--text-secondary)">Included</span>
                ) : service.connected ? (
                  <div className="flex items-center gap-2">
                    <span className="rounded-full border border-(--border) px-3 py-1 text-sm text-(--text-secondary)">Connected</span>
                    <button
                      type="button"
                      onClick={() => { setDisconnectError(''); setServiceToDisconnect(service); }}
                      className="rounded-md bg-(--accent) px-3 py-1.5 text-sm font-medium text-(--text-on-accent)"
                    >
                      Disconnect
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => { window.location.href = `${API_URL}/api/me/services/${service.name}/connect`; }}
                    className="rounded-md bg-(--accent) px-4 py-2 text-sm font-medium text-(--text-on-accent)"
                  >
                    Connect
                  </button>
                )}
              </article>
            ))}
          </div>
        )}
      </div>

      <Dialog open={serviceToDisconnect !== null} onClose={closeDisconnectDialog} className="relative z-50">
        <DialogBackdrop className="fixed inset-0 bg-black/40" />
        <div className="fixed inset-0 flex items-center justify-center overflow-y-auto p-4">
          <DialogPanel className="w-full max-w-md rounded-xl border border-(--border) bg-(--card) p-6 text-(--text-primary) shadow-xl">
            <DialogTitle className="mb-3 text-lg font-semibold">
              Disconnect {serviceToDisconnect?.display_name}?
            </DialogTitle>
            <p className="mb-5 text-sm text-(--text-secondary)">Your widgets for this service will be deleted.</p>
            {disconnectError && <p role="alert" className="mb-4 text-sm text-(--accent-secondary)">{disconnectError}</p>}
            <div className="flex justify-end gap-2">
              <button type="button" disabled={disconnecting} onClick={closeDisconnectDialog} className="rounded-md border border-(--border) px-4 py-2 text-sm disabled:opacity-50">
                Cancel
              </button>
              <button type="button" disabled={disconnecting} onClick={() => void confirmDisconnect()} className="rounded-md bg-(--accent) px-4 py-2 text-sm font-medium text-(--text-on-accent) disabled:opacity-50">
                {disconnecting ? 'Disconnecting…' : 'Disconnect'}
              </button>
            </div>
          </DialogPanel>
        </div>
      </Dialog>
    </div>
  );
}

export default SubscriptionPage;
