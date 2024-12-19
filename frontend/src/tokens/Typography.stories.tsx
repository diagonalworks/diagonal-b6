import type { Meta } from "@storybook/react";

export const Typography = () => {
	return (
		<div className="flex flex-col gap-8 ">
			<h1 className="pb-1 text-sm border-b text-violet-70 border-violet-70 ">
				Typography
			</h1>
			<div className="flex flex-col gap-4">
				<h2 className="text-lg text-graphite-70">Font Family</h2>
				<p className="text-4xl "> Unica 77</p>
			</div>
			<div className="flex flex-col gap-4 ">
				<h2 className="text-lg text-graphite-70">Body</h2>
				<div className="flex flex-col gap-2 ">
					<p className="base">Body base</p>
					<p className="font-medium base">Body medium</p>
				</div>
			</div>
			<div className="flex flex-col gap-4 ">
				<h2 className="text-lg text-graphite-70">Title</h2>
				<div className="flex flex-col gap-2 ">
					<p className="font-light title">Title base</p>
					<p className="title">Title medium</p>
					<p className="font-medium title">Title large</p>
				</div>
			</div>
		</div>
	);
};

const meta = {
	title: "Tokens/Typography",
} as Meta;

export default meta;
