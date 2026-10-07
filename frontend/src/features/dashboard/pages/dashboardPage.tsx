import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import TopBar from '../components/TopBar';
import WidgetCard from '../components/WidgetCard';
import AddWidgetModal from '../components/AddWidgetModal';
import { api, ApiError, type WidgetInstance } from '../../shared/api';
import { buildLayouts } from '../tools/buildLayouts';
import { WIDGETS } from '../widgets/registry';
import DashboardGrid from '../components/DashboardGrid';

function DashboardPage() {
  const navigate = useNavigate();
  const [widgets, setWidgets] = useState<WidgetInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [addOpen, setAddOpen] = useState(false);
  const layouts = useMemo(() => buildLayouts(widgets), [widgets]);
  const [gridKey, setGridKey] = useState(0);

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

  const removeWidget = async (id: string) => {
    setError('');
    try {
      await api.deleteWidget(id);
      setWidgets((current) => current.filter((widget) => widget.id !== id));
      setSelectedId((current) => (current === id ? null : current));
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 401) {
        navigate('/login');
        return;
      }
      setError('Unable to delete this widget.');
    }
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
      } else if (cause instanceof ApiError && (cause.status === 409 || cause.status === 422)) {
        setError('This spot is not available.');
      } else {
        setError('Unable to move this widget.');
      }
    }
  };

  return (
    <div className="flex h-dvh flex-col overflow-hidden">
      <TopBar onAddWidget={() => setAddOpen(true)} />
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
          <>
            {error && <p role="alert">{error}</p>}
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
          </>
        )}
      </main>
      <AddWidgetModal
        open={addOpen}
        onClose={() => setAddOpen(false)}
        widgets={widgets}
        onCreated={(widget) => setWidgets((current) => [...current, widget])}
        // TODO: Populate from GET /api/me/services once that endpoint is active.
        connectedServices={[]}
      />
    </div>
  );
}

export default DashboardPage;
