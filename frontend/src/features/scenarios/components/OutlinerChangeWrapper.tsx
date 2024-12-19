import { MinusIcon, PlusIcon } from "@radix-ui/react-icons";
import { isEqual } from "lodash";
import { PropsWithChildren } from "react";
import { twMerge } from "tailwind-merge";

import { OutlinerSpec } from "@/stores/outliners";
import { StackProto } from "@/types/generated/ui";
import { findAtoms } from "@/utils/atoms";

import { Change, useChangesStore } from "../stores/changes";

/**
 * A wrapper to be used around outliner elements that allows the user to add or remove the element from a change.
 * @param id - The id of the change
 * @param outliner - The outliner specification
 * @param stack - The stack proto response for the outliner
 * @param children - The children to wrap
 * @returns The wrapped children
 */
export const OutlinerChangeWrapper = ({
	id,
	outliner,
	stack,
	children,
}: {
	id: Change["id"];
	outliner: OutlinerSpec;
	stack?: StackProto;
} & PropsWithChildren) => {
	const actions = useChangesStore((state) => state.actions);
	const change = useChangesStore((state) => state.changes[id]);

	const featureId = stack?.id;

	const isInChange = change.spec.features.find((f) => isEqual(f.id, featureId));

	const showChangeElements = featureId && !change.created;

	const labelledIcon = stack?.substacks?.[0]?.lines?.flatMap((l) =>
		findAtoms(l, "labelledIcon"),
	)?.[0]?.labelledIcon;

	if (!showChangeElements) return children;

	return (
		<>
			<div className="flex justify-start">
				<button
					onClick={() => {
						const expression = outliner.request?.expression ?? "";

						if (isInChange) {
							actions.removeFeature(id, {
								expression,
								id: featureId,
								label: labelledIcon,
							});
						}
						actions.addFeature(id, {
							expression,
							id: featureId,
							label: labelledIcon,
						});
					}}
					className="-mb-[2px] p-2 flex gap-1  items-center text-xs  text-rose-90 rounded-t border-b-0 bg-rose-40 hover:bg-rose-30 border border-rose-50"
				>
					{isInChange ? (
						<>
							<MinusIcon /> remove from change
						</>
					) : (
						<>
							<PlusIcon /> add to change
						</>
					)}
				</button>
			</div>
			<div
				className={twMerge(
					"stack-wrapper",
					showChangeElements && " border border-rose-50 rounded",
				)}
			>
				{children}
			</div>
		</>
	);
};
