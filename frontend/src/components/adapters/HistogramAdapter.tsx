import { appAtom } from '@/atoms/app';
import { useStackContext } from '@/lib/context/stack';
import colors from '@/tokens/colors.json';
import { HistogramBarLineProto, SwatchLineProto } from '@/types/generated/ui';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { useAtom } from 'jotai';
import { useEffect, useMemo } from 'react';
import { match } from 'ts-pattern';
import { Histogram } from '../system/Histogram';

const colorInterpolator = interpolateRgbBasis([
    '#fff',
    colors.amber[20],
    colors.violet[80],
]);

export const HistogramAdaptor = ({
    type,
    bars,
    swatches,
}: {
    type: 'swatch' | 'histogram';
    bars?: HistogramBarLineProto[];
    swatches?: SwatchLineProto[];
}) => {
    const [app, setApp] = useAtom(appAtom);
    const stack = useStackContext();

    const data = useMemo(() => {
        return match(type)
            .with(
                'histogram',
                () =>
                    bars?.flatMap((bar) => {
                        return {
                            index: bar.index,
                            label: bar.range?.value ?? '',
                            count: bar.value,
                        };
                    }) ?? []
            )
            .with(
                'swatch',
                () =>
                    swatches?.flatMap((swatch) => {
                        return {
                            index: swatch.index,
                            label: swatch.label?.value ?? '',
                            /* Swatches do not have a count. Should be null but setting it to 0 
                            for now to avoid type errors. */
                            count: 0,
                        };
                    }) ?? []
            )
            .exhaustive();
    }, [type, bars, swatches]);

    const histogramColorScale = useMemo(() => {
        return scaleOrdinal({
            domain: data.map((d) => `${d.index}`),
            range: data.map((_, i) => colorInterpolator(i / data.length)),
        });
    }, [data]);

    useEffect(() => {
        setApp((draft) => {
            const id = stack.state.stack?.id;
            if (id) {
                draft.stacks[id].histogram = {
                    colorScale: histogramColorScale,
                };
            }
        });
    }, [histogramColorScale]);

    return (
        <Histogram
            type={type}
            data={data}
            label={(d) => d.label}
            bucket={(d) => d.index.toString()}
            value={(d) => d.count}
            color={(d) => histogramColorScale(`${d.index}`)}
        />
    );
};
