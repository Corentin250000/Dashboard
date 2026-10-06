import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import TopBar from '../components/TopBar';
import DashboardGrid, { type Layouts } from '../components/DashboardGrid';
import Card from '../components/Card';
import { api, ApiError, type WidgetInstance } from '../../shared/api';
import { buildLayouts } from '../tools/buildLayouts';

function DashboardPage() {
  const navigate = useNavigate();
  const [widgets, setWidgets] = useState<WidgetInstance[]>([]);
  const [layouts, setLayouts] = useState<Layouts>({ lg: [], md: [], sm: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);

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
  useEffect(() => { setLayouts(buildLayouts(widgets)); }, [widgets]);

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
      setLayouts(buildLayouts(widgets));
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
      <TopBar />
      <main className="min-h-0 flex-1">
        {loading ? (
          <p role="status">Loading…</p>
        ) : error && widgets.length === 0 ? (
          <p role="alert">{error} <button type="button" onClick={() => void loadWidgets()}>Retry</button></p>
        ) : widgets.length === 0 ? (
          <div className="flex h-full items-center justify-center text-(--text-secondary)">
            <p>Start to create your dashboard !</p>
          </div>
        ) : (
          <>
            {error && <p role="alert">{error}</p>}
            <DashboardGrid layouts={layouts} onLayoutChange={setLayouts} onDragStop={moveWidget}>
              {widgets.map((widget) => (
                <div key={widget.id}>
                  <Card
                    title={widget.type}
                    selected={selectedId === widget.id}
                    onSelect={() => setSelectedId((current) => (current === widget.id ? null : widget.id))}
                    onRemove={() => void removeWidget(widget.id)}
                  >
                    {widget.type === 'sncf.departures'
                      ? JSON.stringify({ station: widget.config.station, refresh_rate_seconds: widget.refresh_rate_seconds })
                      : JSON.stringify(widget.config)}
                  </Card>
                </div>
              ))}
            </DashboardGrid>
          </>
        )}
      </main>
    </div>
  );
}

export default DashboardPage;
