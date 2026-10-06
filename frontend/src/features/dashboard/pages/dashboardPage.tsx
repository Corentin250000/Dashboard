import { useState } from 'react';
import TopBar from '../components/TopBar';
import DashboardGrid, { type Layouts } from '../components/DashboardGrid';
import { SIZES } from '../tools/Size';

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

  return (
    <div className="flex h-dvh flex-col overflow-hidden">
      <TopBar />
      <main className="min-h-0 flex-1">
        <DashboardGrid layouts={layouts} onLayoutChange={setLayouts}>
          {ids.map((id) => (
            <div key={id} style={{ background: 'steelblue', color: 'white' }}>
              {id}
            </div>
          ))}
        </DashboardGrid>
      </main>
    </div>
  );
}

export default DashboardPage;