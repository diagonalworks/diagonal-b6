import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
	docsSidebar: [
		"api",
		{
			type: "category",
			label: "Backend",
			link: { type: "doc", id: "backend/index" },
			items: ["backend/worlds", "backend/ingest"],
		},
		"frontend/index",
		"contributing/index",
		"quirks",
	],
};

export default sidebars;
