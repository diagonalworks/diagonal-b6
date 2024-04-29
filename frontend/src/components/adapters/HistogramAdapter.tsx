import { appAtom } from '@/atoms/app';
import { useStackContext } from '@/lib/context/stack';
import colors from '@/tokens/colors.json';
import { HistogramBarLineProto, SwatchLineProto } from '@/types/generated/ui';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { useAtom } from 'jotai';
import { isNil } from 'lodash';
import { useCallback, useEffect, useMemo } from 'react';
import { match } from 'ts-pattern';
import { Histogram } from '../system/Histogram';

const colorInterpolator = interpolateRgbBasis([
    '#fff',
    colors.amber[20],
    colors.violet[80],
]);

type HistogramData = {
    index: number;
    label: string;
    count: number;
};

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
    const stackId = stack.state.stack?.id;
    const stackApp = stackId ? app.stacks[stackId] : null;

    const data = useMemo(() => {
        return match(type)
            .with(
                'histogram',
                () =>
                    bars?.flatMap((bar) => {
                        return {
                            index: bar.index ?? 0,
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

    const handleSelect = useCallback(
        (d: HistogramData | null) => {
            if (isNil(stackId)) return;
            setApp((draft) => {
                const histogram = draft.stacks[stackId].histogram;
                if (histogram) {
                    histogram.selected = d?.index ?? null;
                } else {
                    draft.stacks[stackId].histogram = {
                        colorScale: histogramColorScale,
                        selected: d?.index ?? null,
                    };
                }
            });
        },
        [stackId]
    );

    const selected = useMemo(() => {
        if (isNil(stackId)) return null;
        return data.find((d) => d.index === stackApp?.histogram?.selected);
    }, [stackId, stackApp?.histogram?.selected]);

    useEffect(() => {
        if (isNil(stackId)) return;
        setApp((draft) => {
            const histogram = draft.stacks[stackId].histogram;
            if (histogram) {
                histogram.colorScale = histogramColorScale;
            } else {
                draft.stacks[stackId].histogram = {
                    colorScale: histogramColorScale,
                    selected: null,
                };
            }
        });
    }, [stackId, histogramColorScale]);

    return (
        <Histogram
            type={type}
            data={data}
            label={(d) => d.label}
            bucket={(d) => d.index.toString()}
            value={(d) => d.count}
            color={(d) => histogramColorScale(`${d.index}`)}
            onSelect={handleSelect}
            selected={selected}
            selectable
        />
    );
};
