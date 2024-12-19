import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo } from "react";

import { useOutlinersStore } from "@/stores/outliners";
import { useViewStore } from "@/stores/view";
import { useWorkspaceStore } from "@/stores/workspace";
import { useWorldStore } from "@/stores/worlds";
import { StartupRequest, StartupResponse } from "@/types/startup";
import { getWorldFeatureId } from "@/utils/world";

import { b6 } from "./client";

const getStartup = (request: StartupRequest): Promise<StartupResponse> => {
	return b6.post("startup", null, {
		params: request,
	});
};

export const useStartup = () => {
	const root = useWorkspaceStore((state) => state.root);
	const worlds = useWorldStore((state) => state.worlds);
	const view = useViewStore((state) => state.initialView);
	const actions = useOutlinersStore((state) => state.actions);
	const { setFeatureId, createWorld } = useWorldStore((state) => state.actions);

	const request = useMemo(() => {
		return {
			...(root && { r: root || "" }),
			...(view.latitude &&
				view.longitude && { ll: `${view.latitude},${view.longitude}` }),
			...(view.zoom && { z: `${view.zoom}` }),
		};
	}, [root, view]);

	const query = useQuery({
		queryKey: ["startup", request.r, request.ll, request.z],
		queryFn: () => getStartup(request),
	});

	useEffect(() => {
		if (query.data) {
			const featureId = getWorldFeatureId({ ...query.data.root });
			if (!root) {
				createWorld({
					id: "baseline",
					featureId,
					tiles: "collection/diagonal.works/0",
				});
			} else {
				if (worlds[root]) {
					setFeatureId(root, featureId);
				} else {
					createWorld({
						id: root,
						featureId,
						tiles: "collection/diagonal.works/0",
					});
				}
			}

			query.data.docked?.forEach((d, i) => {
				actions.add({
					id: `docked-${i}`,
					world: root || "baseline",
					properties: {
						active: false,
						docked: true,
						transient: false,
						type: "core",
						show: false,
					},
					data: d,
				});
			});
		}
	}, [query.data, actions, root]);

	return query;
};
