import type { WidgetInstance, WidgetSize } from '../../shared/api';
import { SIZES } from './Size';

const COLUMNS = 6;
const ROWS = 6;

export function findFreeSpot(
  widgets: WidgetInstance[],
  size: WidgetSize,
): { x: number; y: number } | null {
  const occupied = Array.from({ length: ROWS }, () => Array<boolean>(COLUMNS).fill(false));

  for (const widget of widgets) {
    const { w, h } = SIZES[widget.size];
    for (let y = widget.position.y; y < widget.position.y + h; y += 1) {
      for (let x = widget.position.x; x < widget.position.x + w; x += 1) {
        if (x >= 0 && x < COLUMNS && y >= 0 && y < ROWS) occupied[y][x] = true;
      }
    }
  }

  const { w, h } = SIZES[size];
  for (let y = 0; y <= ROWS - h; y += 1) {
    for (let x = 0; x <= COLUMNS - w; x += 1) {
      let free = true;
      for (let dy = 0; dy < h && free; dy += 1) {
        for (let dx = 0; dx < w; dx += 1) {
          if (occupied[y + dy][x + dx]) {
            free = false;
            break;
          }
        }
      }
      if (free) return { x, y };
    }
  }

  return null;
}
