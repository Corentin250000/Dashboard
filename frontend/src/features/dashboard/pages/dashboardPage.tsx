import { useState, type ComponentProps } from 'react';
import { Responsive, WidthProvider } from 'react-grid-layout/legacy';
import 'react-grid-layout/css/styles.css';
import { SIZES } from '../tools/Size';
import { useElementHeight } from '../tools/useElementsHeigh';
import TopBar from '../components/TopBar';

const ResponsiveGrid = WidthProvider(Responsive);

type Layouts = NonNullable<ComponentProps<typeof ResponsiveGrid>['layouts']>;

const COLS = { lg: 6, md: 4, sm: 2 };
const BREAKPOINTS = { lg: 1024, md: 640, sm: 0 };
const ROWS = 6;
const MARGIN = 16;
const MIN_ROW_HEIGHT = 80;
const MOBILE_ROW_HEIGHT = 140;

const initialLayouts: Layouts = {
  lg: [
    { i: 'a', x: 0, y: 0, ...SIZES.small },
    { i: 'b', x: 1, y: 0, ...SIZES.wide },
    { i: 'c', x: 3, y: 0, ...SIZES.tall },
    { i: 'd', x: 4, y: 0, ...SIZES.big },
  ],
  md: [
    { i: 'a', x: 0, y: 0, ...SIZES.small },
    { i: 'b', x: 1, y: 0, ...SIZES.wide },
    { i: 'c', x: 3, y: 0, ...SIZES.tall },
    { i: 'd', x: 0, y: 2, ...SIZES.big },
  ],
  sm: [
    { i: 'a', x: 0, y: 0, ...SIZES.small },
    { i: 'b', x: 0, y: 1, ...SIZES.wide },
    { i: 'c', x: 0, y: 2, ...SIZES.tall },
    { i: 'd', x: 0, y: 4, ...SIZES.big },
  ],
};

const ids = ['a', 'b', 'c', 'd'];

function DashboardPage() {
  const [layouts, setLayouts] = useState<Layouts>(initialLayouts);
  const [breakpoint, setBreakpoint] = useState('lg');
  const { ref, height } = useElementHeight<HTMLDivElement>();

  // PC : 6 rangées qui remplissent la hauteur. Autres écrans : hauteur fixe + scroll.
  const rowHeight =
    breakpoint === 'lg'
      ? Math.max(MIN_ROW_HEIGHT, (height - MARGIN * (ROWS + 1)) / ROWS)
      : MOBILE_ROW_HEIGHT;

  return (
    <>
    <div>
      <TopBar/>
    </div>
    <div
      ref={ref}
      style={{ height: '100vh', overflowX: 'hidden', overflowY: 'auto' }}
    >
      <ResponsiveGrid
        layouts={layouts}
        onLayoutChange={(_current, all) => setLayouts(all)}
        onBreakpointChange={(bp) => setBreakpoint(bp)}
        breakpoints={BREAKPOINTS}
        cols={COLS}
        rowHeight={rowHeight}
        margin={[MARGIN, MARGIN]}
        containerPadding={[MARGIN, MARGIN]}
        maxRows={breakpoint === 'lg' ? ROWS : Infinity}
        preventCollision
        compactType={null}
        isResizable={false}
      >
        {ids.map((id) => (
          <div key={id} style={{ background: 'steelblue', color: 'white' }}>
            {id}
          </div>
        ))}
      </ResponsiveGrid>
    </div>
    </>
  );
}

export default DashboardPage;