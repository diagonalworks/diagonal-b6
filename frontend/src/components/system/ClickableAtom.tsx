import React from "react";

import { AtomAdapter } from "@/components/adapters/AtomAdapter";
import { useStackContext } from "@/lib/context/stack";
import { NodeProto } from "@/types/generated/api";
import { AtomProto } from "@/types/generated/ui";

/**
 * Renders an atom that's clickable. Used, for example, in left right lines.
 */
export const ClickableAtom = ({
	atom,
	clickExpression,
	key_,
}: {
	atom: AtomProto;
	clickExpression: NodeProto | undefined;
	key_?: number;
}) => {
	const { evaluateNode } = useStackContext();

	const Wrapper = clickExpression ? "button" : React.Fragment;

	return (
		<Wrapper
			{...(clickExpression && {
				onClick: (e) => {
					e.preventDefault();
					e.stopPropagation();
					evaluateNode(clickExpression);
				},
			})}
		>
			<AtomAdapter key={key_} atom={atom} />
		</Wrapper>
	);
};
