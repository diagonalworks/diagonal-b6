import type { Meta, StoryObj } from "@storybook/react";
import { scaleOrdinal } from "@visx/scale";
import { interpolateRgbBasis } from "d3-interpolate";

import colors from "@/tokens/colors.json";

import { Histogram as HistogramComponent } from "./Histogram";

type Story = StoryObj<typeof HistogramComponent>;

const dummyData: { bucket: string; value: number; total: number }[] = [
	{ bucket: "0", value: 10, total: 128 },
	{ bucket: "1", value: 50, total: 128 },
	{ bucket: "2", value: 30, total: 128 },
	{ bucket: "3", value: 20, total: 128 },
	{ bucket: "4", value: 10, total: 128 },
	{ bucket: "5", value: 8, total: 128 },
];

const colorInterpolator = interpolateRgbBasis([
	colors.amber[20],
	colors.violet[80],
]);

const histogramColorScale = scaleOrdinal({
	domain: dummyData.map((d) => d.bucket),
	range: dummyData.map((_, i) => colorInterpolator(i / dummyData.length)),
});

export const Default: Story = {
	render: () => {
		return (
			<HistogramComponent
				data={dummyData}
				bucket={(d) => d.bucket}
				value={(d) => d.value}
				total={(d) => d.total}
				color={(d) => histogramColorScale(d.bucket)}
				label={(d) =>
					d.bucket === "0"
						? `${d.bucket} health services nearby`
						: `${d.bucket} nearby`
				}
			/>
		);
	},
};

export const Selectable: Story = {
	render: () => {
		return (
			<HistogramComponent
				data={dummyData}
				bucket={(d) => d.bucket}
				value={(d) => d.value}
				total={(d) => d.total}
				color={(d) => histogramColorScale(d.bucket)}
				label={(d) =>
					d.bucket === "0"
						? `${d.bucket} health services nearby`
						: `${d.bucket} nearby`
				}
				selectable
			/>
		);
	},
};

const meta: Meta = {
	title: "Components/Histogram",
};

export default meta;
