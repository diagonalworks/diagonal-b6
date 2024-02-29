import colors from '@/tokens/colors.json';
import type { Meta, StoryObj } from '@storybook/react';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { Histogram as HistogramComponent } from './Histogram';
type Story = StoryObj<typeof HistogramComponent>;

const dummyData: { bucket: string; value: number }[] = [
    { bucket: '0', value: 10 },
    { bucket: '1', value: 50 },
    { bucket: '2', value: 30 },
    { bucket: '3', value: 20 },
    { bucket: '4', value: 10 },
    { bucket: '5', value: 5 },
];

const colorInterpolator = interpolateRgbBasis([
    colors.amber[20],
    colors.violet[80],
]);

const histogramColorScale = scaleOrdinal({
    domain: dummyData.map((d) => d.bucket),
    range: dummyData.map((_, i) => colorInterpolator(i / dummyData.length)),
});

export const Histogram: Story = {
    render: () => {
        return (
            <HistogramComponent
                data={dummyData}
                bucket={(d) => d.bucket}
                value={(d) => d.value}
                color={(d) => histogramColorScale(d.bucket)}
                label={(d) =>
                    d.bucket === '0'
                        ? `${d.bucket} health services nearby`
                        : `${d.bucket} nearby`
                }
            />
        );
    },
};

const meta: Meta = {
    title: 'Components/Histogram',
};

export default meta;
