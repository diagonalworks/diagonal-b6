import type { Meta, StoryObj } from "@storybook/react";
import { useState } from "react";

import { Shop } from "@/assets/icons/circle";

import { Header as HeaderComponent } from "./Header";
import { LabelledIcon as LabelledIconComponent } from "./LabelledIcon";
import { Line as LineComponent } from "./Line";
import { Select as SelectComponent } from "./Select";

type Story = StoryObj<typeof LineComponent>;

export const LineValue: Story = {
	render: () => (
		<div className=" border border-graphite-30 border-dashed w-fit p-2">
			<LineComponent.Value>73</LineComponent.Value>
		</div>
	),
};

export const LabelledIcon: Story = {
	render: () => (
		<div className=" border border-graphite-30 border-dashed w-fit p-2">
			<LabelledIconComponent>
				<LabelledIconComponent.Icon>
					<Shop />
				</LabelledIconComponent.Icon>
				<LabelledIconComponent.Label>Collection</LabelledIconComponent.Label>
			</LabelledIconComponent>
		</div>
	),
};

const OPTIONS = ["5", "10", "15"];

const SelectStory = () => {
	const [value, setValue] = useState(OPTIONS[0]);

	const label = (value: string) => {
		return `${value} min`;
	};

	return (
		<div className=" border border-graphite-30 border-dashed w-fit p-2">
			<SelectComponent value={value} onValueChange={setValue}>
				<SelectComponent.Button>
					<SelectComponent.Primitive.Value>
						{label(value)}
					</SelectComponent.Primitive.Value>
				</SelectComponent.Button>
				<SelectComponent.Options>
					{OPTIONS.map((option) => (
						<SelectComponent.Option key={option} value={option}>
							{label(option)}
						</SelectComponent.Option>
					))}
				</SelectComponent.Options>
			</SelectComponent>
		</div>
	);
};

export const Select: Story = {
	render: () => <SelectStory />,
};

export const Header: Story = {
	render: () => (
		<div className=" border border-graphite-30 border-dashed w-fit p-2">
			<HeaderComponent>
				<HeaderComponent.Label>Header</HeaderComponent.Label>
				<HeaderComponent.Actions share target copy toggleVisible close />
			</HeaderComponent>
		</div>
	),
};

const meta: Meta = {
	title: "Primitives/Atoms",
};

export default meta;
