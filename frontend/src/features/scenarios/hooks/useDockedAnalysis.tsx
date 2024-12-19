import { useMemo } from "react";

import { useStartup } from "@/api/startup";
import { FeatureIDProto } from "@/types/generated/api";
import { HeaderLineProto, LineProto } from "@/types/generated/ui";
import { Docked } from "@/types/startup";

type Analysis = {
	id: FeatureIDProto;
	label?: HeaderLineProto;
	data: Docked;
};

/**
 * Get the docked analysis from the startup data.
 * @returns The docked analysis
 */
export default function useDockedAnalysis() {
	const startup = useStartup();
	const docked = startup.data?.docked;

	const analysis = useMemo(() => {
		return (
			docked?.flatMap((analysis: Docked) => {
				const label = analysis.proto.stack?.substacks?.[0].lines?.map(
					(l: LineProto) => l.header,
				)[0];

				return {
					data: analysis,
					id: analysis.proto.stack?.id,
					label,
				} as Analysis;
			}) ?? []
		);
	}, [docked]);

	return analysis;
}
