import { BandScaleConfig, StringLike, scaleLinear } from '@visx/scale';
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

    return (
        <div className="flex flex-col gap-3">
            {data.map((d) => {
                return (
                    <div>
                        <div className="flex gap-1 mb-0.5">
                            <div
                                className="w-4 h-4 rounded border border-graphite-80"
                                style={{ backgroundColor: color(d) }}
                            />
                            <span className="text-xs text-graphite-100">
                                {label ? label(d) : bucket(d)}
                            </span>
                        </div>
                        <svg width={width} height={4 + 2}>
                            <rect
                                x={1}
                                y={1}
                                width={xScale(value(d))}
                                height={4}
                                fill={color(d)}
                                rx={1}
                                className="stroke-graphite-80"
                                strokeWidth={0.7}
                            />
                        </svg>
                    </div>
                );
            })}
        </div>
    );
}
