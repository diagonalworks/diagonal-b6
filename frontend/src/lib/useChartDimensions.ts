import { useEffect, useRef, useState } from 'react';

export type ChartDimensions = {
    width?: number;
    height?: number;
    marginTop: number;
    marginRight: number;
    marginBottom: number;
    marginLeft: number;
    boundedWidth: number;
    boundedHeight: number;
};

const combineChartDimensions = (
    dimensions: Partial<ChartDimensions>
): ChartDimensions => {
    const parsedDimensions = {
        marginTop: 40,
        marginRight: 30,
        marginBottom: 40,
        marginLeft: 75,
        ...dimensions,
    };

    return {
        ...parsedDimensions,
        boundedHeight: parsedDimensions?.height
            ? Math.max(
                  parsedDimensions.height -
                      parsedDimensions.marginTop -
                      parsedDimensions.marginBottom,
                  0
              )
            : 0,
        boundedWidth: parsedDimensions?.width
            ? Math.max(
                  parsedDimensions.width -
                      parsedDimensions.marginLeft -
                      parsedDimensions.marginRight,
                  0
              )
            : 0,
    };
};

export const useChartDimensions = (
    passedSettings: Partial<ChartDimensions>
): [
    React.RefObject<HTMLDivElement>,
    ChartDimensions & {
        width: number;
        height: number;
    }
] => {
    const ref = useRef<HTMLDivElement>(null);
    const dimensions = combineChartDimensions(passedSettings);

    const [width, setWidth] = useState(0);
    const [height, setHeight] = useState(0);

    useEffect(() => {
        if (dimensions.width && dimensions.height) return;

        const element = ref.current;
        if (!element) return;

        const resizeObserver = new ResizeObserver((entries) => {
            if (!Array.isArray(entries)) return;
            if (!entries.length) return;

            const entry = entries[0];

            if (width != entry.contentRect.width)
                setWidth(entry.contentRect.width);
            if (height != entry.contentRect.height)
                setHeight(entry.contentRect.height);
        });

        resizeObserver.observe(element);

        return () => resizeObserver.unobserve(element);
    }, [dimensions.width, dimensions.height, width, height]);

    const newSettings = combineChartDimensions({
        ...dimensions,
        width: dimensions.width ?? width,
        height: dimensions.height ?? height,
    }) as ChartDimensions & {
        // @TODO: fix this type casting, this should not be necessary.
        width: number;
        height: number;
    };

    return [ref, newSettings];
};
