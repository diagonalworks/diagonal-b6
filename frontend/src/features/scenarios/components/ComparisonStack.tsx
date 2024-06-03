import Outliner from '@/components/Outliner';
import { OutlinerSpec, useOutlinersStore } from '@/stores/outliners';

export default function ComparisonStack({
    id,
    origin,
}: {
    id: OutlinerSpec['id'];
    origin?: OutlinerSpec['id'];
}) {
    const outliner = useOutlinersStore((state) => state.outliners[id]);
    const originOutliner = useOutlinersStore((state) =>
        origin ? state.outliners[origin] : undefined
    );

    if (!outliner) return null;
    return <Outliner outliner={outliner} origin={originOutliner} />;
}
