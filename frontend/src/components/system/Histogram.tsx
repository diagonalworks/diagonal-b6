import { useChartDimensions } from '@/lib/useChartDimensions';
import { scaleLinear } from '@visx/scale';
import { Text } from '@visx/text';
import { ScaleLinear } from 'd3-scale';
import { motion } from 'framer-motion';
import { isNull } from 'lodash';
import React, { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from './Line';

const BAR_MARGIN = {
    marginTop: 1, // 1px to make space for the border
    marginRight: 64,
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
    selected,
    onSelect,
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
    selectable?: boolean;
    /** Optional controlled state for the value of the selected bucket. */
    selected?: T | null;
    /** Optional change handler for the value of the selected bucket. */
    onSelect?: (d: T | null) => void;
    /** The type of histogram to display. */
    type?: 'swatch' | 'histogram';
}) {
    const [internalSelected, setInternalSelected] = useState<T | null>(null);
    const [ref, dimensions] = useChartDimensions({
        ...BAR_MARGIN,
        height: BAR_HEIGHT + BAR_MARGIN.marginTop + BAR_MARGIN.marginBottom,
    });

    const xScale = useMemo(() => {
        const domain = data
            .filter((d) => !isNull(value(d)))
            .map(value) as number[];
        return scaleLinear({
            domain: [0, Math.max(...domain)],
            range: [0, dimensions.boundedWidth],
        });
    }, [data, dimensions.boundedWidth, value]);

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

    const labelRef = React.useRef<SVGSVGElement>(null);

    const Wrapper = selectable ? Line.Button : React.Fragment;

    const isSelected = selectedBucket && bucket(d) === bucket(selectedBucket);

    const lineValue = value(d);
    const barLength = lineValue ? xScale(lineValue) : 0;
    const originValue = origin ? origin(d) : null;
    const originBarLength = originValue ? xScale(originValue) : 0;
    const isDecreasing = originValue && lineValue && lineValue < originValue;

    const textX = isDecreasing ? originBarLength + barLength : barLength;

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
                    {/* current hack to not show 0 bucket */}
                    {lineValue && lineValue > 0 && type === 'histogram' ? (
                        <svg
                            width={dimensions.width}
                            height={dimensions.height}
                            className=" overflow-visible"
                        >
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
                                className="stroke-graphite-80 fill-graphite-80"
                                strokeWidth={0.7}
                            />
                            {originValue ? (
                                <motion.rect
                                    animate={{
                                        width: isDecreasing
                                            ? originBarLength
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
                                            ? 'fill-rose-40 stroke-rose-40'
                                            : 'fill-rose-60 stroke-rose-60'
                                    )}
                                    strokeWidth={0.7}
                                />
                            ) : null}
                            <motion.g
                                animate={{
                                    translateX: textX + 5,
                                    translateY: BAR_HEIGHT / 2,
                                }}
                            >
                                <Text
                                    innerRef={labelRef}
                                    className="  fill-graphite-50"
                                    verticalAnchor="middle"
                                    fontSize={10}
                                >
                                    {lineValue}
                                </Text>
                                {originValue && (
                                    <Text
                                        className={twMerge(
                                            isDecreasing
                                                ? 'fill-rose-40'
                                                : 'fill-rose-60'
                                        )}
                                        verticalAnchor="middle"
                                        fontSize={10}
                                        dx={
                                            (labelRef.current?.getBBox()
                                                .width ?? 0) + 2
                                        }
                                    >
                                        {`(${isDecreasing ? '' : '+'}${
                                            lineValue - originValue
                                        })`}
                                    </Text>
                                )}
                            </motion.g>
                        </svg>
                    ) : (
                        <></>
                    )}
                </div>
            </Wrapper>
        </Line>
    );
}
