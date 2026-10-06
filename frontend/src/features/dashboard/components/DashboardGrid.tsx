import { useState, type ComponentProps, type ReactNode } from 'react';
import { Responsive, WidthProvider } from 'react-grid-layout/legacy';
import type { LayoutItem } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import { useElementHeight } from '../tools/useElementsHeigh';

const ResponsiveGrid = WidthProvider(Responsive);

export type Layouts = NonNullable<ComponentProps<typeof ResponsiveGrid>['layouts']>;

const COLS = { lg: 6, md: 4, sm: 2 };
const BREAKPOINTS = { lg: 1024, md: 640, sm: 0 };
const ROWS = 6;
const MARGIN = 16;
const MIN_ROW_HEIGHT = 80;
const MOBILE_ROW_HEIGHT = 140;

type Props = {
  layouts: Layouts;
  onLayoutChange: (all: Layouts) => void;
  onDragStop?: (id: string, x: number, y: number) => void;
  children: ReactNode;
};

export default function DashboardGrid({ layouts, onLayoutChange, onDragStop, children }: Props) {
  const [breakpoint, setBreakpoint] = useState('lg');
  const { ref, height } = useElementHeight<HTMLDivElement>();

  const rowHeight =
    breakpoint === 'lg'
      ? Math.max(MIN_ROW_HEIGHT, (height - MARGIN * (ROWS + 1)) / ROWS)
      : MOBILE_ROW_HEIGHT;

  return (
    <div ref={ref} className="h-full overflow-x-hidden overflow-y-auto">
      <ResponsiveGrid
        layouts={layouts}
        onLayoutChange={(_current, all) => onLayoutChange(all)}
        onBreakpointChange={(bp) => setBreakpoint(bp)}
        onDragStop={(_layout, _oldItem, newItem: LayoutItem | null) => {
          if (breakpoint === 'lg' && newItem) onDragStop?.(newItem.i, newItem.x, newItem.y);
        }}
        breakpoints={BREAKPOINTS}
        cols={COLS}
        rowHeight={rowHeight}
        margin={[MARGIN, MARGIN]}
        containerPadding={[MARGIN, MARGIN]}
        maxRows={breakpoint === 'lg' ? ROWS : Infinity}
        preventCollision
        compactType={null}
        isResizable={false}
        isDraggable={breakpoint === 'lg'}
        draggableHandle=".drag-handle"
        draggableCancel=".no-drag"
      >
        {children}
      </ResponsiveGrid>
    </div>
  );
}
