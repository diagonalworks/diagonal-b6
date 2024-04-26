import { useChartDimensions } from '@/lib/useChartDimensions';
import { scaleLinear } from '@visx/scale';
import { Text } from '@visx/text';
import { motion } from 'framer-motion';
import { isNull } from 'lodash';
import React, { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from './Line';

const BAR_MARGIN = {
    marginTop: 1, // 1px to make space for the border
    marginRight: 32,
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

    const Wrapper = selectable ? Line.Button : React.Fragment;

    return (
        <div className="flex flex-col [&_.line]:border-t-0  last:[&_.line]:border-b-0">
            {data.map((d) => {
                const isSelected =
                    selectedBucket && bucket(d) === bucket(selectedBucket);

                const lineValue = value(d);

                return (
                    <Line>
                        <Wrapper
                            {...(selectable && {
                                onClick: (e) => handleClick(e, d),
                            })}
                            onClick={(e) =>
                                selectable ? handleClick(e, d) : undefined
                            }
                        >
                            <div
                                ref={ref}
                                className={twMerge(
                                    'transition-opacity w-full',
                                    selectedBucket &&
                                        !isSelected &&
                                        'opacity-50'
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
                                {lineValue &&
                                    lineValue > 0 &&
                                    type === 'histogram' && (
                                        <svg
                                            width={dimensions.width}
                                            height={dimensions.height}
                                            className=" overflow-visible"
                                        >
                                            <motion.rect
                                                animate={{
                                                    width: xScale(lineValue),
                                                }}
                                                x={dimensions.marginLeft}
                                                y={dimensions.marginTop}
                                                width={xScale(lineValue)}
                                                height={BAR_HEIGHT}
                                                fill={color(d)}
                                                rx={1}
                                                className="stroke-graphite-80"
                                                strokeWidth={0.7}
                                            />
                                            <motion.g
                                                animate={{
                                                    translateX:
                                                        xScale(lineValue) + 5,
                                                    translateY: BAR_HEIGHT / 2,
                                                }}
                                            >
                                                <Text
                                                    className="  fill-graphite-50"
                                                    verticalAnchor="middle"
                                                    fontSize={10}
                                                >
                                                    {lineValue}
                                                </Text>
                                            </motion.g>
                                        </svg>
                                    )}
                            </div>
                        </Wrapper>
                    </Line>
                );
            })}
        </div>
    );
}
