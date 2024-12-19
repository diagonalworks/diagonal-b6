import { HTMLAttributes, PropsWithChildren } from "react";
import { twMerge } from "tailwind-merge";

function Root({
	children,
	...props
}: PropsWithChildren & HTMLAttributes<HTMLDivElement>) {
	return (
		<div
			{...props}
			className={twMerge(
				"flex flex-col gap-1 p-2 absolute top-0 left-0",
				props.className,
			)}
		>
			{children}
		</div>
	);
}

const Button = ({
	children,
	...props
}: PropsWithChildren & HTMLAttributes<HTMLButtonElement>) => {
	return (
		<button
			{...props}
			className={twMerge(
				"bg-white flex [&>svg]:h-2.5 [&>svg]:w-2.5 justify-center hover:bg-graphite-10 text-sm border border-graphite-80 p-1",
				props.className,
			)}
		>
			{children}
		</button>
	);
};

export const MapControls = Object.assign(Root, {
	Button,
});
