import { useOutlinerContext } from '@/lib/context/outliner';
import colors from '@/tokens/colors.json';
import { HistogramBarLineProto, SwatchLineProto } from '@/types/generated/ui';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { useCallback, useEffect, useMemo } from 'react';
import { match } from 'ts-pattern';
import { Histogram } from '../system/Histogram';

const colorInterpolator = interpolateRgbBasis([
    colors.green[20],
    colors.cyan[50],
    colors.violet[80],
]);

// default color range uses the colorInterpolator to define a 6 color range
const defaultColorRange = [
    '#fff',
    ...Array.from({ length: 4 }, (_, i) => colorInterpolator(i / 4)),
];

type HistogramData = {
    total: number;
    index: number;
    label: string;
    count: number;
    origin: number | null;
};

export const HistogramAdaptor = ({
    type,
    bars,
    swatches,
    origin,
}: {
    type: 'swatch' | 'histogram';
    bars?: HistogramBarLineProto[];
    swatches?: SwatchLineProto[];
    origin?: {
        bars?: HistogramBarLineProto[];
        swatches?: SwatchLineProto[];
    };
}) => {
    const { outliner, setHistogramColorScale, setHistogramBucket } =
        useOutlinerContext();
    const scale = outliner.histogram?.colorScale;

    const data = useMemo(() => {
        return match(type)
            .with(
                'histogram',
                () =>
                    bars?.flatMap((bar, i) => {
                        return {
                            total: bar.total ?? 0,
                            index: bar.index ?? 0,
                            label: bar.range?.value ?? '',
                            count: bar.value ?? 0,
                            origin: origin
                                ? origin?.bars?.[i]?.value ?? 0
                                : null,
                        };
                    }) ?? []
            )
            .with(
                'swatch',
                () =>
                    swatches?.flatMap((swatch) => {
                        return {
                            index: swatch.index ?? 0,
                            label: swatch.label?.value ?? '',
                            /* Swatches do not have a count. Should be null but setting it to 0 
                            for now to avoid type errors. */
                            count: 0,
                            total: 0,
                            origin: null,
                        };
                    }) ?? []
            )
            .exhaustive();
    }, [type, bars, swatches]);

    useEffect(() => {
        const scale = scaleOrdinal({
            domain: data.map((d) => `${d.index}`),
            range:
                data.length <= defaultColorRange.length
                    ? defaultColorRange
                    : [
                          '#fff',
                          ...data.map((_, i) =>
                              colorInterpolator(i / data.length)
                          ),
                      ],
        });
        setHistogramColorScale(scale);
    }, [data]);

    const handleSelect = useCallback((d: HistogramData | null) => {
        setHistogramBucket(d?.index.toString());
    }, []);

    const selected = useMemo(() => {
        const selected = outliner?.histogram?.selected;
        if (!selected) return null;
        return data.find((d) => d.index.toString() === selected);
    }, [outliner.histogram?.selected, data]);

    return (
        <Histogram
            type={type}
            data={data}
            label={(d) => d.label}
            bucket={(d) => d.index.toString()}
            value={(d) => d.count}
            origin={(d) => d.origin}
            total={(d) => d.total}
            color={(d) => (scale ? scale(`${d.index}`) : '#fff')}
            onSelect={handleSelect}
            selected={selected}
            selectable
        />
    );
};
