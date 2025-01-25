import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const config: Config = {
	title: "diagonal b6",

	// Set the production url of your site here
	url: "https://diagonalworks.github.io/",
	// Set the /<baseUrl>/ pathname under which your site is served
	// For GitHub pages deployment, it is often '/<projectName>/'
	// baseUrl: "/dagonal-b6",
	baseUrl: "/",

	tagline: "Documentation for the diagonal.works geospatial analysis engine",

	// GitHub pages deployment config.
	// If you aren't using GitHub pages, you don't need these.
	organizationName: "diagonalworks",
	projectName: "diagonal-b6",
	onBrokenLinks: "throw",
	onBrokenMarkdownLinks: "warn",

	// Even if you don't use internationalization, you can use this field to set
	// useful metadata like html lang. For example, if your site is Chinese, you
	// may want to replace "en" with "zh-Hans".
	i18n: {
		defaultLocale: "en",
		locales: ["en"],
	},

	presets: [
		[
			"classic",
			{
				docs: {
					sidebarPath: "./sidebars.ts",
				},
				blog: {
					showReadingTime: true,
					feedOptions: {
						type: ["rss", "atom"],
						xslt: true,
					},
					onInlineTags: "warn",
					onInlineAuthors: "warn",
					onUntruncatedBlogPosts: "warn",
				},
				theme: {
					customCss: "./src/css/custom.css",
				},
			} satisfies Preset.Options,
		],
	],
	themeConfig: {
		navbar: {
			title: "b6",
			items: [
				{
					type: "docSidebar",
					sidebarId: "docsSidebar",
					position: "left",
					label: "Documentation",
				},
				{
					href: "https://github.com/diagonal/diagonal-b6",
					label: "GitHub",
					position: "right",
				},
			],
		},
		footer: {
			style: "light",
			links: [
				{
					title: "Docs",
					items: [
						{ label: "API documentation", to: "/docs/api" },
						{ label: "Backend", to: "/docs/backend" },
						{ label: "Frontend", to: "/docs/frontend" },
						{ label: "Contributing", to: "/docs/contributing" },
						{ label: "Quirks", to: "/docs/quirks" },
					],
				},
				{
					title: "More",
					items: [
						{
							label: "GitHub",
							href: "https://github.com/diagonalworks/diagonal-b6",
						},
					],
				},
			],
			copyright: `Copyright © ${new Date().getFullYear()} diagonal.works.`,
		},
		prism: {
			theme: prismThemes.github,
			darkTheme: prismThemes.dracula,
		},
	} satisfies Preset.ThemeConfig,
};

export default config;
