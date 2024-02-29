import { Group } from '@visx/group';
import {
    BandScaleConfig,
    StringLike,
    scaleBand,
    scaleLinear,
} from '@visx/scale';
import { Text } from '@visx/text';
import { useMemo } from 'react';

export function Histogram<T>({
    data,
    bucket,
    value,
    color,
    label,
    width = 200,
    barHeight = 20,
    bandProps = {
        paddingInner: 0.4,
        paddingOuter: 0,
    },
    margin = { top: 5, right: 0, bottom: 0, left: 20 },
}: {
    data: T[];
    bucket: (d: T) => string;
    value: (d: T) => number;
    color: (d: T) => string;
    label?: (d: T) => string;
    width?: number;
    barHeight?: number;
    bandProps?: Omit<BandScaleConfig<StringLike>, 'type'>;
    margin?: { top: number; right: number; bottom: number; left: number };
}) {
    const boundedWidth = useMemo(
        () => width - margin.left - margin.right,
        [width, margin.left, margin.right]
    );

    const xScale = useMemo(() => {
        return scaleLinear({
            domain: [0, Math.max(...data.map(value))],
            range: [0, boundedWidth],
        });
    }, [data, boundedWidth, value]);

    const boundedHeight = useMemo(() => {
        const step = data.length / (1 - (bandProps.paddingInner ?? 0.2));
        return step * barHeight + (bandProps.paddingOuter ?? 0) * step * 2;
    }, [bandProps.paddingInner, bandProps.paddingOuter, barHeight, data]);

    const yScale = useMemo(() => {
        const buckets = data.map(bucket);

        return scaleBand({
            ...bandProps,
            domain: buckets,
            range: [0, boundedHeight],
        });
    }, [boundedHeight, bucket, data, bandProps]);

    return (
        <svg width={width} height={boundedHeight + margin.top + margin.bottom}>
            <Group top={margin.top}>
                {data.map((d) => {
                    console.log({
                        d,
                        bucket: bucket(d),
                        value: value(d),
                        x: xScale(value(d)),
                        y: yScale(bucket(d)),
                    });
                    return (
                        <Group top={yScale(bucket(d))}>
                            <rect
                                x={0}
                                y={yScale.bandwidth() - 3}
                                key={bucket(d)}
                                width={xScale(value(d))}
                                height={3}
                                fill={color(d)}
                                className=" stroke-graphite-80"
                                strokeWidth={0.7}
                                rx={2}
                            />
                            <rect
                                x={1} // to account for stroke width
                                width={yScale.bandwidth() - 4 - 2}
                                height={yScale.bandwidth() - 4 - 2}
                                fill={color(d)}
                                className=" stroke-graphite-80 "
                                strokeWidth={0.7}
                                rx={2}
                            />
                            <Text
                                x={yScale.bandwidth() - 4 - 2 + 6}
                                y={(yScale.bandwidth() - 4 - 2) / 2}
                                verticalAnchor="middle"
                                className="text-xs"
                            >
                                {label ? label(d) : bucket(d)}
                            </Text>
                        </Group>
                    );
                })}
            </Group>
        </svg>
    );
}
