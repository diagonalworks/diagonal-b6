import { scaleLinear } from '@visx/scale';
import { Text } from '@visx/text';
import { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from './Line';

export function Histogram<T>({
    data,
    bucket,
    value,
    color,
    label,
    width = 200,
    margin = { top: 5, right: 0, bottom: 0, left: 20 },
    selected,
    onSelect,
    selectable = false,
}: {
    data: T[];
    bucket: (d: T) => string;
    value: (d: T) => number;
    color: (d: T) => string;
    label?: (d: T) => string;
    selectable?: boolean;
    selected?: T | null;
    onSelect?: (d: T | null) => void;
    width?: number;
    margin?: { top: number; right: number; bottom: number; left: number };
}) {
    const [internalSelected, setInternalSelected] = useState<T | null>(null);

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

    const selectedBucket = selectable ? selected ?? internalSelected : null;

    const handleClick = (
        _: React.MouseEvent<HTMLDivElement, MouseEvent>,
        d: T
    ) => {
        const onSelectHandler = onSelect ?? setInternalSelected;
        if (selectedBucket && bucket(d) === bucket(selectedBucket)) {
            onSelectHandler(null);
        } else {
            onSelectHandler(d);
        }
    };

    return (
        <div className="flex flex-col [&_.line]:border-t-0 first:[&_.line]:border-t">
            {data.map((d) => {
                const isSelected =
                    selectedBucket && bucket(d) === bucket(selectedBucket);
                return (
                    <Line
                        className={twMerge(selectable && 'cursor-pointer')}
                        onClick={(e) =>
                            selectable ? handleClick(e, d) : undefined
                        }
                    >
                        <div
                            className={twMerge(
                                'transition-opacity',
                                selectedBucket && !isSelected && 'opacity-50'
                            )}
                        >
                            <div className="flex gap-1 mb-1">
                                <div
                                    className="w-4 h-4 rounded border border-graphite-80"
                                    style={{
                                        backgroundColor: color(d),
                                    }}
                                />
                                <span className="text-xs text-graphite-100">
                                    {label ? label(d) : bucket(d)}
                                </span>
                            </div>
                            <svg
                                width={width}
                                height={4 + 2}
                                className=" overflow-visible"
                            >
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
                                <Text
                                    x={xScale(value(d)) + 5}
                                    y={2}
                                    className="  fill-graphite-50"
                                    verticalAnchor="middle"
                                    fontSize={10}
                                >
                                    {value(d)}
                                </Text>
                            </svg>
                        </div>
                    </Line>
                );
            })}
        </div>
    );
}
