import { useComparatorContext } from '@/lib/context/comparator';
import { createTeleporter } from 'react-teleporter';
import { HeaderAdapter } from './adapters/HeaderAdapter';
import { Line } from './system/Line';

export const LeftComparatorTeleporter = createTeleporter();
export const RightComparatorTeleporter = createTeleporter();

export function Comparator() {
    const { analysis } = useComparatorContext();

    const analysisTitle = analysis?.proto.stack?.substacks?.[0]?.lines?.map(
        (l) => l.header
    )[0];
    return (
        <div>
            <div className="border-t border-x border-graphite-30">
                {analysisTitle && (
                    <Line className="border-b-0">
                        <HeaderAdapter header={analysisTitle} />
                    </Line>
                )}
            </div>
            <div className="flex flex-row ">
                <div className="flex-grow">
                    <LeftComparatorTeleporter.Target />
                </div>
                <div className="flex-grow">
                    <RightComparatorTeleporter.Target />
                </div>
            </div>
        </div>
    );
}
