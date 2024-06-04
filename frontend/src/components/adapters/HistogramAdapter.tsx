import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { useCallback, useEffect, useMemo } from 'react';
import { match } from 'ts-pattern';

import { Histogram } from '@/components/system/Histogram';
import { useStackContext } from '@/lib/context/stack';
import { useMapStore } from '@/stores/map';
import { useWorldStore } from '@/stores/worlds';
import colors from '@/tokens/colors.json';
import { HistogramBarLineProto, SwatchLineProto } from '@/types/generated/ui';

const colorInterpolator = interpolateRgbBasis([
    colors.graphite[40],
    colors.violet[60],
    colors.red[60],
    colors.red[70],
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
    chartLabel,
}: {
    type: 'swatch' | 'histogram';
    bars?: HistogramBarLineProto[];
    swatches?: SwatchLineProto[];
    origin?: {
        bars?: HistogramBarLineProto[];
        swatches?: SwatchLineProto[];
    };
    chartLabel?: string;
}) => {
    const { outliner, data: outlinerData } = useStackContext();
    const world = useWorldStore((state) =>
        outliner ? state.worlds[outliner.world] : state.worlds.baseline
    );
    const mapActions = useMapStore((state) => state.actions);
    const histogram = useMapStore((state) =>
        outliner ? state.layers.histogram[outliner.id] : null
    );
    const scale = histogram?.spec.colorScale;

    useEffect(() => {
        const histogramLayer = outlinerData?.proto.layers?.find(
            (l) => l.path === 'histogram'
        );

        if (!histogram && outliner && histogramLayer) {
            mapActions.setHistogramLayer(outliner.id, {
                world: outliner.world,
                spec: {
                    tiles: `/tiles/${histogramLayer.path}/{z}/{x}/{y}.mvt?q=${histogramLayer.q}&r=collection/${world.featureId.namespace}/${world.featureId.value}`,
                    show:
                        outliner.properties.active ||
                        outliner.properties.transient,
                    selected: undefined,
                },
            });
        }
    }, [outlinerData, outliner, mapActions]);

    useEffect(() => {
        return () => {
            if (outliner) {
                mapActions.removeHistogramLayer(outliner.id);
            }
        };
    }, []);

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
        if (!outliner) return;
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
        mapActions.setHistogramScale(outliner.id, scale);
    }, [data, outliner, mapActions]);

    const handleSelect = useCallback(
        (d: HistogramData | null) => {
            if (!outliner) return;
            mapActions.setHistogramBucket(outliner.id, d?.index.toString());
        },
        [mapActions, outliner]
    );

    const selected = useMemo(() => {
        const selected = histogram?.spec.selected;
        if (!selected) return null;
        return data.find((d) => d.index.toString() === selected);
    }, [histogram?.spec, data]);

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
            chartLabel={chartLabel}
            selectable
        />
    );
};
