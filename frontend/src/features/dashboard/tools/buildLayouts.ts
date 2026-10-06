import type { WidgetInstance } from '../../shared/api';
import { SIZES } from './Size';
import type { Layouts } from '../components/DashboardGrid';

export function buildLayouts(widgets: WidgetInstance[]): Layouts {
  const lg = widgets.map((widget) => ({
    i: widget.id,
    x: widget.position.x,
    y: widget.position.y,
    ...SIZES[widget.size],
  }));

  const ordered = [...widgets].sort(
    (a, b) => a.position.y - b.position.y || a.position.x - b.position.x,
  );

  const stack = (columns: number) => {
    let y = 0;
    return ordered.map((widget) => {
      const size = SIZES[widget.size];
      const item = { i: widget.id, x: 0, y, w: Math.min(size.w, columns), h: size.h };
      y += size.h;
      return item;
    });
  };

  return { lg, md: stack(4), sm: stack(2) };
}
