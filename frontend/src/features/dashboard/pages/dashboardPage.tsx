import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import TopBar from '../components/TopBar';
import WidgetCard from '../components/WidgetCard';
import AddWidgetModal from '../components/AddWidgetModal';
import { api, ApiError, type ServiceInfo, type WidgetInstance } from '../../shared/api';
import { buildLayouts } from '../tools/buildLayouts';
import { WIDGETS } from '../widgets/registry';
import DashboardGrid from '../components/DashboardGrid';
import Notifications from '../components/Notification'

function DashboardPage() {
  const navigate = useNavigate();
  const [widgets, setWidgets] = useState<WidgetInstance[]>([]);
  const [services, setServices] = useState<ServiceInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const [editing, setEditing] = useState<WidgetInstance | null>(null);
  const layouts = useMemo(() => buildLayouts(widgets), [widgets]);
  const connectedServices = useMemo(
    () => services
      .filter((service) => service.connected || !service.requires_connection)
      .map((service) => service.name),
    [services],
  );
  const [gridKey, setGridKey] = useState(0);
  const [notice, setNotice] = useState('');
  const dismissNotice = useCallback(() => setNotice(''), []);

  const loadWidgets = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const response = await api.getWidgets();
      setWidgets(response.widgets);
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate('/login');
      } else {
        setError('Unable to load your dashboard.');
      }
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  useEffect(() => { void loadWidgets(); }, [loadWidgets]);

  const loadServices = useCallback(async () => {
    try {
      const response = await api.getServices();
      setServices(response.services);
    } catch {
      setServices([]);
    }
  }, []);

  useEffect(() => {
    void loadServices();
    window.addEventListener('focus', loadServices);
    return () => window.removeEventListener('focus', loadServices);
  }, [loadServices]);
  const removeWidget = async (id: string) => {
  try {
    await api.deleteWidget(id);
    forgetWidget(id);
  } catch (cause) {
    if (cause instanceof ApiError && cause.status === 401) {
      navigate('/login');
    } else if (cause instanceof ApiError && cause.status === 404) {
      forgetWidget(id);
      setNotice('This widget no longer exists.');
    } else {
      setNotice('Unable to delete this widget.');
    }
  }
};

const forgetWidget = (id: string) => {
  setWidgets((current) => current.filter((widget) => widget.id !== id));
  setSelectedId((current) => (current === id ? null : current));
  setEditing((current) => (current?.id === id ? null : current));
};

const moveWidget = async (id: string, x: number, y: number) => {
  const widget = widgets.find((item) => item.id === id);
  if (!widget || (widget.position.x === x && widget.position.y === y)) return;
  try {
    await api.updateWidget(id, { position: { x, y } });
    setWidgets((current) => current.map((item) =>
      item.id === id ? { ...item, position: { x, y } } : item,
    ));
  } catch (cause) {
    setGridKey((k) => k + 1);
    if (cause instanceof ApiError && cause.status === 401) {
      navigate('/login');
    } else if (cause instanceof ApiError && cause.status === 404) {
      forgetWidget(id);
      setNotice('This widget no longer exists.');
    } else if (cause instanceof ApiError && (cause.status === 409 || cause.status === 422)) {
      setNotice('This spot is not available.');
    } else {
      setNotice('Unable to move this widget.');
    }
  }
};

  return (
    <div className="flex h-dvh flex-col overflow-hidden">
      <TopBar onNotify={setNotice} onAddWidget={() => setAddOpen(true)} />
      <main className="min-h-0 flex-1">
        {loading ? (
          <p role="status">Loading…</p>
        ) : error && widgets.length === 0 ? (
          <p role="alert">{error} <button type="button" onClick={() => void loadWidgets()}>Retry</button></p>
        ) : widgets.length === 0 ? (
          <div className="flex h-full items-center justify-center text-(--text-secondary)">
            <p>Start to create your dashboard with the + button!</p>
          </div>
        ) : (
            <DashboardGrid key={gridKey} layouts={layouts} onDragStop={moveWidget}>
              {widgets.map((widget) => {
              const def = WIDGETS[widget.type];
              const Widget = def?.component;
              return (
                <div key={widget.id}>
                  <WidgetCard
                    title={def?.label ?? widget.type}
                    selected={selectedId === widget.id}
                    onSelect={() => setSelectedId((c) => (c === widget.id ? null : widget.id))}
                    onConfigure={() => setEditing(widget)}
                    onRemove={() => void removeWidget(widget.id)}
                  >
                    {Widget ? (
                      <Widget config={widget.config} refreshSeconds={widget.refresh_rate_seconds} />
                    ) : (
                      <p>Unknown widget type.</p>
                    )}
                  </WidgetCard>
                </div>
              );
            })}
            </DashboardGrid>
        )}
      </main>
      <AddWidgetModal
        open={addOpen || editing !== null}
        editing={editing ?? undefined}
        onClose={() => {
          setAddOpen(false);
          setEditing(null);
        }}
        widgets={widgets}
        onCreated={(widget) => setWidgets((current) => [...current, widget])}
        onSaved={(widget) => setWidgets((current) => current.map((item) => (item.id === widget.id ? widget : item)))}
        onGone={(id) => {
          forgetWidget(id);
          setNotice('This widget no longer exists.')
        }}
        connectedServices={connectedServices}
      />
      <Notifications message={notice} onClose={dismissNotice} />
    </div>
  );
}

export default DashboardPage;
