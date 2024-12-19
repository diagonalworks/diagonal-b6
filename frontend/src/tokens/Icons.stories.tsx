import * as circleIcons from "@/assets/icons/circle";
import * as solidIcons from "@/assets/icons/solid";
import type { Meta } from "@storybook/react";
import { useMemo, useState } from "react";
import { twMerge } from "tailwind-merge";

const colors = [
	{ color: "violet", fill: "fill-violet-70", text: "text-violet-70" },
	{
		color: "ultramarine",
		fill: "fill-ultramarine-70",
		text: "text-ultramarine-70",
	},
	{ color: "graphite", fill: "fill-graphite-70", text: "text-graphite-70" },
	{ color: "rose", fill: "fill-rose-70", text: "text-rose-70" },
];

export const Icons = () => {
	const [color, setColor] = useState("violet");

	const selectedColor = useMemo(() => {
		return colors.find((c) => c.color === color);
	}, [color]);

	return (
		<div className="flex flex-col gap-8 ">
			<h1 className="pb-1 text-sm border-b text-violet-70 border-violet-70 ">
				Icons
			</h1>
			<div>
				<label className=" font-medium text-xs mr-2" htmlFor="color">
					Color
				</label>
				<select
					className={twMerge(
						"text-xs border border-graphite-30 hover:bg-graphite-10",
						selectedColor?.text,
					)}
					onChange={(e) => setColor(e.target.value)}
				>
					{colors.map((c) => (
						<option value={c.color}>{c.color}</option>
					))}
				</select>
			</div>
			<div className="flex flex-col gap-5">
				<h2>Solid Icons</h2>
				<div className="grid w-full grid-cols-6 gap-3">
					{Object.entries(solidIcons).map(([key, value]) => {
						const Icon = value;
						return (
							<div className="flex flex-row items-center gap-1">
								<Icon
									className={twMerge("fill-graphite-50", selectedColor?.fill)}
									width={20}
									height={20}
								/>
								<span className="text-xs text-graphite-50">{key}</span>
							</div>
						);
					})}
				</div>
			</div>
			<div className="flex flex-col gap-5">
				<h2>Circle Icons</h2>
				<div className="grid w-full grid-cols-6 gap-3">
					{Object.entries(circleIcons).map(([key, value]) => {
						const Icon = value;
						return (
							<div className="flex flex-row items-center gap-1">
								<Icon
									className={twMerge("fill-graphite-50", selectedColor?.fill)}
								/>
								<span className="text-xs text-graphite-50">{key}</span>
							</div>
						);
					})}
				</div>
			</div>
		</div>
	);
};

const meta = {
	title: "Tokens/Icons",
} as Meta;

export default meta;
