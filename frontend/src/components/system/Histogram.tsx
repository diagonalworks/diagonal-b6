import { useChartDimensions } from '@/lib/useChartDimensions';
import { scaleLinear } from '@visx/scale';
import { ScaleLinear } from 'd3-scale';
import { motion } from 'framer-motion';
import { isNull } from 'lodash';
import React, { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from './Line';

const BAR_MARGIN = {
    marginTop: 1, // 1px to make space for the border
    marginRight: 20,
    marginBottom: 1,
    marginLeft: 1,
};
const BAR_HEIGHT = 4;

/**
 * Histogram component.
 * This component is used to display a histogram. It can be used with or without selection. The
 * selection can be controlled by the parent component or it can be controlled by the component
 * itself.
 */
export function Histogram<T>({
    data,
    bucket,
    value,
    color,
    label,
    origin,
    total,
    selected,
    onSelect,
    chartLabel,
    selectable = false,
    type = 'histogram',
}: {
    /** The data to display. */
    data: T[];
    /** The accessor function for the bucket. */
    bucket: (d: T) => string;
    /** The accessor function for the value. */
    value: (d: T) => number | null;
    /** The accessor function for the color. */
    color: (d: T) => string;
    /** The accessor function for the label. */
    label?: (d: T) => string;
    /** The accessor function for the origin. */
    origin?: (d: T) => number | null;
    total: (d: T) => number | null;
    selectable?: boolean;
    /** Optional controlled state for the value of the selected bucket. */
    selected?: T | null;
    /** Optional change handler for the value of the selected bucket. */
    onSelect?: (d: T | null) => void;
    /** The type of histogram to display. */
    chartLabel?: string;
    type?: 'swatch' | 'histogram';
}) {
    const [internalSelected, setInternalSelected] = useState<T | null>(null);
    const [ref, dimensions] = useChartDimensions({
        ...BAR_MARGIN,
        height: BAR_HEIGHT + BAR_MARGIN.marginTop + BAR_MARGIN.marginBottom,
    });

    const maxValue = useMemo(() => {
        if (total) {
            return Math.max(...data.flatMap((d) => total(d) ?? []));
        } else {
            return Math.max(...data.flatMap((d) => value(d) ?? []));
        }
    }, [data, total, value]);

    const xScale = useMemo(() => {
        return scaleLinear({
            domain: [0, maxValue],
            range: [0, dimensions.boundedWidth],
        });
    }, [dimensions.boundedWidth, maxValue]);

    const selectedBucket = selectable ? selected ?? internalSelected : null;

    const handleClick = (
        _: React.MouseEvent<HTMLButtonElement, MouseEvent>,
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
        <div
            className="flex flex-col [&_.line]:border-t-0  last:[&_.line]:border-b-0"
            ref={ref}
        >
            <Line className=" justify-between text-graphite-70 ">
                <div>{chartLabel}</div>
                <div>buildings</div>
            </Line>
            {data.map((d, i) => {
                return (
                    <HistogramBar
                        key={i}
                        {...{
                            d,
                            value,
                            color,
                            label,
                            bucket,
                            origin,
                            selectedBucket,
                            selectable,
                            handleClick,
                            xScale,
                            type,
                        }}
                    />
                );
            })}
        </div>
    );
}

function HistogramBar<T>({
    d,
    value,
    color,
    label,
    origin,
    bucket,
    selectable,
    handleClick,
    xScale,
    selectedBucket,
    type,
}: {
    d: T;
    value: (d: T) => number | null;
    color: (d: T) => string;
    label?: (d: T) => string;
    bucket: (d: T) => string;
    origin?: (d: T) => number | null;
    selectedBucket: T | null;
    selectable?: boolean;
    xScale: ScaleLinear<number, number, never>;
    handleClick: (
        e: React.MouseEvent<HTMLButtonElement, MouseEvent>,
        d: T
    ) => void;
    type: 'histogram' | 'swatch';
}) {
    const [ref, dimensions] = useChartDimensions({
        ...BAR_MARGIN,
        height: BAR_HEIGHT + BAR_MARGIN.marginTop + BAR_MARGIN.marginBottom,
    });

    const Wrapper = selectable ? Line.Button : React.Fragment;

    const isSelected = selectedBucket && bucket(d) === bucket(selectedBucket);

    const lineValue = value(d);
    const barLength = lineValue ? xScale(lineValue) : 0;
    const originValue = origin ? origin(d) : null;
    const originBarLength = originValue ? xScale(originValue) : 0;
    const isDecreasing = originValue && lineValue && lineValue < originValue;
    const diff = lineValue ? lineValue - (originValue ?? 0) : 0;

    return (
        <Line>
            <Wrapper
                {...(selectable && {
                    onClick: (e) => handleClick(e, d),
                })}
            >
                <div
                    ref={ref}
                    className={twMerge(
                        'transition-opacity w-full',
                        selectedBucket && !isSelected && 'opacity-50'
                    )}
                >
                    <div className="flex justify-between">
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
                        {lineValue !== 0 && (
                            <div className="text-xs text-graphite-70">
                                {diff !== 0 && (
                                    <span
                                        className={twMerge(
                                            'mr-1',
                                            isDecreasing
                                                ? 'text-ultramarine-50'
                                                : 'text-ultramarine-70'
                                        )}
                                    >
                                        {`${isDecreasing ? '' : '+'}${diff}`}
                                    </span>
                                )}
                                {lineValue !== 0 && <span>{lineValue}</span>}
                            </div>
                        )}
                    </div>
                    {/* current hack to not show 0 bucket */}
                    {lineValue && lineValue > 0 && type === 'histogram' ? (
                        <svg
                            width={dimensions.width}
                            height={dimensions.height}
                            className=" overflow-visible"
                        >
                            {
                                <motion.rect
                                    animate={{
                                        width: xScale.range()[1],
                                    }}
                                    x={dimensions.marginLeft}
                                    y={dimensions.marginTop}
                                    height={BAR_HEIGHT}
                                    rx={1}
                                    className={twMerge(
                                        'stroke-graphite-20 fill-graphite-20'
                                    )}
                                    strokeWidth={0.7}
                                />
                            }
                            {originValue !== 0 && (
                                <motion.rect
                                    animate={{
                                        width:
                                            isDecreasing || !originValue
                                                ? barLength
                                                : originBarLength,
                                    }}
                                    x={dimensions.marginLeft}
                                    y={dimensions.marginTop}
                                    height={BAR_HEIGHT}
                                    //fill={color(d)}
                                    rx={1}
                                    className={twMerge(
                                        'stroke-graphite-80 fill-graphite-80'
                                    )}
                                    strokeWidth={0.7}
                                />
                            )}
                            {!isNull(originValue) ? (
                                <motion.rect
                                    animate={{
                                        width: isDecreasing
                                            ? originBarLength - barLength
                                            : barLength - originBarLength,
                                    }}
                                    x={
                                        !isDecreasing
                                            ? originBarLength
                                            : barLength
                                    }
                                    y={dimensions.marginTop}
                                    height={BAR_HEIGHT}
                                    rx={1}
                                    className={twMerge(
                                        isDecreasing
                                            ? 'fill-ultramarine-40 stroke-ultramarine-40'
                                            : 'fill-ultramarine-60 stroke-ultramarine-60'
                                    )}
                                    strokeWidth={0.7}
                                />
                            ) : null}
                        </svg>
                    ) : (
                        <></>
                    )}
                </div>
            </Wrapper>
        </Line>
    );
}
