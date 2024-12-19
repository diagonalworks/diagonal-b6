import { DotIcon, FrameIcon, SquareIcon } from "@radix-ui/react-icons";
import { match } from "ts-pattern";

import * as circleIcons from "@/assets/icons/circle";
import { LabelledIcon } from "@/components/system/LabelledIcon";
import { LabelledIconProto } from "@/types/generated/ui";

export const LabelledIconAdapter = ({
	labelledIcon,
}: {
	labelledIcon: LabelledIconProto;
}) => {
	const icon = match(labelledIcon.icon)
		.with("area", () => <FrameIcon />)
		.with("point", () => <DotIcon />)
		.otherwise(() => {
			if (!labelledIcon.icon) return <SquareIcon />;
			const iconComponentName = `${labelledIcon.icon
				.charAt(0)
				.toUpperCase()}${labelledIcon.icon.slice(1)}`;

			if (circleIcons[iconComponentName as keyof typeof circleIcons]) {
				const Icon = circleIcons[iconComponentName as keyof typeof circleIcons];
				return <Icon />;
			}
			return <SquareIcon />;
		});

	return (
		<LabelledIcon>
			<LabelledIcon.Icon className=" text-ultramarine-60">
				{icon}
			</LabelledIcon.Icon>
			{/* otherwise hard for elements to fit in line */}
			<LabelledIcon.Label className="text-sm">
				{labelledIcon.label}
			</LabelledIcon.Label>
		</LabelledIcon>
	);
};
