import { useStackContext } from '@/lib/context/stack';
import colors from '@/tokens/colors.json';
import { SwatchLineProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { Histogram } from '../system/Histogram';

const colorInterpolator = interpolateRgbBasis([
    '#fff',
    colors.amber[20],
    colors.violet[80],
]);

export const HistogramAdaptor = ({
    swatches,
}: {
    swatches: SwatchLineProto[];
}) => {
    const { state } = useStackContext();

    // @ts-expect-error: mismatch between current collection and new typed BE.
    const matchedCondition = state.stack?.proto?.bucketed.find((b: $FixMe) => {
        console.log({ b });
        return b.condition.indices.every((index: number, i: number) => {
            const chip = state.choiceChips[index];
            console.log({ chip });
            if (!chip) return false;
            return chip.value === b.condition.values[i];
        });
    });

    const buckets = matchedCondition?.buckets ?? [];

    const data = swatches.flatMap((swatch) => {
        if (swatch.index === -1) return [];
        return {
            index: swatch.index,
            label: swatch.label?.value ?? '',
            count: buckets?.[swatch.index]?.ids
                ? // this is probably wrong, but just to get a value to show
                  buckets?.[swatch.index]?.ids.reduce(
                      (acc: number, curr: { ids: Array<$FixMe> }) =>
                          acc + curr.ids.length,
                      0
                  ) ?? 0
                : 0,
        };
    });

    const histogramColorScale = scaleOrdinal({
        domain: data.map((d) => `${d.index}`),
        range: data.map((_, i) => colorInterpolator(i / data.length)),
    });

    return (
        <Histogram
            data={data}
            label={(d) => d.label}
            bucket={(d) => d.index.toString()}
            value={(d) => d.count}
            color={(d) => histogramColorScale(`${d.index}`)}
        />
    );
};
